// Package websocket provides a websocket server(store) instance
// which used for receiving websocket messages. 
// Clients who want to communicate via websocket
// must register themselves to a "room"
package websocket

import (
	"encoding/json"
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/websocket"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

type Server struct {
	Store *websocket.Server
	ConnectionHeaders map[string]ConnectionHeaders
	Events map[string][]EventHandlerFunc
}

type (
	EventHandlerFunc func(headers ConnectionHeaders, message []byte)
)

var store *Server

// Return a websocket store
func Store() Server {
	if store != nil {
		return *store
	}
	websocketServer := websocket.New(websocket.Config{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	})

	store = &Server {
		Store: websocketServer,
		ConnectionHeaders: make(map[string]ConnectionHeaders),
		Events: make(map[string][]EventHandlerFunc),
	}

	websocketServer.OnConnection(store.HandleOnConnection)

	return *store
}

func (ws *Server) Handler() context.Handler {
	// the func implementation is copied from 
	// https://github.com/kataras/iris/blob/master/websocket/server.go#Handler

	return func(ctx context.Context) {
		s := ws.Store
		c := s.Upgrade(ctx)
		if c.Err() != nil {
			return
		}

		var connectionHeaders ConnectionHeaders = ConnectionHeaders {
			AppID: util.GetAppID(ctx),
			UserID: util.GetUserID(ctx),
			UserIP: util.GetUserIP(ctx),
		}

		ws.ConnectionHeaders[c.ID()] = connectionHeaders
		ws.HandleOnConnection(c)

		// start the ping and the messages reader
		c.Wait()
	}
}

func (ws *Server) On(event string, handler EventHandlerFunc) {
	if ws.Events[event] == nil {
		ws.Events[event] = make([]EventHandlerFunc, 0)
	}

	util.LogTrace.Printf("On Event: %s", event)
	ws.Events[event] = append(ws.Events[event], handler)
}

func (ws *Server) messageHandlerCurryFunc(connection websocket.Connection) websocket.NativeMessageFunc {
	return func(rowMsg []byte) {
		headers := ws.ConnectionHeaders[connection.ID()]

		var event Event = Event{}
		json.Unmarshal(rowMsg, &event)
		if event.Name != "" {
			callbacks, ok := ws.Events[event.Name]
			if ok {
				for i := range callbacks {
					callbacks[i](headers, rowMsg)
				}
			}
		}
	}
}

func (ws *Server) HandleOnConnection (connection websocket.Connection) {
	connection.OnMessage(ws.messageHandlerCurryFunc(connection))

	headers, ok := ws.ConnectionHeaders[connection.ID()]
	if !ok {
		// TODO: may handle error here
		return
	}
	connection.Join(headers.AppID)
}




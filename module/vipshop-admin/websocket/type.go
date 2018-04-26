package websocket

type ConnectionHeaders struct {
	AppID string
	UserID string
	UserIP string
}

type RealtimeEvent struct {
	Event string
	Handle EventHandlerFunc
}

type Event struct {
	Name string `json:"event"`
}

func EventEntrypoint(event string, handle EventHandlerFunc) RealtimeEvent {
	return RealtimeEvent {
		Event: event,
		Handle: handle,
	}
}

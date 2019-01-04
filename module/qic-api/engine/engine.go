package engine

import (
	"emotibot.com/emotigo/module/qic-api/util/general"
	_ "emotibot.com/emotigo/pkg/logger"
	"fmt"
	"strings"
)

var (
	ErrSyntaxError = fmt.Errorf("illegal engine syntax")
	ErrOutOfIndex  = fmt.Errorf("Index out of range")
)

type StateNode struct {
	Type           string
	Next           string
	Name           string
	Final          bool
	SentenceGroups map[string]bool
	Remain         int
}

type StateMachine interface {
	Digest(expression string) error
	IsAccept(s string) bool
}

type FlowStateMachine struct {
	Nodes map[string]StateNode
}

// IsAccept accept input string if it in the final state
// the input format should be like A,B|C|D...
// that A,B,C,D is in uuid V3 format
func (sm *FlowStateMachine) IsAccept(s string) bool {
	patterns := strings.Split(s, "|")
	node, ok := sm.Nodes["start"]
	if !ok {
		return false
	}

	for _, pattern := range patterns {
		uuids := strings.Split(pattern, ",")
		for _, uuid := range uuids {
			if fulfill, ok := node.SentenceGroups[uuid]; !fulfill && ok {
				node.Remain--
				if node.Remain == 0 {
					// should go to next state
					node = sm.Nodes[node.Next]
				}
			}
		}
	}
	return node.Final
}

// Digest takes a string as input and creates a state machine inside
// when IsAccept is called, it verifies the input string with the state machine
func (sm *FlowStateMachine) Digest(expression string) error {
	var err error
	tokens := strings.Split(expression, " ")
	sm.Nodes = map[string]StateNode{}

	idx := 0
	currentNodeName := "start"
	nextNodeName := general.RandStringRunes(8)

	for idx < len(tokens) {
		token := tokens[idx]
		if idx == 0 {
			token = strings.ToLower(token)
			if token != "if" && token != "must" {
				err = ErrSyntaxError
				return err
			}
		} else if token != "then" {
			err = ErrSyntaxError
			return err
		}

		// token is if must or then
		uuids, lastUUIDIdx, err := parseUUIDs(tokens, idx+1)
		if err != nil {
			return err
		}

		uuidsMap := map[string]bool{}
		for _, uuid := range uuids {
			uuidsMap[uuid] = false
		}

		node := StateNode{
			Name:           currentNodeName,
			Next:           nextNodeName,
			SentenceGroups: uuidsMap,
			Type:           "and",
			Remain:         len(uuidsMap),
		}

		if token == "if" {
			node.Final = true
		}
		sm.Nodes[node.Name] = node

		currentNodeName = nextNodeName
		nextNodeName = general.RandStringRunes(8)
		idx = lastUUIDIdx
	}

	node := StateNode{
		Name:           currentNodeName,
		Next:           "",
		SentenceGroups: map[string]bool{},
		Type:           "and",
		Final:          true,
	}
	sm.Nodes[node.Name] = node

	return err
}

func parseUUIDs(tokens []string, idx int) (uuids []string, lastIndex int, err error) {
	if idx >= len(tokens) {
		err = ErrOutOfIndex
		return
	}

	uuids = []string{}
	lastIndex = idx

	i := idx
	uuid := tokens[i]
	currentState := "parseID"
	for uuid != "then" && i < len(tokens) {
		if currentState == "parseID" {
			if uuid == "and" {
				err = ErrSyntaxError
				return
			}

			uuids = append(uuids, uuid)
			currentState = "parseAnd"
		} else {
			if uuid != "and" {
				err = ErrSyntaxError
				return
			}
			currentState = "parseID"
		}

		i += 1
		if i < len(tokens) {
			uuid = tokens[i]
		}
	}
	lastIndex = i
	return
}

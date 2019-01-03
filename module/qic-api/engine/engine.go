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
	SentenceGroups []string
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

func Parse(expression string) (map[string]StateNode, error) {
	var err error
	tokens := strings.Split(expression, " ")
	stateMachine := make(map[string]StateNode)

	idx := 0
	currentNodeName := "start"
	nextNodeName := general.RandStringRunes(8)

	for idx < len(tokens) {
		token := tokens[idx]
		if idx == 0 {
			token = strings.ToLower(token)
			if token != "if" && token != "must" {
				err = ErrSyntaxError
				return stateMachine, err
			}
		} else if token != "then" {
			err = ErrSyntaxError
			return stateMachine, err
		}

		// token is if must or then
		uuids, lastUUIDIdx, err := parseUUIDs(tokens, idx+1)
		if err != nil {
			return stateMachine, err
		}

		node := StateNode{
			Name:           currentNodeName,
			Next:           nextNodeName,
			SentenceGroups: uuids,
			Type:           "and",
		}

		if token == "if" {
			node.Final = true
		}
		stateMachine[node.Name] = node

		currentNodeName = nextNodeName
		nextNodeName = general.RandStringRunes(8)
		idx = lastUUIDIdx
	}

	node := StateNode{
		Name:           currentNodeName,
		Next:           "",
		SentenceGroups: []string{},
		Type:           "and",
		Final:          true,
	}

	stateMachine[node.Name] = node
	return stateMachine, err
}

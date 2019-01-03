package engine

import (
	"testing"
)

func TestParseUUIDs(t *testing.T) {
	tokens1 := []string{
		"if",
		"A",
		"and",
		"B",
		"and",
		"C",
	}

	uuids, lastIndex, err := parseUUIDs(tokens1, 1)
	if err != nil {
		t.Error(err)
		return
	}

	if lastIndex != len(tokens1) {
		t.Error("lastIndex error")
		return
	}

	if len(uuids) != 3 {
		t.Error("parse uuid failed")
		return
	}
}

func TestParseIf(t *testing.T) {
	expression := "if A then B then C"
	stateMachine, err := Parse(expression)
	if err != nil {
		t.Error(err)
		return
	}

	if len(stateMachine) != 4 {
		t.Errorf("parse expression(%s) failed", expression)
		return
	}

	node, ok := stateMachine["start"]
	if !ok {
		t.Error("does not have start node")
		return
	}

	if !node.Final {
		t.Errorf("parse expression(%s) failed", expression)
		return
	}
}

func TestParseMust(t *testing.T) {
	expression := "must a then b"
	stateMachine, err := Parse(expression)

	if err != nil {
		t.Error(err)
		return
	}

	if len(stateMachine) != 3 {
		t.Errorf("parse expression(%s) failed", expression)
		return
	}

	node, ok := stateMachine["start"]
	if !ok {
		t.Error("does not have start node")
		return
	}

	if node.Final {
		t.Errorf("parse expression(%s) failed", expression)
		return
	}
}

func TestParseSingleMust(t *testing.T) {
	expression := "must a"
	stateMachine, err := Parse(expression)

	if err != nil {
		t.Error(err)
		return
	}

	if len(stateMachine) != 2 {
		t.Errorf("parse expression(%s) failed", expression)
		return
	}

	node, ok := stateMachine["start"]
	if !ok {
		t.Error("does not have start node")
		return
	}

	if node.Final {
		t.Errorf("parse expression(%s) failed", expression)
		return
	}
}

func TestParseAnd(t *testing.T) {
	expression := "if a then b and c"
	stateMachine, err := Parse(expression)

	if err != nil {
		t.Error(err)
		return
	}

	if len(stateMachine) != 3 {
		t.Errorf("parse expression(%s) failed", expression)
		return
	}

	node, ok := stateMachine["start"]
	if !ok {
		t.Error("does not have start node")
		return
	}

	if !node.Final {
		t.Errorf("parse expression(%s) failed", expression)
		return
	}
}

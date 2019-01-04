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
	sm := &FlowStateMachine{}
	expression := "if A then B then C"
	err := sm.Digest(expression)
	if err != nil {
		t.Error(err)
		return
	}

	if len(sm.Nodes) != 4 {
		t.Errorf("parse expression(%s) failed", expression)
		return
	}

	node, ok := sm.Nodes["start"]
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

	sm := &FlowStateMachine{}
	err := sm.Digest(expression)

	if err != nil {
		t.Error(err)
		return
	}

	if len(sm.Nodes) != 3 {
		t.Errorf("parse expression(%s) failed", expression)
		return
	}

	node, ok := sm.Nodes["start"]
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
	sm := FlowStateMachine{}
	err := sm.Digest(expression)

	if err != nil {
		t.Error(err)
		return
	}

	if len(sm.Nodes) != 2 {
		t.Errorf("parse expression(%s) failed", expression)
		return
	}

	node, ok := sm.Nodes["start"]
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
	sm := &FlowStateMachine{}
	err := sm.Digest(expression)

	if err != nil {
		t.Error(err)
		return
	}

	if len(sm.Nodes) != 3 {
		t.Errorf("parse expression(%s) failed", expression)
		return
	}

	node, ok := sm.Nodes["start"]
	if !ok {
		t.Error("does not have start node")
		return
	}

	if !node.Final {
		t.Errorf("parse expression(%s) failed", expression)
		return
	}
}

func TestIsAccept(t *testing.T) {
	sm := &FlowStateMachine{}
	expression := "if A then B then C"
	err := sm.Digest(expression)
	if err != nil {
		t.Error(err)
		return
	}

	target := "A|B|C"
	accept := sm.IsAccept(target)

	if !accept {
		t.Error("accept input failed")
		return
	}

	target = "B|C"
	accept = sm.IsAccept(target)

	if !accept {
		t.Error("accept input failed")
		return
	}
}

func TestNotAccept(t *testing.T) {
	sm := &FlowStateMachine{}
	expression := "if A then B then C"
	err := sm.Digest(expression)
	if err != nil {
		t.Error(err)
		return
	}

	target := "A|B|D"
	accept := sm.IsAccept(target)
	if accept {
		t.Error("accept wrong input")
		return
	}

}

func TestAcceptMust(t *testing.T) {
	sm := &FlowStateMachine{}
	expression := "must A then B then C"
	err := sm.Digest(expression)
	if err != nil {
		t.Error(err)
		return
	}

	target := "A|B|C"
	accept := sm.IsAccept(target)
	if !accept {
		t.Error("accept input failed")
		return
	}

	target = "B|C"
	accept = sm.IsAccept(target)
	if accept {
		t.Error("accept wrong input")
		return
	}
}

func TestAccpetAnd(t *testing.T) {
	sm := &FlowStateMachine{}
	expression := "must A then B then C and D"
	err := sm.Digest(expression)
	if err != nil {
		t.Error(err)
		return
	}

	target := "A|B|C|B|D"
	accept := sm.IsAccept(target)
	if !accept {
		t.Error("accept input failed")
		return
	}

	target = "A|B|B|D|C"
	accept = sm.IsAccept(target)
	if !accept {
		t.Error("accept input failed")
		return
	}

	target = "B|C"
	accept = sm.IsAccept(target)
	if accept {
		t.Error("accept wrong input")
		return
	}
}

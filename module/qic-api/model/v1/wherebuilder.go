package model

import (
	"fmt"
	"strings"
)

type whereBuilder struct {
	ConcatLogic boolLogic
	data        []interface{}
	conditions  []string
}

type boolLogic string

const (
	andLogic boolLogic = "AND"
	orLogic  boolLogic = "OR"
)

func (w *whereBuilder) Parse() (string, []interface{}) {
	rawsql := strings.Join(w.conditions, fmt.Sprintf(" %s ", w.ConcatLogic))
	return rawsql, w.data
}

// In will create a condition that field should include inputs.
func (w *whereBuilder) In(fieldName string, inputs []interface{}) {
	if len(inputs) > 0 {
		in := fmt.Sprintf("`%s` IN (? %s)", fieldName, strings.Repeat(",?", len(inputs)-1))
		w.conditions = append(w.conditions, in)
		w.data = append(w.data, inputs...)
	}
}

// Eq will create a condition that field is equal to input.
func (w *whereBuilder) Eq(fieldName string, input interface{}) {
	w.conditions = append(w.conditions, fmt.Sprintf("`%s` = ?", fieldName))
	w.data = append(w.data, input)
}

func int64ToWildCard(inputs ...int64) []interface{} {
	var outputs = make([]interface{}, len(inputs))
	if len(outputs) == 0 {
		return nil
	}
	for i, val := range inputs {
		outputs[i] = val
	}
	return outputs
}

func int8ToWildCard(inputs ...int8) []interface{} {
	var outputs = make([]interface{}, len(inputs))
	if len(outputs) == 0 {
		return nil
	}
	for i, val := range inputs {
		outputs[i] = val
	}
	return outputs
}

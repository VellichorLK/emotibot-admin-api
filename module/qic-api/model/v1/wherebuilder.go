package model

import (
	"fmt"
	"strings"
)

type whereBuilder struct {
	ConcatLogic boolLogic
	alias       string
	data        []interface{}
	conditions  []string
}

func NewWhereBuilder(logic boolLogic, alias string) *whereBuilder {
	if alias != "" {
		alias = fmt.Sprintf("`%s`.", alias)
	}
	return &whereBuilder{
		ConcatLogic: logic,
		alias:       alias,
	}
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
		in := fmt.Sprintf("%s`%s` IN (? %s)", w.alias, fieldName, strings.Repeat(",?", len(inputs)-1))
		w.conditions = append(w.conditions, in)
		w.data = append(w.data, inputs...)
	}
}

// Eq will create a condition that field is equal to input.
func (w *whereBuilder) Eq(fieldName string, input interface{}) {
	w.conditions = append(w.conditions, fmt.Sprintf("%s`%s` = ?", w.alias, fieldName))
	w.data = append(w.data, input)
}

func (w *whereBuilder) Gte(fieldName string, input interface{}) {
	if input != nil {
		w.conditions = append(w.conditions, fmt.Sprintf("%s`%s` >= ?", w.alias, fieldName))
		w.data = append(w.data, input)
	}

}

func (w *whereBuilder) Lt(fieldName string, input interface{}) {
	if input != nil {
		w.conditions = append(w.conditions, fmt.Sprintf("%s`%s` < ?", w.alias, fieldName))
		w.data = append(w.data, input)
	}
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

func uint64ToWildCard(inputs ...uint64) []interface{} {
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

func stringToWildCard(inputs ...string) []interface{} {
	var outputs = make([]interface{}, len(inputs))
	if len(outputs) == 0 {
		return nil
	}
	for i, val := range inputs {
		outputs[i] = val
	}
	return outputs
}

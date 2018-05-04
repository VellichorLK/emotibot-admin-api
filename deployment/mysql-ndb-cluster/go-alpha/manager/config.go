package main

import (
	"fmt"
	"io"
	"text/template"
)

// Parser should be able to parse the template with the data
type Parser interface {
	Parse(string, io.Writer) error
}

// ClusterConfig contains the setting data of mysq-cluster.conf.
type ClusterConfig struct {
	ManageNodes []node
	NDBNodes    []node
	SQLNodes    []node
}

// node is for template's internal usage
type node struct {
	ID       int
	Hostname string
}

//MyConfig is the setting data of my.cnf
type MyConfig struct {
	ManageNode string
}

//Parse will parse the input template string with ClusterConfig's data into output writer.
func (c ClusterConfig) Parse(input string, output io.Writer) error {
	t, err := template.New("ClusterConfig").Parse(input)
	if err != nil {
		return fmt.Errorf("template parse failed, %v", err)
	}
	return t.Execute(output, c)
}

//Parse will parse the input template string with MyConfig's data into output writer.
func (m MyConfig) Parse(input string, output io.Writer) error {
	t, err := template.New("MySQLConfig").Parse(input)
	if err != nil {
		return fmt.Errorf("template parse failed, %v", err)
	}
	return t.Execute(output, m)
}

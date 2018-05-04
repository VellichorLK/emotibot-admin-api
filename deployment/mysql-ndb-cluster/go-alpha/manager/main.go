package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"gopkg.in/urfave/cli.v1"
)

//Action Should have
//	* Show: Show the info of current config or cluster.
//  * Init: Init the mysql cluster with the yaml file.
//  * Join: Join a node into current Cluster.
//  * Leave: delete a node from current Cluster.
var yamlFile string

func main() {
	log.SetFlags(log.Lshortfile)
	app := cli.NewApp()
	app.Name = "mysql cluster manager"
	app.Usage = "manage multiple hosts ndb cluster"
	app.Commands = []cli.Command{
		cli.Command{
			Name:  "show",
			Usage: "Show the info of current config or cluster",
			Subcommands: []cli.Command{
				cli.Command{
					Name:   "config",
					Usage:  "configuration",
					Action: ShowConfig,
				},
			},
		},
		cli.Command{
			Name:  "init",
			Usage: "initialize the Cluster",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "file, f",
					Value:       "ClusterFile",
					Usage:       "yaml file for initialize cluster",
					Destination: &yamlFile,
				},
			},
			Action: Init,
		},
	}
	// app.Action = ParseCluster
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

//ShowConfig is a CLI action for parse and print out the different config to Stdout.
func ShowConfig(c *cli.Context) error {
	f, err := ioutil.ReadFile("../template/mysql-cluster.cnf")
	if err != nil {
		return err
	}
	var parser Parser
	//TODO: Implement mysql config
	parser = ClusterConfig{
		ManageNodes: []node{
			node{ID: 1, Hostname: "192.168.0.0"},
		},
		NDBNodes: []node{
			node{ID: 2, Hostname: "192.168.0.2"},
			node{ID: 3, Hostname: "192.168.0.3"},
		},
		SQLNodes: []node{
			node{ID: 3, Hostname: "192.168.0.3"},
		},
	}
	err = parser.Parse(string(f), os.Stdout)
	if err != nil {
		return err
	}
	return nil
}

//Init is a cmd Action for initialize Cluster.
//Should return error if Cluster already initialize.
func Init(ctx *cli.Context) error {
	client, err := docker.NewEnvClient()
	if err != nil {
		return err
	}
	c := context.Background()
	lists, err := client.ContainerList(c, types.ContainerListOptions{})
	fmt.Printf("%+v\n", lists)
	customClient, err := docker.NewClient("tcp://172.16.101.47:2376", "", nil, nil)
	if err != nil {
		return err
	}
	lists, err = customClient.ContainerList(c, types.ContainerListOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("%+v/n", lists)
	return errors.New("Not implement yet")
}

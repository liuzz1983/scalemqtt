package main

import (
	"fmt"

	"github.com/liuzz1983/scalemqtt/mqtt"
	_ "github.com/pkg/profile"
)

func main() {
	//defer profile.Start(profile.CPUProfile).Stop()
	config, err := mqtt.LoadConfig("application.yml")
	if err != nil {
		fmt.Println("error in load config")
	}
	server, err := mqtt.NewServer(config)
	if err != nil {
		fmt.Println("error in build server")
	}
	err = server.Listen()
	if err != nil {
		fmt.Printf("error in build server %s", err)
	}
}

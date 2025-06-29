package main

import (
	"github.com/pilosus/goradieschen/protocol"
	"github.com/pilosus/goradieschen/server"
	"github.com/pilosus/goradieschen/store"
	"log"
)

func main() {
	log.Print("Server initializing...")

	s := store.NewStore()
	err := server.Start(":6380", func(command string) string {
		return protocol.ParseCommand(command, s)
	})
	if err != nil {
		log.Fatal(err)
	}
}

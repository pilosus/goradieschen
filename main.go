package main

import (
	"bufio"
	"context"
	"github.com/pilosus/goradieschen/protocol"
	"github.com/pilosus/goradieschen/server"
	"github.com/pilosus/goradieschen/store"
	"github.com/pilosus/goradieschen/ttlstore"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Print("Server initializing...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handleSignals(cancel)

	s := store.NewStore()

	ttl := ttlstore.NewTTLStore(
		ctx,
		func(key string) {
			// Add logging callback for key expiration
			log.Printf("Key expired: %s", key)
			// Remove key from the main key store
			s.Delete(key)
		})
	defer ttl.Stop()

	err := server.Start(ctx, ":6380", func(reader *bufio.Reader) string {
		return protocol.ParseCommand(reader, s, ttl)
	})
	if err != nil {
		log.Fatal(err)
	}
}

func handleSignals(cancel context.CancelFunc) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig
		log.Println("Shutdown signal received...")
		cancel()
	}()
}

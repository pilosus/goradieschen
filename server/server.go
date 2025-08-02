package server

import (
	"bufio"
	"context"
	"log"
	"net"
	"strings"
)

func Start(ctx context.Context, addr string, handler func(string) string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log.Printf("Server is listening on port: %s", addr)

	go func() {
		<-ctx.Done()
		log.Println("Server shutdown initiated")
		ln.Close() // this unblocks Accept()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil // graceful shutdown
			default:
				log.Println("Accept error:", err)
				continue
			}
		}
		go handleConnection(conn, handler)
	}
}

func handleConnection(conn net.Conn, handler func(string) string) {
	defer conn.Close()

	log.Printf("Client connected: %s", conn.RemoteAddr())
	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Connection closed: %s", err)
			return
		}
		response := handler(strings.TrimSpace(line)) + "\n"
		conn.Write([]byte(response))
	}
}

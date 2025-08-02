package server

import (
	"bufio"
	"context"
	"log"
	"net"
)

func Start(ctx context.Context, addr string, handler func(*bufio.Reader) string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log.Printf("Server is listening on port: %s", addr)

	go func() {
		<-ctx.Done()
		log.Println("Server shutdown initiated")
		if err := ln.Close(); err != nil {
			log.Printf("Error closing listener: %s", err)
		}
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

func handleConnection(conn net.Conn, handler func(*bufio.Reader) string) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing connection: %s", err)
		}
	}()

	log.Printf("Client connected: %s", conn.RemoteAddr())
	reader := bufio.NewReader(conn)

	for {
		response := handler(reader)
		if response == "" {
			log.Printf("Connection closed by handler")
			return
		}
		if _, err := conn.Write([]byte(response)); err != nil {
			log.Printf("Write error: %s", err)
			return
		}
	}
}

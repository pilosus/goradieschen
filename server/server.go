package server

import (
	"bufio"
	"log"
	"net"
	"strings"
)

func Start(addr string, handler func(string) string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	log.Printf("Server is listening on port: %s", addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Connection accept error: %s", err)
			continue
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

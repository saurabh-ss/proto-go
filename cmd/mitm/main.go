package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unicode"

	"github.com/saurabh/protohackers/internal/logger"
)

var port = flag.String("port", "50001", "Port to listen on")

var chatURL = flag.String("chat-url", "chat.protohackers.com", "URL of the chat server")
var chatPort = flag.Int("chat-port", 16963, "Port of the chat server")
var tonyAddress = flag.String("tony-address", "7YWHMfk9JZe0LM0g1ZauHuiSxhI", "Tony's boguscoin address")

func main() {
	flag.Parse()

	// Setup logging to logs directory
	logFile, err := logger.Setup("mitm")
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	ln, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	log.Println("Listening on port " + *port)

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down gracefully...")
		cancel()
		ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				log.Println("Server stopped")
				return
			default:
				log.Println("Accept error:", err)
				continue
			}
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	log.Println("New connection from", conn.RemoteAddr())
	defer conn.Close()

	upConn, err := net.Dial("tcp", net.JoinHostPort(*chatURL, fmt.Sprintf("%d", *chatPort)))
	if err != nil {
		log.Println("Failed to connect to chat server:", err)
		return
	}
	log.Println("Connected to chat server at", *chatURL, *chatPort)
	defer upConn.Close()

	done := make(chan struct{}, 2)

	// Forward messages from client to chat server
	go func() {
		defer func() {
			done <- struct{}{}
		}()
		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			log.Println("Received message from client:", strings.TrimSpace(line))
			line = handleMessage(line)
			upConn.Write([]byte(line))
		}
	}()

	// Forward messages from chat server to client
	go func() {
		defer func() {
			done <- struct{}{}
		}()
		upReader := bufio.NewReader(upConn)
		for {
			line, err := upReader.ReadString('\n')
			if err != nil {
				return
			}
			log.Println("Received message from chat server:", strings.TrimSpace(line))
			line = handleMessage(line)
			conn.Write([]byte(line))
		}
	}()

	<-done
	log.Println("Connection closed from", conn.RemoteAddr())
}

func handleMessage(message string) string {
	hasNewline := strings.HasSuffix(message, "\n")
	content := strings.TrimSuffix(message, "\n")

	fields := strings.Split(content, " ")
	modified := false
	for i, field := range fields {
		if isBoguscoinAddress(field) {
			log.Println("Replacing Boguscoin address:", field)
			fields[i] = *tonyAddress
			modified = true
		}
	}

	if !modified {
		return message
	}

	result := strings.Join(fields, " ")
	if hasNewline {
		result += "\n"
	}
	return result
}

func isBoguscoinAddress(field string) bool {
	n := len(field)
	if n < 26 || n > 35 {
		return false
	}
	if field[0] != '7' {
		return false
	}
	for _, r := range field {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

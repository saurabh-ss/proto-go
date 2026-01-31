package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/saurabh/protohackers/internal/logger"
)

var port = flag.String("port", "50001", "Port to listen on")

func main() {
	flag.Parse()

	// Setup logging to logs directory
	logFile, err := logger.Setup("voracious-code-storage")
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

	scanner := bufio.NewScanner(conn)
	writer := bufio.NewWriter(conn)

	// Helper function to write a line and flush
	writeLine := func(s string) error {
		if _, err := writer.WriteString(s + "\n"); err != nil {
			return err
		}
		return writer.Flush()
	}

	// Send initial READY message
	if err := writeLine("READY"); err != nil {
		log.Println("Write error:", err)
		return
	}

	// Read and process lines
	for scanner.Scan() {
		line := scanner.Text()
		log.Println("Received message:", line)

		// TODO: Add message processing logic here
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		if !errors.Is(err, io.EOF) {
			log.Println("Scanner error:", err)
			response, _ := json.Marshal(map[string]any{"status": "error", "error": "Scanner error: " + err.Error()})
			writeLine(string(response))
		}
	}

	log.Println("Connection closed from", conn.RemoteAddr())
}

package main

import (
	"bufio"
	"context"
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

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	cipher, err := reader.ReadBytes(0x00)
	if err != nil {
		log.Println("Read error:", err)
		return
	}

	if len(cipher) > 0 {
		cipher = cipher[:len(cipher)-1]
	}

	if len(cipher) == 0 {
		log.Println("No cipher received")
		return
	}

	for {
		data, err := reader.ReadBytes(0x0A)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println("Read error:", err)
			return
		}

		log.Println("Received data:", string(data))

		if len(data) > 0 {
			data = data[:len(data)-1]
		}

	}

	writer.Write(cipher)
	writer.Flush()
}

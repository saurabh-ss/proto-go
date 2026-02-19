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
	"syscall"

	"github.com/saurabh/protohackers/internal/logger"
)

var port = flag.String("port", "50001", "Port to listen on")

var authURL = flag.String("auth-url", "pestcontrol.protohackers.com", "URL of the authority server")
var authPort = flag.Int("auth-port", 20547, "Port of the authority server")

func main() {
	flag.Parse()

	// Setup logging to logs directory
	logFile, err := logger.Setup("insecure-sockets-layer")
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
	addr := conn.RemoteAddr().String()
	clog := log.New(log.Writer(), fmt.Sprintf("[%s] ", addr), log.Flags())

	clog.Println("New connection")
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	_ = reader
	_ = writer

}

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/saurabh/protohackers/internal/logger"
)

type request struct {
	Method *string  `json:"method"`
	Number *float64 `json:"number"`
}

type response struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

var port = flag.String("port", "50001", "Port to listen on")

func main() {
	flag.Parse()

	// Setup logging to logs directory
	logFile, err := logger.Setup("prime")
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
	defer log.Println("Connection closed from", conn.RemoteAddr())

	// Set connection timeout
	conn.SetDeadline(time.Now().Add(30 * time.Second))

	// Use buffered reader to handle line-delimited messages
	reader := bufio.NewReader(conn)
	encoder := json.NewEncoder(conn)

	for {
		// Reset deadline for each operation
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))

		// Read until newline (complete JSON message)
		line, err := reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				log.Println("Connection closed by client")
			} else {
				log.Println("Read error:", err)
			}
			return
		}

		// Trim the newline and any whitespace properly
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		log.Println("Received message:", line)

		var req request
		resp := response{Method: "isPrime", Prime: false}

		if err = json.Unmarshal([]byte(line), &req); err != nil {
			sendErrorAndClose(encoder, fmt.Errorf("unmarshal: %w", err))
			return
		}

		// Check if required fields are present
		if req.Method == nil || req.Number == nil {
			sendErrorAndClose(encoder, fmt.Errorf("missing required fields (method or number)"))
			return
		}

		// Check if method is valid
		if *req.Method != "isPrime" {
			sendErrorAndClose(encoder, fmt.Errorf("invalid method: %s", *req.Method))
			return
		}

		// Validate that number is a valid finite value
		if math.IsNaN(*req.Number) || math.IsInf(*req.Number, 0) {
			sendErrorAndClose(encoder, fmt.Errorf("number must be finite"))
			return
		}

		resp.Prime = isPrime(int64(*req.Number))
		if err = encoder.Encode(resp); err != nil {
			log.Println("Encode error:", err)
			return
		}
	}
}

// sendErrorAndClose sends an error response and logs the error
func sendErrorAndClose(encoder *json.Encoder, err error) {
	log.Println("Error:", err)
	resp := response{Method: "error", Prime: false}
	if encErr := encoder.Encode(resp); encErr != nil {
		log.Println("Failed to send error response:", encErr)
	}
}

func isPrime(n int64) bool {
	if n <= 1 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false // Even numbers aren't prime
	}
	// Only check odd divisors
	for i := int64(3); i*i <= n; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}

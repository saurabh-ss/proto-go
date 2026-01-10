package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net"

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

func main() {
	// Setup logging to logs directory
	logFile, err := logger.Setup("prime")
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	port := "50001"
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(err)
	}

	log.Println("Listening on port " + port)
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
	log.Println("New connection from", conn.RemoteAddr())
	defer conn.Close()
	defer log.Println("Connection closed from", conn.RemoteAddr())

	// Use buffered reader to handle line-delimited messages
	reader := bufio.NewReader(conn)

	for {
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

		// Trim the newline and any whitespace
		line = line[:len(line)-1] // Remove the '\n'

		if line == "" {
			continue
		}

		log.Println("Received message:", line)

		req := request{}
		response := response{Method: "isPrime", Prime: false}

		if err = json.Unmarshal([]byte(line), &req); err != nil {
			log.Println("Unmarshal error:", err)
			response.Method = "error"
			err2 := json.NewEncoder(conn).Encode(response)
			if err2 != nil {
				log.Println("Encode error:", err2)
			}
			return
		}

		// Check if required fields are present
		if req.Method == nil || req.Number == nil {
			log.Println("Missing required fields (method or number)")
			response.Method = "error"
			err2 := json.NewEncoder(conn).Encode(response)
			if err2 != nil {
				log.Println("Encode error:", err2)
			}
			return
		}

		// Check if method is valid
		if *req.Method != "isPrime" {
			log.Printf("Invalid method: %s", *req.Method)
			response.Method = "error"
			err2 := json.NewEncoder(conn).Encode(response)
			if err2 != nil {
				log.Println("Encode error:", err2)
			}
			return
		}

		response.Prime = isPrime(int64(*req.Number))
		if err = json.NewEncoder(conn).Encode(response); err != nil {
			log.Println("Encode error:", err)
			return
		}
	}
}

func isPrime(n int64) bool {
	if n <= 1 {
		return false
	}
	for i := int64(2); i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/saurabh/protohackers/internal/logger"
)

var port = flag.String("port", "50001", "Port to listen on")

func main() {
	flag.Parse()

	// Setup logging to logs directory
	logFile, err := logger.Setup("unusual-db")
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	conn, err := net.ListenPacket("udp", ":"+*port)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	log.Println("Listening for UDP packets on port " + *port)

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
		conn.Close()
	}()

	// All requests and responses must be shorter than 1000 bytes.
	buffer := make([]byte, 1024)

	db := make(map[string]string)
	db["version"] = "Saurabh's Key-Value Store 1.0"

	for {
		select {
		case <-ctx.Done():
			log.Println("Server stopped")
			return
		default:
			n, addr, err := conn.ReadFrom(buffer)
			if err != nil {
				select {
				case <-ctx.Done():
					log.Println("Server stopped")
					return
				default:
					log.Println("Read error:", err)
					continue
				}
			}
			// Handle each packet in a goroutine if you need concurrent processing
			strbuf := string(buffer[:n])
			parts := strings.SplitN(strbuf, "=", 2)
			if len(parts) == 2 {
				// SET request: key=value
				log.Println("SET request:", parts[0], "=", parts[1])
				if parts[0] == "version" {
					continue
				}
				db[parts[0]] = parts[1]
				// No response needed for SET requests in typical key-value UDP protocol
			} else if len(parts) == 1 {
				// GET request: key
				key := parts[0]
				log.Println("GET request:", key)
				value := db[key]
				// Response format: key=value
				response := key + "=" + value
				conn.WriteTo([]byte(response), addr)
			}
		}
	}

}

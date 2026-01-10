package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/saurabh/protohackers/internal/logger"
)

var port = flag.String("port", "50001", "Port to listen on")

func main() {
	flag.Parse()

	// Setup logging to logs directory
	logFile, err := logger.Setup("means")
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
	prices := make(map[int32]int32)

	mLen := 9
	buf := make([]byte, mLen)

	for {
		// Reset deadline for each operation
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))

		_, err := io.ReadFull(reader, buf)
		log.Println("Read", len(buf), "bytes", buf)
		if err != nil {
			if err == io.EOF {
				log.Println("Connection closed by client")
			} else {
				log.Println("Read error:", err)
			}
			return
		}

		if buf[0] != 'I' && buf[0] != 'Q' {
			log.Println("Invalid message type", buf[0])
			conn.Close()
			return
		}

		if buf[0] == 'I' {
			// Example [73 0 0 48 57 0 0 0 101]
			ts := int32(binary.BigEndian.Uint32(buf[1:5]))
			price := int32(binary.BigEndian.Uint32(buf[5:9]))
			_, exists := prices[ts]
			if exists {
				log.Println("Duplicate timestamp", ts)
				conn.Close()
				return
			}
			prices[ts] = price
			log.Println("Insert", ts, price)
		}

		if buf[0] == 'Q' {
			tsMin := int32(binary.BigEndian.Uint32(buf[1:5]))
			tsMax := int32(binary.BigEndian.Uint32(buf[5:9]))
			log.Println("Query", tsMin, tsMax)
			sum := 0
			count := 0
			for ts, price := range prices {
				if ts >= tsMin && ts <= tsMax {
					sum += int(price)
					count++
				}
			}
			var mean int32
			mean = 0
			if count > 0 {
				mean = int32(sum / count)
			}
			log.Println("Mean", mean, "from sum", sum, "count", count)
			response := make([]byte, 4)
			binary.BigEndian.PutUint32(response, uint32(mean))
			_, err := conn.Write(response)
			if err != nil {
				log.Println("Write error:", err)
				conn.Close()
				return
			}
		}
	}
}

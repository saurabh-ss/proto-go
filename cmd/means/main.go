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
	"sort"
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

type priceEntry struct {
	timestamp int32
	price     int32
}

func handleConnection(conn net.Conn) {
	log.Println("New connection from", conn.RemoteAddr())
	defer conn.Close()
	defer log.Println("Connection closed from", conn.RemoteAddr())

	reader := bufio.NewReader(conn)
	prices := make(map[int32]int32)
	sortedPrices := make([]priceEntry, 0)

	buf := make([]byte, 9)

	for {
		// Reset deadline for each operation
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))

		_, err := io.ReadFull(reader, buf)
		if err != nil {
			if err != io.EOF {
				log.Println("Read error:", err)
			}
			return
		}

		switch buf[0] {
		case 'I':
			ts := int32(binary.BigEndian.Uint32(buf[1:5]))
			price := int32(binary.BigEndian.Uint32(buf[5:9]))

			if _, exists := prices[ts]; exists {
				log.Println("Duplicate timestamp", ts)
				return
			}

			prices[ts] = price

			// Insert into sorted slice using binary search
			idx := sort.Search(len(sortedPrices), func(i int) bool {
				return sortedPrices[i].timestamp >= ts
			})

			// Insert at the correct position
			sortedPrices = append(sortedPrices, priceEntry{})
			copy(sortedPrices[idx+1:], sortedPrices[idx:])
			sortedPrices[idx] = priceEntry{timestamp: ts, price: price}

			log.Println("Insert", ts, price)

		case 'Q':
			tsMin := int32(binary.BigEndian.Uint32(buf[1:5]))
			tsMax := int32(binary.BigEndian.Uint32(buf[5:9]))
			log.Println("Query", tsMin, tsMax)

			// Binary search for range boundaries
			startIdx := sort.Search(len(sortedPrices), func(i int) bool {
				return sortedPrices[i].timestamp >= tsMin
			})

			endIdx := sort.Search(len(sortedPrices), func(i int) bool {
				return sortedPrices[i].timestamp > tsMax
			})

			var sum int64
			count := endIdx - startIdx

			for i := startIdx; i < endIdx; i++ {
				sum += int64(sortedPrices[i].price)
			}

			mean := int32(0)
			if count > 0 {
				mean = int32(sum / int64(count))
			}

			log.Println("Mean", mean, "from sum", sum, "count", count)

			response := make([]byte, 4)
			binary.BigEndian.PutUint32(response, uint32(mean))
			if _, err := conn.Write(response); err != nil {
				log.Println("Write error:", err)
				return
			}

		default:
			log.Println("Invalid message type", buf[0])
			return
		}
	}
}

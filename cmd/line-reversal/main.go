package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/saurabh/protohackers/internal/logger"
)

var port = flag.String("port", "50001", "Port to listen on")

type Session struct {
	Id  int32
	Pos int32
}

func main() {
	flag.Parse()

	// Setup logging to logs directory
	logFile, err := logger.Setup("line-reversal")
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

	sessions := make(map[int32]Session)

	buffer := make([]byte, 1024)
	for {
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
		strbuf := string(buffer[:n])
		log.Println("Received request:", strbuf)
		if strbuf[0] != '/' || strbuf[n-1] != '/' {
			log.Println("Invalid request not starting or ending with /:", strbuf)
			continue
		}

		parts := strings.Split(strbuf, "/")
		if len(parts) < 2 {
			log.Println("Invalid request not enough parts:", strbuf)
			continue
		}
		msgType := parts[1]

		switch msgType {
		case "connect":
			sid, err := strconv.Atoi(parts[2])
			if err != nil || sid < 0 {
				log.Println("Invalid sessionId:", parts[2], err)
				continue
			}
			s := int32(sid)
			_, exists := sessions[s]
			if !exists {
				sessions[s] = Session{Id: s, Pos: 0}
			}
			response := fmt.Sprintf("/ack/%d/%d/", s, sessions[s].Pos)
			conn.WriteTo([]byte(response), addr)

		case "data":
			sid, err := strconv.Atoi(parts[2])
			if err != nil || sid < 0 {
				log.Println("Invalid sessionId:", parts[2], err)
				continue
			}
			s := int32(sid)

			session, exists := sessions[s]
			if !exists {
				log.Println("Session not found:", s)
				response := fmt.Sprintf("/close/%d/", s)
				conn.WriteTo([]byte(response), addr)
				continue
			}
			// Example:/data/1234567/0/hello/
			pos, err := strconv.Atoi(parts[3])
			if err != nil || pos < 0 {
				log.Println("Invalid increment:", parts[3], err)
				continue
			}
			p := int32(pos)
			if p != session.Pos {
				response := fmt.Sprintf("/ack/%d/%d/", s, session.Pos)
				conn.WriteTo([]byte(response), addr)
				continue
			}

		case "close":
			log.Println("Close request:", msgType)
		default:
			log.Println("Invalid request:", msgType)
		}

	}
}

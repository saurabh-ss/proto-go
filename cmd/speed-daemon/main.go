package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/saurabh/protohackers/internal/logger"
)

var port = flag.String("port", "50001", "Port to listen on")

var clients = make(map[net.Conn]string)
var clientsMu sync.RWMutex

var heartbeats = make(map[net.Conn]time.Time)
var heartbeatsMu sync.RWMutex

type Camera struct {
	road  uint16
	mile  uint16
	limit uint16
}

type Dispatcher struct {
	numRoads uint8
	roads    []uint16
}

func handlePlate(conn net.Conn) {

}

func handleWantHeartbeat(conn net.Conn) {
	if _, exists := heartbeats[conn]; exists {
		sendError(conn, "already sent a heartbeat")
		return
	}

	heartbeatsMu.Lock()
	heartbeats[conn] = time.Now()
	heartbeatsMu.Unlock()

	data := make([]byte, 4)
	conn.Read(data)
	beat := binary.BigEndian.Uint32(data)
	ticker := time.NewTicker(time.Duration(beat) * time.Second / 10)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			err := sendHeartbeat(conn)
			if err != nil {
				log.Printf("failed to send heartbeat to %s: %v", conn.RemoteAddr(), err)
				return
			}
		}
	}()
}

func sendHeartbeat(conn net.Conn) error {
	_, err := conn.Write([]byte{0x41})
	if err != nil {
		log.Printf("failed to send heartbeat to %s: %v", conn.RemoteAddr(), err)
		return err
	}
	return nil
}

func handleIAmCamera(conn net.Conn) {
	if client, exists := clients[conn]; exists {
		sendError(conn, "already registered as a "+client)
	} else {
		clientsMu.Lock()
		clients[conn] = "camera"
		clientsMu.Unlock()
	}
}

func handleIAmDispatcher(conn net.Conn) {
	if client, exists := clients[conn]; exists {
		sendError(conn, "already registered as a "+client)
	} else {
		clientsMu.Lock()
		clients[conn] = "dispatcher"
		clientsMu.Unlock()
	}
}

func sendError(conn net.Conn, msg string) error {
	// 1. Safety check for the protocol limit (u8 length)
	if len(msg) > 255 {
		msg = msg[:255] // Or return an error
	}

	// 2. Pre-allocate exactly what we need
	msgBytes := []byte(msg)
	packet := make([]byte, 0, 2+len(msgBytes))

	// 3. Use append to make it more readable
	packet = append(packet, 0x10)                // Type Tag
	packet = append(packet, byte(len(msgBytes))) // String Length
	packet = append(packet, msgBytes...)         // String Content

	// 4. Handle the write error
	_, err := conn.Write(packet)
	if err != nil {
		log.Printf("failed to send error to %s: %v", conn.RemoteAddr(), err)
		return err
	}

	err = conn.Close()
	if err != nil {
		log.Printf("failed to close connection to %s: %v", conn.RemoteAddr(), err)
		return err
	}

	clientsMu.Lock()
	delete(clients, conn)
	clientsMu.Unlock()

	heartbeatsMu.Lock()
	delete(heartbeats, conn)
	heartbeatsMu.Unlock()

	return nil
}

func sendTicket(conn net.Conn) {
}

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

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadByte()
		if err != nil {
			log.Println("Read error:", err)
			return
		}
		log.Println("Received message:", msg)
		switch msg {
		case 0x20:
			handlePlate(conn)
		case 0x40:
			handleWantHeartbeat(conn)
		case 0x80:
			handleIAmCamera(conn)
		case 0x81:
			handleIAmDispatcher(conn)
		default:
			sendError(conn, "incorrect message type")
			conn.Close()
			return
		}
	}
}

/*
Number message format:
Type | Hex data    | Value
-------------------------------
u8   |          20 |         32
u8   |          e3 |        227
u16  |       00 20 |         32
u16  |       12 45 |       4677
u16  |       a8 23 |      43043
u32  | 00 00 00 20 |         32
u32  | 00 00 12 45 |       4677
u32  | a6 a9 b5 67 | 2796139879


String message format:
Type | Hex data                   | Value
----------------------------------------------
str  | 00                         | ""
str  | 03 66 6f 6f                | "foo"
str  | 08 45 6C 62 65 72 65 74 68 | "Elbereth"


0x10: Error (Server->Client)

Fields:

    msg: str

Hexadecimal:                            Decoded:
10                                      Error{
03 62 61 64                                 msg: "bad"
                                        }

10                                      Error{
0b 69 6c 6c 65 67 61 6c 20 6d 73 67         msg: "illegal msg"
                                        }

0x20: Plate (Client->Server)

Fields:

    plate: str
    timestamp: u32

Hexadecimal:                Decoded:
20                          Plate{
04 55 4e 31 58                  plate: "UN1X",
00 00 03 e8                     timestamp: 1000
                            }

20                          Plate{
07 52 45 30 35 42 4b 47         plate: "RE05BKG",
00 01 e2 40                     timestamp: 123456
                            }

*/

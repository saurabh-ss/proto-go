package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/saurabh/protohackers/internal/logger"
)

var port = flag.String("port", "50001", "Port to listen on")

func validUsername(username string) bool {
	for _, r := range username {
		if !((r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9')) {
			return false
		}
	}
	return len(username) > 0
}

type ChatRoom struct {
	clients   map[net.Conn]string // conn -> username
	usernames map[string]net.Conn // username -> conn
	mu        sync.RWMutex
	greeting  string
}

func NewChatRoom() *ChatRoom {
	return &ChatRoom{
		clients:   make(map[net.Conn]string),
		usernames: make(map[string]net.Conn),
		greeting:  "Welcome to budgetchat! What shall I call you?\n",
	}
}

func (cr *ChatRoom) AddUser(conn net.Conn, username string) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if _, exists := cr.usernames[username]; exists {
		return fmt.Errorf("username already exists: %s", username)
	}

	if !validUsername(username) {
		return fmt.Errorf("invalid username: %s", username)
	}

	var currUsers []string
	for uName, uConn := range cr.usernames {
		// * bob has entered the room
		currUsers = append(currUsers, uName)
		uConn.Write([]byte("*" + username + " has entered the room\n"))
	}

	// * The room contains: bob, charlie, dave
	fmt.Fprintf(conn, "* The room contains: %s\n", strings.Join(currUsers, ", "))

	cr.clients[conn] = username
	cr.usernames[username] = conn

	return nil
}

func (cr *ChatRoom) RemoveUser(conn net.Conn) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	username := cr.clients[conn]
	delete(cr.clients, conn)
	delete(cr.usernames, username)

	for _, uConn := range cr.usernames {
		uConn.Write([]byte("*" + username + " has left the room\n"))
	}

}

func (cr *ChatRoom) Broadcast(conn net.Conn, message string) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	username := cr.clients[conn]

	for c := range cr.clients {
		if c == conn {
			continue
		}
		// [bob] hi alice
		c.Write([]byte("[" + username + "] " + message))
	}
}

func (cr *ChatRoom) handleConnection(conn net.Conn) {
	log.Println("New connection from", conn.RemoteAddr())
	defer conn.Close()
	defer log.Println("Connection closed from", conn.RemoteAddr())

	conn.Write([]byte(cr.greeting))

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Read error:", err)
		return
	}

	err = cr.AddUser(conn, strings.TrimSpace(line))
	if err != nil {
		log.Println("Add user error:", err)
		conn.Write([]byte(err.Error() + "\n"))
		return
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("Client closed connection (EOF)")
				cr.RemoveUser(conn)
				return
			}
			log.Println("Read error:", err)
			return
		}
		cr.Broadcast(conn, line)
	}
}

func main() {
	flag.Parse()

	// Setup logging to logs directory
	logFile, err := logger.Setup("budget-chat")
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

	cr := NewChatRoom()

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
		go cr.handleConnection(conn)
	}
}

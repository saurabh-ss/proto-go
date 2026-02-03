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
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/saurabh/protohackers/internal/logger"
)

type File struct {
	Data []byte
}

type VersionedFile struct {
	mu       sync.RWMutex
	Versions []*File
}

type VCS struct {
	files sync.Map // map with name as key and VersionedFile as value
}

func (v *VCS) Put(filename string, data []byte) int {
	val, _ := v.files.LoadOrStore(filename, &VersionedFile{})
	vf := val.(*VersionedFile)

	vf.mu.Lock()
	defer vf.mu.Unlock()

	vf.Versions = append(vf.Versions, &File{Data: data})
	return len(vf.Versions)
}

func (v *VCS) Get(filename string, revision string) ([]byte, error) {
	val, ok := v.files.Load(filename)
	if !ok {
		return nil, fmt.Errorf("no such file")
	}

	vf := val.(*VersionedFile)
	vf.mu.RLock()
	defer vf.mu.RUnlock()

	var r int
	var err error
	if revision == "latest" {
		r = len(vf.Versions)
	} else {
		r, err = strconv.Atoi(revision)
		if err != nil || r < 0 || r > len(vf.Versions) {
			return nil, fmt.Errorf("no such revision")
		}
	}

	return vf.Versions[r-1].Data, nil
}

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

	vcs := VCS{}

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
		go handleConnection(conn, &vcs)
	}
}

func handleConnection(conn net.Conn, vcs *VCS) {
	log.Println("New connection from", conn.RemoteAddr())
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Helper function to write a line and flush
	writeLine := func(s string) error {
		if _, err := writer.WriteString(s + "\n"); err != nil {
			return err
		}
		return writer.Flush()
	}

	writeData := func(data []byte) error {
		if _, err := writer.Write(data); err != nil {
			return err
		}
		return writer.Flush()
	}

	handleHelp := func() {
		writeLine("OK usage: HELP|GET|PUT|LIST")
		writeLine("READY")
	}

	handleGet := func(parts []string) {
		if len(parts) != 3 && len(parts) != 2 {
			writeLine("ERR usage: GET file [revision]")
			writeLine("READY")
			return
		}
		filename := parts[1]
		revision := "latest"
		if len(parts) == 3 {
			revision = parts[2]
		}
		data, err := vcs.Get(filename, revision)
		if err != nil {
			writeLine("ERR " + err.Error())
			writeLine("READY")
			return
		}
		writeLine("OK " + fmt.Sprint(len(data)))
		writeData(data)
		writeLine("READY")
	}

	handlePut := func(parts []string) {
		if len(parts) != 3 {
			writeLine("ERR usage: PUT file length newline data")
			writeLine("READY")
			return
		}
		fileName := parts[1]
		_ = fileName
		length, err := strconv.Atoi(parts[2])
		if err != nil {
			writeLine("ERR usage: PUT file length newline data")
			writeLine("READY")
		}
		if length < 0 {
			writeLine("ERR usage: PUT file length newline data")
			writeLine("READY")
		}
		data := make([]byte, length)
		_, err = io.ReadFull(reader, data)
		log.Println("Read data:", string(data))
		if err != nil {
			writeLine("ERR usage: PUT file length newline data")
			writeLine("READY")
			return
		}
		revision := vcs.Put(fileName, data)
		writeLine("OK r" + strconv.Itoa(revision))
		writeLine("READY")
	}

	handleList := func(parts []string) {
		if len(parts) != 2 {
			writeLine("ERR usage: LIST dir")
			writeLine("READY")
		}

	}

	// Send initial READY message
	if err := writeLine("READY"); err != nil {
		log.Println("Write error:", err)
		return
	}

	// Read and process lines
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println("Read error:", err)
			return
		}
		line = strings.TrimSuffix(line, "\n")
		log.Println("Received message:", line)

		parts := strings.Split(line, " ")
		switch strings.ToUpper(parts[0]) {
		case "HELP":
			handleHelp()
		case "GET":
			handleGet(parts)
		case "PUT":
			handlePut(parts)
		case "LIST":
			handleList(parts)
		default:
			writeLine("ERR illegal method: " + parts[0])
			return
		}
	}

	log.Println("Connection closed from", conn.RemoteAddr())
}

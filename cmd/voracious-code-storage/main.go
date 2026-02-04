package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sort"
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
	Versions []*File
}

type VCS struct {
	files map[string]*VersionedFile // map with name as key and VersionedFile as value
	mu    sync.RWMutex
}

func (v *VCS) Put(filename string, data []byte) int {
	v.mu.Lock()
	defer v.mu.Unlock()

	vf, ok := v.files[filename]
	if !ok {
		vf = &VersionedFile{Versions: make([]*File, 0)}
		v.files[filename] = vf
	}
	if n := len(vf.Versions); n > 0 && bytes.Equal(vf.Versions[n-1].Data, data) {
		return n
	}
	vf.Versions = append(vf.Versions, &File{Data: data})
	return len(vf.Versions)
}

func (v *VCS) Get(filename string, revision string) ([]byte, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	vf, ok := v.files[filename]
	if !ok {
		return nil, fmt.Errorf("no such file")
	}

	var r int
	var err error
	if revision == "latest" {
		r = len(vf.Versions)
	} else {
		if revision[0] == 'r' {
			r, err = strconv.Atoi(revision[1:])
		} else {
			r, err = strconv.Atoi(revision)
		}
		if err != nil || r <= 0 || r > len(vf.Versions) {
			return nil, fmt.Errorf("no such revision")
		}
	}

	return vf.Versions[r-1].Data, nil
}

func (v *VCS) GetLatestVersion(filename string) int {
	v.mu.RLock()
	defer v.mu.RUnlock()

	vf, ok := v.files[filename]
	if !ok {
		return 0
	}

	return len(vf.Versions)
}

func (v *VCS) List(dir string) ([]string, []string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	// Normalize directory path - ensure it ends with /
	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}

	files := make([]string, 0)
	dirSet := make(map[string]bool) // Use map to track unique directories

	for file := range v.files {
		if after, ok := strings.CutPrefix(file, dir); ok {
			// Get the relative path after the directory
			relativePath := after

			if strings.Contains(relativePath, "/") {
				// It's in a subdirectory - extract the first directory component
				dirName := strings.Split(relativePath, "/")[0]
				dirSet[dirName] = true
			} else if relativePath != "" {
				// It's a file directly in this directory
				files = append(files, relativePath)
			}
		}
	}

	// Convert map keys to slice
	directories := make([]string, 0, len(dirSet))
	for dir := range dirSet {
		directories = append(directories, dir)
	}

	sort.Strings(files)
	sort.Strings(directories)
	return files, directories, nil
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

	vcs := &VCS{files: make(map[string]*VersionedFile)}

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
		go handleConnection(conn, vcs)
	}
}

func isValidPath(path string, allowDirectory bool) bool {
	if len(path) == 0 || path[0] != '/' {
		return false
	}

	endsWithSlash := path[len(path)-1] == '/'

	// If it ends with /, it must be a directory
	if endsWithSlash && !allowDirectory {
		return false
	}

	if strings.Contains(path, "//") {
		return false
	}

	for _, char := range path {
		if (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '/' || char == '-' || char == '_' || char == '.' {
			continue
		}

		return false
	}

	return true
}

// isValidFilename checks if a filename is valid (must not end with /)
func isValidFilename(filename string) bool {
	return isValidPath(filename, false)
}

// isValidDirectory checks if a directory path is valid (can end with / or not)
func isValidDirectory(dir string) bool {
	return isValidPath(dir, true)
}

func isValidTextData(data []byte) bool {
	for _, b := range data {
		if !(b >= 32 && b <= 126) && b != '\n' && b != '\r' && b != '\t' {
			return false
		}
	}
	return true
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
		if !isValidFilename(fileName) {
			writeLine("ERR illegal file name")
			writeLine("READY")
			return
		}
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

		// Validate that data is text only
		if !isValidTextData(data) {
			writeLine("ERR text files only")
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
			return
		}

		dir := parts[1]
		if !isValidDirectory(dir) {
			writeLine("ERR illegal dir name")
			writeLine("READY")
			return
		}

		files, directories, err := vcs.List(dir)
		if err != nil {
			log.Println("List error:", err)
			return
		}
		log.Println("files:", files)
		log.Println("directories:", directories)
		total := len(files) + len(directories)
		writeLine("OK " + strconv.Itoa(total))

		for _, directory := range directories {
			writeLine(directory + "/" + " DIR")
		}
		for _, file := range files {
			writeLine(file + " r" + strconv.Itoa(vcs.GetLatestVersion(file)))
		}

		writeLine("READY")

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

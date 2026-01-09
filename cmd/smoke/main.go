package main

import (
	"io"
	"log"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":10001")
	if err != nil {
		panic(err)
	}
	log.Println("Listening on port 10001")
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

	buf := make([]byte, 1024*8) // 8KB buffer

	for {
		n, err := conn.Read(buf)

		// If we read some data, echo it back before handling errors
		if n > 0 {
			log.Printf("Got %d bytes: %s", n, string(buf[:n]))

			written, writeErr := conn.Write(buf[:n])
			if writeErr != nil {
				log.Println("Write error:", writeErr)
				return
			}
			if written != n {
				log.Printf("Written bytes mismatch: wrote %d, read %d", written, n)
				return
			}
		}

		// Now handle read errors
		if err != nil {
			if err == io.EOF {
				log.Println("Client closed connection (EOF)")
				return // Close connection gracefully
			}
			log.Println("Read error:", err)
			return
		}
	}
}

package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math/bits"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/saurabh/protohackers/internal/logger"
)

var port = flag.String("port", "50001", "Port to listen on")

// Cipher operation constants
const (
	OpNoop        byte = 0x00
	OpReverseBits byte = 0x01
	OpXor         byte = 0x02
	OpXorPos      byte = 0x03
	OpAdd         byte = 0x04
	OpAddPos      byte = 0x05
	NewLine       byte = 0x0A
	Comma         byte = 0x2C
	XChar         byte = 0x78
)

// CipherOp represents a single cipher operation
type CipherOp struct {
	Op  byte
	Arg byte // only used for OpXor and OpAdd
}

// Helper functions to create cipher operations
func NoopOp() CipherOp {
	return CipherOp{Op: OpNoop}
}

func ReverseBitsOp() CipherOp {
	return CipherOp{Op: OpReverseBits}
}

func XorOp(n byte) CipherOp {
	return CipherOp{Op: OpXor, Arg: n}
}

func XorPosOp() CipherOp {
	return CipherOp{Op: OpXorPos}
}

func AddOp(n byte) CipherOp {
	return CipherOp{Op: OpAdd, Arg: n}
}

func AddPosOp() CipherOp {
	return CipherOp{Op: OpAddPos}
}

// parseCipher converts a byte array into a slice of CipherOp
func parseCipher(cipherBytes []byte) []CipherOp {
	var ops []CipherOp
	i := 0
	for i < len(cipherBytes) {
		switch cipherBytes[i] {
		case OpNoop:
			ops = append(ops, NoopOp())
			i++
		case OpReverseBits:
			ops = append(ops, ReverseBitsOp())
			i++
		case OpXor:
			if i+1 < len(cipherBytes) {
				ops = append(ops, XorOp(cipherBytes[i+1]))
				i += 2
			} else {
				i++
			}
		case OpXorPos:
			ops = append(ops, XorPosOp())
			i++
		case OpAdd:
			if i+1 < len(cipherBytes) {
				ops = append(ops, AddOp(cipherBytes[i+1]))
				i += 2
			} else {
				i++
			}
		case OpAddPos:
			ops = append(ops, AddPosOp())
			i++
		default:
			i++
		}
	}
	return ops
}

func main() {
	flag.Parse()

	// Setup logging to logs directory
	logFile, err := logger.Setup("insecure-sockets-layer")
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
	writer := bufio.NewWriter(conn)

	data, err := reader.ReadBytes(0x00)
	if err != nil {
		log.Println("Read error:", err)
		return
	}
	log.Printf("Received cipher data (hex): %x\n", data)

	cipher := parseCipher(data)
	log.Printf("Parsed cipher: %d operations\n", len(cipher))

	for {
		data, err := reader.ReadBytes(encrypt([]byte{NewLine}, cipher)[0])
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println("Read error:", err)
			return
		}

		log.Println("Received message:", data)

		decrypted := decrypt(data, cipher)
		if bytes.Equal(decrypted, data) {
			log.Println("Decrypted message is the same as the original message")
			return
		}

		decString := string(decrypted)
		log.Println("Decrypted message:", decString)

		result := getMaxCountPart(decString)
		log.Println("Part with max count: ", result)

		encrypted := encrypt([]byte(result), cipher)

		writer.Write(encrypted)
		writer.Flush()
	}

}

func decrypt(data []byte, cipher []CipherOp) []byte {
	if len(data) == 0 || len(cipher) == 0 {
		return data
	}

	// Apply inverse operations in reverse order
	for i := len(cipher) - 1; i >= 0; i-- {
		op := cipher[i]
		switch op.Op {
		case OpNoop:
			// do nothing
		case OpReverseBits:
			data = reverseBits(data) // reverseBits is its own inverse
		case OpXor:
			data = xor(data, op.Arg) // XOR is its own inverse
		case OpXorPos:
			data = xorPos(data) // XorPos is its own inverse
		case OpAdd:
			data = sub(data, op.Arg) // inverse of add is sub
		case OpAddPos:
			data = subPos(data) // inverse of addPos is subPos
		}
	}
	return data
}

func encrypt(data []byte, cipher []CipherOp) []byte {
	if len(data) == 0 || len(cipher) == 0 {
		return data
	}

	for _, op := range cipher {
		switch op.Op {
		case OpNoop:
			// do nothing
		case OpReverseBits:
			data = reverseBits(data)
		case OpXor:
			data = xor(data, op.Arg)
		case OpXorPos:
			data = xorPos(data)
		case OpAdd:
			data = add(data, op.Arg)
		case OpAddPos:
			data = addPos(data)
		}
	}
	return data
}

func reverseBits(data []byte) []byte {
	result := make([]byte, len(data))
	for i := range data {
		result[i] = bits.Reverse8(data[i])
	}
	return result
}

func xor(data []byte, N byte) []byte {
	result := make([]byte, len(data))
	for i := range data {
		result[i] = data[i] ^ N
	}
	return result
}

func xorPos(data []byte) []byte {
	result := make([]byte, len(data))
	for i := range data {
		result[i] = data[i] ^ byte(i)
	}
	return result
}

func add(data []byte, N byte) []byte {
	result := make([]byte, len(data))
	for i := range data {
		result[i] = (data[i] + N) & 0xFF
	}
	return result
}

func addPos(data []byte) []byte {
	result := make([]byte, len(data))
	for i := range data {
		result[i] = (data[i] + byte(i)) & 0xFF
	}
	return result
}

func sub(data []byte, N byte) []byte {
	result := make([]byte, len(data))
	for i := range data {
		result[i] = data[i] - N
	}
	return result
}

func subPos(data []byte) []byte {
	result := make([]byte, len(data))
	for i := range data {
		result[i] = data[i] - byte(i)
	}
	return result
}

func getMaxCountPart(data string) string {
	parts := strings.Split(data, ",")
	result := ""
	maxCount := 0
	for _, part := range parts {
		var count int
		_, err := fmt.Sscanf(part, "%dx", &count)
		if err != nil {
			panic("Couldn't find count in part: " + part)
		}
		if count > maxCount {
			maxCount = count
			result = part
		}
	}
	return strings.TrimSpace(result) + "\n"
}

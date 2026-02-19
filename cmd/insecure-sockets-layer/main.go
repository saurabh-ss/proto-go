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

// NoopOp Helper functions to create cipher operations
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

// readCipher reads cipher ops from the stream byte-by-byte, consuming argument
// bytes explicitly so that a 0x00 argument to Xor/Add does not terminate early.
// The cipher spec ends at a 0x00 opcode (OpNoop used as terminator).
func readCipher(r *bufio.Reader, clog *log.Logger) ([]CipherOp, error) {
	var ops []CipherOp
	var raw []byte
	for {
		opByte, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		raw = append(raw, opByte)
		switch opByte {
		case OpNoop:
			// 0x00 is the end-of-cipher terminator
			clog.Printf("Received cipher data (hex): %x\n", raw)
			clog.Printf("Parsed cipher: %d operations\n", len(ops))
			return ops, nil
		case OpReverseBits:
			ops = append(ops, ReverseBitsOp())
		case OpXorPos:
			ops = append(ops, XorPosOp())
		case OpAddPos:
			ops = append(ops, AddPosOp())
		case OpXor:
			arg, err := r.ReadByte()
			if err != nil {
				return nil, err
			}
			raw = append(raw, arg)
			ops = append(ops, XorOp(arg))
		case OpAdd:
			arg, err := r.ReadByte()
			if err != nil {
				return nil, err
			}
			raw = append(raw, arg)
			ops = append(ops, AddOp(arg))
		}
	}
}

// isCipherNoop returns true if the cipher has no observable effect on any byte
// at any stream position. Position-dependent ops (XorPos, AddPos) use byte(pos),
// which wraps every 256 bytes, so checking all 256 positions is exhaustive.
func isCipherNoop(ops []CipherOp) bool {
	for pos := range 256 {
		for b := range 256 {
			in := []byte{byte(b)}
			out := encrypt(in, ops, pos)
			if out[0] != in[0] {
				return false
			}
		}
	}
	return true
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
	addr := conn.RemoteAddr().String()
	clog := log.New(log.Writer(), fmt.Sprintf("[%s] ", addr), log.Flags())

	clog.Println("New connection")
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	cipher, err := readCipher(reader, clog)
	if err != nil {
		clog.Println("Error reading cipher:", err)
		return
	}
	clog.Printf("Parsed cipher: %d operations\n", len(cipher))

	if isCipherNoop(cipher) {
		clog.Println("Cipher is a no-op, dropping connection")
		return
	}

	// Initialize position counters for the connection
	requestPos := 0
	responsePos := 0

	remaining := []byte{}

	// Buffer. Clients won't send lines longer than 5000 characters.
	buffer := make([]byte, 8192)
	for {
		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			clog.Println("Read error:", err)
			return
		}
		chunk := buffer[:n]
		clog.Printf("Received %d bytes (hex): %x\n", n, chunk)

		decrypted := decrypt(chunk, cipher, requestPos)
		requestPos += n

		clog.Println("Decrypted:", string(decrypted))

		remaining = append(remaining, decrypted...)

		for {
			idx := bytes.IndexByte(remaining, NewLine)
			if idx == -1 {
				break
			}

			line := remaining[:idx+1]
			remaining = remaining[idx+1:]

			result := getMaxCountPartFromDecrypted(line)
			clog.Println("Responding with:", string(result))

			encrypted := encrypt(result, cipher, responsePos)
			responsePos += len(result)

			writer.Write(encrypted)
			writer.Flush()
		}
	}
}

func decrypt(data []byte, cipher []CipherOp, pos int) []byte {
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
			data = xorPos(data, pos) // XorPos is its own inverse
		case OpAdd:
			data = sub(data, op.Arg) // inverse of add is sub
		case OpAddPos:
			data = subPos(data, pos) // inverse of addPos is subPos
		}
	}
	return data
}

func encrypt(data []byte, cipher []CipherOp, pos int) []byte {
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
			data = xorPos(data, pos)
		case OpAdd:
			data = add(data, op.Arg)
		case OpAddPos:
			data = addPos(data, pos)
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

func xorPos(data []byte, offset int) []byte {
	result := make([]byte, len(data))
	for i := range data {
		result[i] = data[i] ^ byte(offset+i)
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

func addPos(data []byte, offset int) []byte {
	result := make([]byte, len(data))
	for i := range data {
		result[i] = (data[i] + byte(offset+i)) & 0xFF
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

func subPos(data []byte, offset int) []byte {
	result := make([]byte, len(data))
	for i := range data {
		result[i] = data[i] - byte(offset+i)
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

func getMaxCountPartFromDecrypted(data []byte) []byte {
	// Find the newline to get the complete line
	before, _, ok := bytes.Cut(data, []byte{NewLine})
	if !ok {
		panic("No newline found in decrypted message")
	}

	// Work with the line up to the newline
	line := before

	// Split by comma
	var maxPart []byte
	maxCount := 0
	start := 0

	for i := 0; i <= len(line); i++ {
		// Check if we hit a comma or end of line
		if i == len(line) || line[i] == Comma {
			part := line[start:i]

			// Find 'x' in this part
			before, _, ok := bytes.Cut(part, []byte{XChar})
			if !ok {
				log.Println("No 'x' found in part: " + string(part))
				continue
			}

			// Parse the count (everything before 'x')
			countStr := string(before)
			var count int
			_, err := fmt.Sscanf(countStr, "%d", &count)
			if err != nil {
				log.Println("Couldn't parse count in part: " + string(part))
				continue
			}

			// Update max if this count is larger
			if count > maxCount {
				maxCount = count
				maxPart = bytes.TrimSpace(part)
			}

			// Move start to after the comma
			start = i + 1
		}
	}

	// Append newline and return
	result := make([]byte, len(maxPart)+1)
	copy(result, maxPart)
	result[len(maxPart)] = NewLine
	return result
}

package main

import (
	"bytes"
	"testing"
)

func TestReverseBits(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "single byte",
			input:    []byte{0b10101010},
			expected: []byte{0b01010101},
		},
		{
			name:     "multiple bytes",
			input:    []byte{0b11110000, 0b00001111},
			expected: []byte{0b00001111, 0b11110000},
		},
		{
			name:     "empty slice",
			input:    []byte{},
			expected: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reverseBits(tt.input)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("reverseBits(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestXor(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		N        byte
		expected []byte
	}{
		{
			name:     "basic xor",
			input:    []byte{0xFF, 0x00, 0xAA},
			N:        0xFF,
			expected: []byte{0x00, 0xFF, 0x55},
		},
		{
			name:     "xor with zero",
			input:    []byte{0x12, 0x34},
			N:        0x00,
			expected: []byte{0x12, 0x34},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := xor(tt.input, tt.N)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("xor(%v, %x) = %v, want %v", tt.input, tt.N, result, tt.expected)
			}
		})
	}
}

func TestXorPos(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		offset   int
		expected []byte
	}{
		{
			name:     "xor index with each byte, offset 0",
			input:    []byte{0xFF, 0xFF, 0xFF},
			offset:   0,
			expected: []byte{0xFF, 0xFE, 0xFD}, // 0xFF^0, 0xFF^1, 0xFF^2
		},
		{
			name:     "xor with zero bytes, offset 0",
			input:    []byte{0x00, 0x00, 0x00, 0x00},
			offset:   0,
			expected: []byte{0x00, 0x01, 0x02, 0x03}, // 0^0, 0^1, 0^2, 0^3
		},
		{
			name:     "xor pattern, offset 0",
			input:    []byte{0x0A, 0x0B, 0x0C, 0x0D},
			offset:   0,
			expected: []byte{0x0A, 0x0A, 0x0E, 0x0E}, // 0x0A^0, 0x0B^1, 0x0C^2, 0x0D^3
		},
		{
			name:     "xor with offset 10",
			input:    []byte{0xFF, 0xFF, 0xFF},
			offset:   10,
			expected: []byte{0xF5, 0xF4, 0xF3}, // 0xFF^10, 0xFF^11, 0xFF^12
		},
		{
			name:     "empty slice",
			input:    []byte{},
			offset:   0,
			expected: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := xorPos(tt.input, tt.offset)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("xorPos(%v, %d) = %v, want %v", tt.input, tt.offset, result, tt.expected)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		N        byte
		expected []byte
	}{
		{
			name:     "basic add",
			input:    []byte{0x01, 0x02, 0x03},
			N:        0x10,
			expected: []byte{0x11, 0x12, 0x13},
		},
		{
			name:     "add with overflow",
			input:    []byte{0xFF, 0xFE},
			N:        0x02,
			expected: []byte{0x01, 0x00},
		},
		{
			name:     "add with wrap around",
			input:    []byte{0xFF},
			N:        0x01,
			expected: []byte{0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := add(tt.input, tt.N)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("add(%v, %x) = %v, want %v", tt.input, tt.N, result, tt.expected)
			}
		})
	}
}

func TestAddPos(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		offset   int
		expected []byte
	}{
		{
			name:     "add index to each byte, offset 0",
			input:    []byte{0x00, 0x10, 0x20},
			offset:   0,
			expected: []byte{0x00, 0x11, 0x22}, // adds 0, 1, 2 respectively
		},
		{
			name:     "add with overflow, offset 0",
			input:    []byte{0xFF, 0xFE, 0xFD},
			offset:   0,
			expected: []byte{0xFF, 0xFF, 0xFF}, // 0xFF+0, 0xFE+1, 0xFD+2
		},
		{
			name:     "simple sequence, offset 0",
			input:    []byte{0x0A, 0x0A, 0x0A, 0x0A},
			offset:   0,
			expected: []byte{0x0A, 0x0B, 0x0C, 0x0D}, // 10+0, 10+1, 10+2, 10+3
		},
		{
			name:     "with offset 5",
			input:    []byte{0x00, 0x00, 0x00},
			offset:   5,
			expected: []byte{0x05, 0x06, 0x07}, // 0+5, 0+6, 0+7
		},
		{
			name:     "empty slice",
			input:    []byte{},
			offset:   0,
			expected: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addPos(tt.input, tt.offset)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("addPos(%v, %d) = %v, want %v", tt.input, tt.offset, result, tt.expected)
			}
		})
	}
}

func TestSub(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		N        byte
		expected []byte
	}{
		{
			name:     "basic sub",
			input:    []byte{0x11, 0x12, 0x13},
			N:        0x10,
			expected: []byte{0x01, 0x02, 0x03},
		},
		{
			name:     "sub with underflow",
			input:    []byte{0x01, 0x00},
			N:        0x02,
			expected: []byte{0xFF, 0xFE}, // wraps around
		},
		{
			name:     "sub zero",
			input:    []byte{0x12, 0x34},
			N:        0x00,
			expected: []byte{0x12, 0x34},
		},
		{
			name:     "inverse of add",
			input:    []byte{0x50, 0x60, 0x70},
			N:        0x20,
			expected: []byte{0x30, 0x40, 0x50},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sub(tt.input, tt.N)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("sub(%v, %x) = %v, want %v", tt.input, tt.N, result, tt.expected)
			}
		})
	}
}

func TestSubPos(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		offset   int
		expected []byte
	}{
		{
			name:     "sub index from each byte, offset 0",
			input:    []byte{0x00, 0x11, 0x22},
			offset:   0,
			expected: []byte{0x00, 0x10, 0x20}, // subs 0, 1, 2 respectively
		},
		{
			name:     "sub with underflow, offset 0",
			input:    []byte{0x00, 0x00, 0x00},
			offset:   0,
			expected: []byte{0x00, 0xFF, 0xFE}, // 0-0, 0-1, 0-2 with wraparound
		},
		{
			name:     "inverse of addPos, offset 0",
			input:    []byte{0x0A, 0x0B, 0x0C, 0x0D},
			offset:   0,
			expected: []byte{0x0A, 0x0A, 0x0A, 0x0A}, // 10-0, 11-1, 12-2, 13-3
		},
		{
			name:     "with offset 5",
			input:    []byte{0x05, 0x06, 0x07},
			offset:   5,
			expected: []byte{0x00, 0x00, 0x00}, // 5-5, 6-6, 7-7
		},
		{
			name:     "empty slice",
			input:    []byte{},
			offset:   0,
			expected: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := subPos(tt.input, tt.offset)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("subPos(%v, %d) = %v, want %v", tt.input, tt.offset, result, tt.expected)
			}
		})
	}
}

func TestAddSubInverse(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		N     byte
	}{
		{
			name:  "add then sub",
			input: []byte{0x10, 0x20, 0x30, 0x40},
			N:     0x15,
		},
		{
			name:  "with overflow",
			input: []byte{0xFF, 0xFE, 0x00, 0x01},
			N:     0x10,
		},
		{
			name:  "edge cases",
			input: []byte{0x00, 0xFF, 0x7F, 0x80},
			N:     0xFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted := add(tt.input, tt.N)
			decrypted := sub(encrypted, tt.N)
			if !bytes.Equal(decrypted, tt.input) {
				t.Errorf("add then sub: got %v, want %v", decrypted, tt.input)
			}
		})
	}
}

func TestAddPosSubPosInverse(t *testing.T) {
	tests := []struct {
		name   string
		input  []byte
		offset int
	}{
		{
			name:   "addPos then subPos, offset 0",
			input:  []byte{0x10, 0x20, 0x30, 0x40, 0x50},
			offset: 0,
		},
		{
			name:   "with potential overflow, offset 0",
			input:  []byte{0xFF, 0xFE, 0xFD, 0xFC},
			offset: 0,
		},
		{
			name:   "all zeros, offset 0",
			input:  []byte{0x00, 0x00, 0x00, 0x00},
			offset: 0,
		},
		{
			name:   "sequential, offset 0",
			input:  []byte{0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
			offset: 0,
		},
		{
			name:   "with offset 100",
			input:  []byte{0x10, 0x20, 0x30},
			offset: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted := addPos(tt.input, tt.offset)
			decrypted := subPos(encrypted, tt.offset)
			if !bytes.Equal(decrypted, tt.input) {
				t.Errorf("addPos then subPos: got %v, want %v", decrypted, tt.input)
			}
		})
	}
}

func TestEncrypt(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		cipher   []CipherOp
		pos      int
		expected []byte
	}{
		{
			name:  "xor(1),reversebits",
			input: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
			cipher: []CipherOp{
				XorOp(0x01),
				ReverseBitsOp(),
			},
			pos:      0,
			expected: []byte{0x96, 0x26, 0xb6, 0xb6, 0x76},
		},
		{
			name:  "addpos,addpos with offset 0",
			input: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
			cipher: []CipherOp{
				AddPosOp(),
				AddPosOp(),
			},
			pos:      0,
			expected: []byte{0x68, 0x67, 0x70, 0x72, 0x77},
		},
		{
			name:  "no-op with xor",
			input: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21},
			cipher: []CipherOp{
				XorOp(0xa0),
				XorOp(0x0b),
				XorOp(0xab),
			},
			pos:      0,
			expected: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21},
		},
		{
			name:  "no-op reverse twice",
			input: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21},
			cipher: []CipherOp{
				ReverseBitsOp(),
				ReverseBitsOp(),
			},
			pos:      0,
			expected: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21},
		},
		{
			name:  "noop operation",
			input: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
			cipher: []CipherOp{
				NoopOp(),
				NoopOp(),
			},
			pos:      0,
			expected: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
		},
		{
			name:  "addpos with offset 10",
			input: []byte{0x00, 0x00, 0x00},
			cipher: []CipherOp{
				AddPosOp(),
			},
			pos:      10,
			expected: []byte{0x0A, 0x0B, 0x0C}, // 0+10, 0+11, 0+12
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encrypt(tt.input, tt.cipher, tt.pos)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("encrypt(%v, %v, %d) = %v, want %v", tt.input, tt.cipher, tt.pos, result, tt.expected)
			}
		})
	}
}

func TestDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		encrypted []byte
		cipher    []CipherOp
		pos       int
		expected  []byte
	}{
		{
			name:      "xor(1),reversebits",
			encrypted: []byte{0x96, 0x26, 0xb6, 0xb6, 0x76},
			cipher: []CipherOp{
				XorOp(0x01),
				ReverseBitsOp(),
			},
			pos:      0,
			expected: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
		},
		{
			name:      "addpos,addpos with offset 0",
			encrypted: []byte{0x68, 0x67, 0x70, 0x72, 0x77},
			cipher: []CipherOp{
				AddPosOp(),
				AddPosOp(),
			},
			pos:      0,
			expected: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
		},
		{
			name:      "no-op with xor",
			encrypted: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21},
			cipher: []CipherOp{
				XorOp(0xa0),
				XorOp(0x0b),
				XorOp(0xab),
			},
			pos:      0,
			expected: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21},
		},
		{
			name:      "no-op reverse twice",
			encrypted: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21},
			cipher: []CipherOp{
				ReverseBitsOp(),
				ReverseBitsOp(),
			},
			pos:      0,
			expected: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21},
		},
		{
			name:      "noop operation",
			encrypted: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
			cipher: []CipherOp{
				NoopOp(),
				NoopOp(),
			},
			pos:      0,
			expected: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
		},
		{
			name:      "complex cipher with add",
			encrypted: []byte{0x78, 0x76, 0x7d, 0x7d, 0x80},
			cipher: []CipherOp{
				AddOp(0x10),
			},
			pos:      0,
			expected: []byte{0x68, 0x66, 0x6d, 0x6d, 0x70},
		},
		{
			name:      "xorpos operation with offset 0",
			encrypted: []byte{0x68, 0x64, 0x6e, 0x6f, 0x6b},
			cipher: []CipherOp{
				XorPosOp(),
			},
			pos:      0,
			expected: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
		},
		{
			name:      "addpos with offset 10",
			encrypted: []byte{0x0A, 0x0B, 0x0C},
			cipher: []CipherOp{
				AddPosOp(),
			},
			pos:      10,
			expected: []byte{0x00, 0x00, 0x00},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decrypt(tt.encrypted, tt.cipher, tt.pos)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("decrypt(%v, %v, %d) = %v, want %v", tt.encrypted, tt.cipher, tt.pos, result, tt.expected)
			}
		})
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		plaintext []byte
		cipher    []CipherOp
		pos       int
	}{
		{
			name:      "xor(1),reversebits",
			plaintext: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
			cipher: []CipherOp{
				XorOp(0x01),
				ReverseBitsOp(),
			},
			pos: 0,
		},
		{
			name:      "addpos,addpos with offset 0",
			plaintext: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
			cipher: []CipherOp{
				AddPosOp(),
				AddPosOp(),
			},
			pos: 0,
		},
		{
			name:      "hello world",
			plaintext: []byte("hello world!"),
			cipher: []CipherOp{
				XorOp(0xa0),
				XorOp(0x0b),
				XorOp(0xab),
			},
			pos: 0,
		},
		{
			name:      "complex cipher",
			plaintext: []byte("The quick brown fox jumps over the lazy dog"),
			cipher: []CipherOp{
				AddOp(0x10),
				XorOp(0x55),
				ReverseBitsOp(),
				AddPosOp(),
				XorPosOp(),
			},
			pos: 0,
		},
		{
			name:      "all operations",
			plaintext: []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF},
			cipher: []CipherOp{
				NoopOp(),
				ReverseBitsOp(),
				XorOp(0xAA),
				XorPosOp(),
				AddOp(0x42),
				AddPosOp(),
				ReverseBitsOp(),
			},
			pos: 0,
		},
		{
			name:      "empty cipher",
			plaintext: []byte("test"),
			cipher:    []CipherOp{},
			pos:       0,
		},
		{
			name:      "only noops",
			plaintext: []byte("noop test"),
			cipher: []CipherOp{
				NoopOp(),
				NoopOp(),
				NoopOp(),
			},
			pos: 0,
		},
		{
			name:      "addpos with offset 50",
			plaintext: []byte("hello"),
			cipher: []CipherOp{
				AddPosOp(),
			},
			pos: 50,
		},
		{
			name:      "xorpos with offset 100",
			plaintext: []byte{0xFF, 0xFE, 0xFD},
			cipher: []CipherOp{
				XorPosOp(),
			},
			pos: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted := encrypt(tt.plaintext, tt.cipher, tt.pos)
			decrypted := decrypt(encrypted, tt.cipher, tt.pos)
			if !bytes.Equal(decrypted, tt.plaintext) {
				t.Errorf("Round trip failed: plaintext=%v, encrypted=%v, decrypted=%v",
					tt.plaintext, encrypted, decrypted)
			}
		})
	}
}

func TestGetMaxCountPart(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "example from spec",
			input:    "10x toy car,15x dog on a string,4x inflatable motorcycle",
			expected: "15x dog on a string\n",
		},
		{
			name:     "single item",
			input:    "5x widget",
			expected: "5x widget\n",
		},
		{
			name:     "max at beginning",
			input:    "100x first,50x second,25x third",
			expected: "100x first\n",
		},
		{
			name:     "max at end",
			input:    "10x first,20x second,30x third",
			expected: "30x third\n",
		},
		{
			name:     "all same count",
			input:    "5x first,5x second,5x third",
			expected: "5x first\n", // Ties can be broken arbitrarily
		},
		{
			name:     "with whitespace",
			input:    " 10x item1 , 20x item2 , 15x item3 ",
			expected: "20x item2\n",
		},
		{
			name:     "large numbers",
			input:    "1000x small,9999x large,500x medium",
			expected: "9999x large\n",
		},
		{
			name:     "zero count",
			input:    "0x nothing,5x something,2x other",
			expected: "5x something\n",
		},
		{
			name:     "item descriptions without commas",
			input:    "5x red ball,10x blue square,3x green triangle",
			expected: "10x blue square\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMaxCountPart(tt.input)
			if result != tt.expected {
				t.Errorf("getMaxCountPart(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetMaxCountPartEdgeCases(t *testing.T) {
	// Test that function panics on invalid input
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for invalid input")
		}
	}()

	// This should panic because "invalid" doesn't match the pattern
	getMaxCountPart("invalid,no,numbers")
}

func TestMultipleMessagesWithPositionCounter(t *testing.T) {
	// Simulate multiple messages in a connection with addPos/xorPos
	cipher := []CipherOp{AddPosOp()}

	// First message: "hello" (5 bytes) at position 0
	msg1 := []byte("hello")
	encrypted1 := encrypt(msg1, cipher, 0)
	decrypted1 := decrypt(encrypted1, cipher, 0)
	if !bytes.Equal(decrypted1, msg1) {
		t.Errorf("Message 1 round trip failed: got %v, want %v", decrypted1, msg1)
	}

	// Second message: "world" (5 bytes) at position 5 (after first message)
	msg2 := []byte("world")
	encrypted2 := encrypt(msg2, cipher, 5)
	decrypted2 := decrypt(encrypted2, cipher, 5)
	if !bytes.Equal(decrypted2, msg2) {
		t.Errorf("Message 2 round trip failed: got %v, want %v", decrypted2, msg2)
	}

	// Third message: "!!!" (3 bytes) at position 10 (after first two messages)
	msg3 := []byte("!!!")
	encrypted3 := encrypt(msg3, cipher, 10)
	decrypted3 := decrypt(encrypted3, cipher, 10)
	if !bytes.Equal(decrypted3, msg3) {
		t.Errorf("Message 3 round trip failed: got %v, want %v", decrypted3, msg3)
	}

	// Verify that encrypted messages are different even though plaintext might be similar
	// Because they're at different positions in the stream
	if bytes.Equal(encrypted1, encrypted2) && bytes.Equal(msg1, msg2) {
		t.Error("Encrypted messages should differ when positions differ")
	}
}

func TestXorPosAcrossMultipleMessages(t *testing.T) {
	// Test xorPos with position counter across multiple messages
	cipher := []CipherOp{XorPosOp()}

	// Message 1: 3 bytes at position 0
	msg1 := []byte{0xFF, 0xFF, 0xFF}
	encrypted1 := encrypt(msg1, cipher, 0)
	expected1 := []byte{0xFF, 0xFE, 0xFD} // 0xFF^0, 0xFF^1, 0xFF^2
	if !bytes.Equal(encrypted1, expected1) {
		t.Errorf("Message 1: got %v, want %v", encrypted1, expected1)
	}

	// Message 2: 3 bytes at position 3 (continuing from message 1)
	msg2 := []byte{0xFF, 0xFF, 0xFF}
	encrypted2 := encrypt(msg2, cipher, 3)
	expected2 := []byte{0xFC, 0xFB, 0xFA} // 0xFF^3, 0xFF^4, 0xFF^5
	if !bytes.Equal(encrypted2, expected2) {
		t.Errorf("Message 2: got %v, want %v", encrypted2, expected2)
	}

	// Verify decryption works
	if !bytes.Equal(decrypt(encrypted1, cipher, 0), msg1) {
		t.Error("Message 1 decryption failed")
	}
	if !bytes.Equal(decrypt(encrypted2, cipher, 3), msg2) {
		t.Error("Message 2 decryption failed")
	}
}

func TestParseCipher(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []CipherOp
	}{
		{
			name:  "mixed operations",
			input: []byte{0x02, 0x01, 0x01, 0x00},
			expected: []CipherOp{
				XorOp(0x01),
				ReverseBitsOp(),
				NoopOp(),
			},
		},
		{
			name:  "multiple addpos",
			input: []byte{0x05, 0x05, 0x00},
			expected: []CipherOp{
				AddPosOp(),
				AddPosOp(),
				NoopOp(),
			},
		},
		{
			name:  "xor chain",
			input: []byte{0x02, 0xa0, 0x02, 0x0b, 0x02, 0xab, 0x00},
			expected: []CipherOp{
				XorOp(0xa0),
				XorOp(0x0b),
				XorOp(0xab),
				NoopOp(),
			},
		},
		{
			name:  "all operation types",
			input: []byte{0x00, 0x01, 0x02, 0xFF, 0x03, 0x04, 0x10, 0x05},
			expected: []CipherOp{
				NoopOp(),
				ReverseBitsOp(),
				XorOp(0xFF),
				XorPosOp(),
				AddOp(0x10),
				AddPosOp(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCipher(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseCipher(%v) returned %d ops, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i].Op != tt.expected[i].Op || result[i].Arg != tt.expected[i].Arg {
					t.Errorf("parseCipher(%v)[%d] = {Op: %x, Arg: %x}, want {Op: %x, Arg: %x}",
						tt.input, i, result[i].Op, result[i].Arg, tt.expected[i].Op, tt.expected[i].Arg)
				}
			}
		})
	}
}

func TestGetMaxCountPartFromDecrypted(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "example from spec",
			input:    []byte("10x toy car,15x dog on a string,4x inflatable motorcycle\n"),
			expected: []byte("15x dog on a string\n"),
		},
		{
			name:     "single item",
			input:    []byte("5x widget\n"),
			expected: []byte("5x widget\n"),
		},
		{
			name:     "max at beginning",
			input:    []byte("100x first,50x second,25x third\n"),
			expected: []byte("100x first\n"),
		},
		{
			name:     "max at end",
			input:    []byte("10x first,20x second,30x third\n"),
			expected: []byte("30x third\n"),
		},
		{
			name:     "all same count",
			input:    []byte("5x first,5x second,5x third\n"),
			expected: []byte("5x first\n"), // First one wins in ties
		},
		{
			name:     "with whitespace",
			input:    []byte(" 10x item1 , 20x item2 , 15x item3 \n"),
			expected: []byte("20x item2\n"),
		},
		{
			name:     "large numbers",
			input:    []byte("1000x small,9999x large,500x medium\n"),
			expected: []byte("9999x large\n"),
		},
		{
			name:     "zero count",
			input:    []byte("0x nothing,5x something,2x other\n"),
			expected: []byte("5x something\n"),
		},
		{
			name:     "item descriptions without commas",
			input:    []byte("5x red ball,10x blue square,3x green triangle\n"),
			expected: []byte("10x blue square\n"),
		},
		{
			name:     "two items only",
			input:    []byte("3x apple,7x banana\n"),
			expected: []byte("7x banana\n"),
		},
		{
			name:     "with extra data after newline",
			input:    []byte("10x first,20x second\nextra data here"),
			expected: []byte("20x second\n"),
		},
		{
			name:     "multi-digit counts",
			input:    []byte("123x item1,456x item2,789x item3\n"),
			expected: []byte("789x item3\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMaxCountPartFromDecrypted(tt.input)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("getMaxCountPartFromDecrypted(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetMaxCountPartFromDecryptedEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		shouldPanic bool
		panicMsg  string
	}{
		{
			name:        "no newline",
			input:       []byte("10x item"),
			shouldPanic: true,
			panicMsg:    "No newline found",
		},
		{
			name:        "no x character",
			input:       []byte("10 item,20 other\n"),
			shouldPanic: true,
			panicMsg:    "No 'x' found",
		},
		{
			name:        "invalid count format",
			input:       []byte("abcx item,10x other\n"),
			shouldPanic: true,
			panicMsg:    "Couldn't parse count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.shouldPanic {
					if r == nil {
						t.Errorf("Expected panic for input %q", tt.input)
					}
				} else {
					if r != nil {
						t.Errorf("Unexpected panic for input %q: %v", tt.input, r)
					}
				}
			}()
			getMaxCountPartFromDecrypted(tt.input)
		})
	}
}

func TestGetMaxCountPartConsistency(t *testing.T) {
	// Test that both string and byte versions produce the same results
	tests := []string{
		"10x toy car,15x dog on a string,4x inflatable motorcycle",
		"5x widget",
		"100x first,50x second,25x third",
		"10x first,20x second,30x third",
		"1000x small,9999x large,500x medium",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			// Add newline for byte version
			byteInput := []byte(input + "\n")
			
			stringResult := getMaxCountPart(input)
			byteResult := getMaxCountPartFromDecrypted(byteInput)
			
			if string(byteResult) != stringResult {
				t.Errorf("Results differ:\n  string version: %q\n  byte version:   %q", 
					stringResult, string(byteResult))
			}
		})
	}
}

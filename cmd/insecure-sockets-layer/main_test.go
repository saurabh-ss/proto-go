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
		expected []byte
	}{
		{
			name:     "xor index with each byte",
			input:    []byte{0xFF, 0xFF, 0xFF},
			expected: []byte{0xFF, 0xFE, 0xFD}, // 0xFF^0, 0xFF^1, 0xFF^2
		},
		{
			name:     "xor with zero bytes",
			input:    []byte{0x00, 0x00, 0x00, 0x00},
			expected: []byte{0x00, 0x01, 0x02, 0x03}, // 0^0, 0^1, 0^2, 0^3
		},
		{
			name:     "xor pattern",
			input:    []byte{0x0A, 0x0B, 0x0C, 0x0D},
			expected: []byte{0x0A, 0x0A, 0x0E, 0x0E}, // 0x0A^0, 0x0B^1, 0x0C^2, 0x0D^3
		},
		{
			name:     "empty slice",
			input:    []byte{},
			expected: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := xorPos(tt.input)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("xorPos(%v) = %v, want %v", tt.input, result, tt.expected)
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
		expected []byte
	}{
		{
			name:     "add index to each byte",
			input:    []byte{0x00, 0x10, 0x20},
			expected: []byte{0x00, 0x11, 0x22}, // adds 0, 1, 2 respectively
		},
		{
			name:     "add with overflow",
			input:    []byte{0xFF, 0xFE, 0xFD},
			expected: []byte{0xFF, 0xFF, 0xFF}, // 0xFF+0, 0xFE+1, 0xFD+2
		},
		{
			name:     "simple sequence",
			input:    []byte{0x0A, 0x0A, 0x0A, 0x0A},
			expected: []byte{0x0A, 0x0B, 0x0C, 0x0D}, // 10+0, 10+1, 10+2, 10+3
		},
		{
			name:     "empty slice",
			input:    []byte{},
			expected: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addPos(tt.input)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("addPos(%v) = %v, want %v", tt.input, result, tt.expected)
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
		expected []byte
	}{
		{
			name:     "sub index from each byte",
			input:    []byte{0x00, 0x11, 0x22},
			expected: []byte{0x00, 0x10, 0x20}, // subs 0, 1, 2 respectively
		},
		{
			name:     "sub with underflow",
			input:    []byte{0x00, 0x00, 0x00},
			expected: []byte{0x00, 0xFF, 0xFE}, // 0-0, 0-1, 0-2 with wraparound
		},
		{
			name:     "inverse of addPos",
			input:    []byte{0x0A, 0x0B, 0x0C, 0x0D},
			expected: []byte{0x0A, 0x0A, 0x0A, 0x0A}, // 10-0, 11-1, 12-2, 13-3
		},
		{
			name:     "empty slice",
			input:    []byte{},
			expected: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := subPos(tt.input)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("subPos(%v) = %v, want %v", tt.input, result, tt.expected)
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
		name  string
		input []byte
	}{
		{
			name:  "addPos then subPos",
			input: []byte{0x10, 0x20, 0x30, 0x40, 0x50},
		},
		{
			name:  "with potential overflow",
			input: []byte{0xFF, 0xFE, 0xFD, 0xFC},
		},
		{
			name:  "all zeros",
			input: []byte{0x00, 0x00, 0x00, 0x00},
		},
		{
			name:  "sequential",
			input: []byte{0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted := addPos(tt.input)
			decrypted := subPos(encrypted)
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
		expected []byte
	}{
		{
			name:  "xor(1),reversebits",
			input: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
			cipher: []CipherOp{
				XorOp(0x01),
				ReverseBitsOp(),
			},
			expected: []byte{0x96, 0x26, 0xb6, 0xb6, 0x76},
		},
		{
			name:  "addpos,addpos",
			input: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
			cipher: []CipherOp{
				AddPosOp(),
				AddPosOp(),
			},
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
			expected: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21},
		},
		{
			name:  "no-op reverse twice",
			input: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21},
			cipher: []CipherOp{
				ReverseBitsOp(),
				ReverseBitsOp(),
			},
			expected: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21},
		},
		{
			name:  "noop operation",
			input: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
			cipher: []CipherOp{
				NoopOp(),
				NoopOp(),
			},
			expected: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encrypt(tt.input, tt.cipher)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("encrypt(%v, %v) = %v, want %v", tt.input, tt.cipher, result, tt.expected)
			}
		})
	}
}

func TestDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		encrypted []byte
		cipher    []CipherOp
		expected  []byte
	}{
		{
			name:      "xor(1),reversebits",
			encrypted: []byte{0x96, 0x26, 0xb6, 0xb6, 0x76},
			cipher: []CipherOp{
				XorOp(0x01),
				ReverseBitsOp(),
			},
			expected: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
		},
		{
			name:      "addpos,addpos",
			encrypted: []byte{0x68, 0x67, 0x70, 0x72, 0x77},
			cipher: []CipherOp{
				AddPosOp(),
				AddPosOp(),
			},
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
			expected: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21},
		},
		{
			name:      "no-op reverse twice",
			encrypted: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21},
			cipher: []CipherOp{
				ReverseBitsOp(),
				ReverseBitsOp(),
			},
			expected: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x21},
		},
		{
			name:      "noop operation",
			encrypted: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
			cipher: []CipherOp{
				NoopOp(),
				NoopOp(),
			},
			expected: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
		},
		{
			name:      "complex cipher with add",
			encrypted: []byte{0x78, 0x76, 0x7d, 0x7d, 0x80},
			cipher: []CipherOp{
				AddOp(0x10),
			},
			expected: []byte{0x68, 0x66, 0x6d, 0x6d, 0x70},
		},
		{
			name:      "xorpos operation",
			encrypted: []byte{0x68, 0x64, 0x6e, 0x6f, 0x6b},
			cipher: []CipherOp{
				XorPosOp(),
			},
			expected: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decrypt(tt.encrypted, tt.cipher)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("decrypt(%v, %v) = %v, want %v", tt.encrypted, tt.cipher, result, tt.expected)
			}
		})
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		plaintext []byte
		cipher    []CipherOp
	}{
		{
			name:      "xor(1),reversebits",
			plaintext: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
			cipher: []CipherOp{
				XorOp(0x01),
				ReverseBitsOp(),
			},
		},
		{
			name:      "addpos,addpos",
			plaintext: []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f},
			cipher: []CipherOp{
				AddPosOp(),
				AddPosOp(),
			},
		},
		{
			name:      "hello world",
			plaintext: []byte("hello world!"),
			cipher: []CipherOp{
				XorOp(0xa0),
				XorOp(0x0b),
				XorOp(0xab),
			},
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
		},
		{
			name:      "empty cipher",
			plaintext: []byte("test"),
			cipher:    []CipherOp{},
		},
		{
			name:      "only noops",
			plaintext: []byte("noop test"),
			cipher: []CipherOp{
				NoopOp(),
				NoopOp(),
				NoopOp(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted := encrypt(tt.plaintext, tt.cipher)
			decrypted := decrypt(encrypted, tt.cipher)
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
			expected: "5x first\n", // returns first one when all are equal (uses > not >=)
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

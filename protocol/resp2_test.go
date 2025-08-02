package protocol

import (
	"bufio"
	"strings"
	"testing"
)

func TestDecodeCommand(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedCmd   string
		expectedArgs  []string
		expectedError string
	}{
		{
			name:         "SET command with key and value",
			input:        "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n",
			expectedCmd:  "SET",
			expectedArgs: []string{"key", "value"},
		},
		{
			name:         "GET command with key",
			input:        "*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n",
			expectedCmd:  "GET",
			expectedArgs: []string{"key"},
		},
		{
			name:         "DEL command with key",
			input:        "*2\r\n$3\r\nDEL\r\n$8\r\nmykey123\r\n",
			expectedCmd:  "DEL",
			expectedArgs: []string{"mykey123"},
		},
		{
			name:         "PING command with no arguments",
			input:        "*1\r\n$4\r\nPING\r\n",
			expectedCmd:  "PING",
			expectedArgs: []string{},
		},
		{
			name:         "EXPIRE command with key and seconds",
			input:        "*3\r\n$6\r\nEXPIRE\r\n$4\r\ntest\r\n$2\r\n60\r\n",
			expectedCmd:  "EXPIRE",
			expectedArgs: []string{"test", "60"},
		},
		{
			name:         "KEYS command with pattern",
			input:        "*2\r\n$4\r\nKEYS\r\n$1\r\n*\r\n",
			expectedCmd:  "KEYS",
			expectedArgs: []string{"*"},
		},
		{
			name:         "FLUSHALL command",
			input:        "*1\r\n$8\r\nFLUSHALL\r\n",
			expectedCmd:  "FLUSHALL",
			expectedArgs: []string{},
		},
		{
			name:         "Command with empty string argument",
			input:        "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$0\r\n\r\n",
			expectedCmd:  "SET",
			expectedArgs: []string{"key", ""},
		},
		{
			name:         "Command with spaces in value",
			input:        "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$11\r\nhello world\r\n",
			expectedCmd:  "SET",
			expectedArgs: []string{"key", "hello world"},
		},
		// Error cases
		{
			name:          "Invalid format - not starting with *",
			input:         "SET key value\r\n",
			expectedError: "expected array (*), got:",
		},
		{
			name:          "Invalid array length - not a number",
			input:         "*abc\r\n$3\r\nSET\r\n",
			expectedError: "invalid array length:",
		},
		{
			name:          "Zero array length",
			input:         "*0\r\n",
			expectedError: "command must contain at least one element",
		},
		{
			name:          "Negative array length",
			input:         "*-1\r\n",
			expectedError: "command must contain at least one element",
		},
		{
			name:          "Invalid bulk string - not starting with $",
			input:         "*2\r\n#3\r\nSET\r\n$3\r\nkey\r\n",
			expectedError: "expected bulk string ($), got:",
		},
		{
			name:          "Invalid bulk string length - not a number",
			input:         "*2\r\n$abc\r\nSET\r\n$3\r\nkey\r\n",
			expectedError: "invalid bulk string length:",
		},
		{
			name:          "Incomplete command - missing data",
			input:         "*2\r\n$3\r\nSET\r\n$3\r\n",
			expectedError: "EOF",
		},
		{
			name:          "Bulk string length mismatch",
			input:         "*2\r\n$5\r\nSET\r\n$3\r\nkey\r\n",
			expectedError: "expected bulk string ($), got:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
		cmd, args, err := DecodeCommand(reader)

			if tt.expectedError != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, but got nil", tt.expectedError)
				}
				if !containsString(err.Error(), tt.expectedError) {
					t.Fatalf("expected error containing %q, but got %q", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cmd != tt.expectedCmd {
				t.Errorf("expected command %q, got %q", tt.expectedCmd, cmd)
			}

			if len(args) != len(tt.expectedArgs) {
				t.Errorf("expected %d arguments, got %d", len(tt.expectedArgs), len(args))
				return
			}

			for i, expectedArg := range tt.expectedArgs {
				if args[i] != expectedArg {
					t.Errorf("expected argument[%d] %q, got %q", i, expectedArg, args[i])
				}
			}
		})
	}
}

func TestDecodeCommandEdgeCases(t *testing.T) {
	t.Run("Single character command", func(t *testing.T) {
		input := "*1\r\n$1\r\nX\r\n"
		reader := bufio.NewReader(strings.NewReader(input))
	cmd, args, err := DecodeCommand(reader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cmd != "X" {
			t.Errorf("expected command 'X', got %q", cmd)
		}
		if len(args) != 0 {
			t.Errorf("expected 0 arguments, got %d", len(args))
		}
	})

	t.Run("Command with special characters", func(t *testing.T) {
		input := "*3\r\n$3\r\nSET\r\n$7\r\nkey:123\r\n$10\r\nvalue@#$%^\r\n"
		reader := bufio.NewReader(strings.NewReader(input))
	cmd, args, err := DecodeCommand(reader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cmd != "SET" {
			t.Errorf("expected command 'SET', got %q", cmd)
		}
		expected := []string{"key:123", "value@#$%^"}
		if len(args) != len(expected) {
			t.Errorf("expected %d arguments, got %d", len(expected), len(args))
		}
		for i, expectedArg := range expected {
			if args[i] != expectedArg {
				t.Errorf("expected argument[%d] %q, got %q", i, expectedArg, args[i])
			}
		}
	})

	t.Run("Large number of arguments", func(t *testing.T) {
		// MSET key1 val1 key2 val2 key3 val3
		input := "*7\r\n$4\r\nMSET\r\n$4\r\nkey1\r\n$4\r\nval1\r\n$4\r\nkey2\r\n$4\r\nval2\r\n$4\r\nkey3\r\n$4\r\nval3\r\n"
		reader := bufio.NewReader(strings.NewReader(input))
	cmd, args, err := DecodeCommand(reader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cmd != "MSET" {
			t.Errorf("expected command 'MSET', got %q", cmd)
		}
		expected := []string{"key1", "val1", "key2", "val2", "key3", "val3"}
		if len(args) != len(expected) {
			t.Errorf("expected %d arguments, got %d", len(expected), len(args))
		}
		for i, expectedArg := range expected {
			if args[i] != expectedArg {
				t.Errorf("expected argument[%d] %q, got %q", i, expectedArg, args[i])
			}
		}
	})
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(substr) > 0 && strings.Contains(s, substr)))
}

func TestEncodeSimpleString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "OK response",
			input:    "OK",
			expected: "+OK\r\n",
		},
		{
			name:     "PONG response",
			input:    "PONG",
			expected: "+PONG\r\n",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "+\r\n",
		},
		{
			name:     "String with spaces",
			input:    "Hello World",
			expected: "+Hello World\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeSimpleString(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestEncodeError(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Generic error",
			input:    "ERR unknown command",
			expected: "-ERR unknown command\r\n",
		},
		{
			name:     "Syntax error",
			input:    "ERR syntax error",
			expected: "-ERR syntax error\r\n",
		},
		{
			name:     "Empty error",
			input:    "",
			expected: "-\r\n",
		},
		{
			name:     "Error with special characters",
			input:    "ERR invalid key: 'test@123'",
			expected: "-ERR invalid key: 'test@123'\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeError(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestEncodeInteger(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{
			name:     "Positive integer",
			input:    123,
			expected: ":123\r\n",
		},
		{
			name:     "Zero",
			input:    0,
			expected: ":0\r\n",
		},
		{
			name:     "Negative integer",
			input:    -456,
			expected: ":-456\r\n",
		},
		{
			name:     "Large positive integer",
			input:    9223372036854775807, // max int64
			expected: ":9223372036854775807\r\n",
		},
		{
			name:     "Large negative integer",
			input:    -9223372036854775808, // min int64
			expected: ":-9223372036854775808\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeInteger(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestEncodeBulkString(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{
			name:     "Regular string",
			input:    stringPtr("hello"),
			expected: "$5\r\nhello\r\n",
		},
		{
			name:     "Empty string",
			input:    stringPtr(""),
			expected: "$0\r\n\r\n",
		},
		{
			name:     "String with spaces",
			input:    stringPtr("hello world"),
			expected: "$11\r\nhello world\r\n",
		},
		{
			name:     "String with special characters",
			input:    stringPtr("hello\nworld\r\ntest"),
			expected: "$17\r\nhello\nworld\r\ntest\r\n",
		},
		{
			name:     "Nil string",
			input:    nil,
			expected: "$-1\r\n",
		},
		{
			name:     "String with numbers",
			input:    stringPtr("123456"),
			expected: "$6\r\n123456\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeBulkString(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestEncodeArray(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "Two element array",
			input:    []string{"hello", "world"},
			expected: "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
		},
		{
			name:     "Single element array",
			input:    []string{"test"},
			expected: "*1\r\n$4\r\ntest\r\n",
		},
		{
			name:     "Empty array",
			input:    []string{},
			expected: "*0\r\n",
		},
		{
			name:     "Array with empty string",
			input:    []string{"", "test"},
			expected: "*2\r\n$0\r\n\r\n$4\r\ntest\r\n",
		},
		{
			name:     "Nil array",
			input:    nil,
			expected: "*-1\r\n",
		},
		{
			name:     "Array with special characters",
			input:    []string{"key:123", "value@#$"},
			expected: "*2\r\n$7\r\nkey:123\r\n$8\r\nvalue@#$\r\n",
		},
		{
			name:     "Large array",
			input:    []string{"a", "b", "c", "d", "e"},
			expected: "*5\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n$1\r\nd\r\n$1\r\ne\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeArray(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestEncodeNullBulkString(t *testing.T) {
	result := EncodeNullBulkString()
	expected := "$-1\r\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestEncodeNullArray(t *testing.T) {
	result := EncodeNullArray()
	expected := "*-1\r\n"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

func TestEncodeArrayMixed(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected string
	}{
		{
			name:     "Mixed string and integer",
			input:    []interface{}{"hello", int64(42)},
			expected: "*2\r\n$5\r\nhello\r\n:42\r\n",
		},
		{
			name:     "Array with nil element",
			input:    []interface{}{"test", nil, "end"},
			expected: "*3\r\n$4\r\ntest\r\n$-1\r\n$3\r\nend\r\n",
		},
		{
			name:     "Nested array",
			input:    []interface{}{"outer", []interface{}{"inner1", "inner2"}, "last"},
			expected: "*3\r\n$5\r\nouter\r\n*2\r\n$6\r\ninner1\r\n$6\r\ninner2\r\n$4\r\nlast\r\n",
		},
		{
			name:     "Array with string slice",
			input:    []interface{}{"first", []string{"a", "b"}, "third"},
			expected: "*3\r\n$5\r\nfirst\r\n*2\r\n$1\r\na\r\n$1\r\nb\r\n$5\r\nthird\r\n",
		},
		{
			name:     "Array with string pointer",
			input:    []interface{}{stringPtr("pointer"), "regular"},
			expected: "*2\r\n$7\r\npointer\r\n$7\r\nregular\r\n",
		},
		{
			name:     "Array with nil string pointer",
			input:    []interface{}{(*string)(nil), "regular"},
			expected: "*2\r\n$-1\r\n$7\r\nregular\r\n",
		},
		{
			name:     "Array with int types",
			input:    []interface{}{int(123), int64(456)},
			expected: "*2\r\n:123\r\n:456\r\n",
		},
		{
			name:     "Empty mixed array",
			input:    []interface{}{},
			expected: "*0\r\n",
		},
		{
			name:     "Nil mixed array",
			input:    nil,
			expected: "*-1\r\n",
		},
		{
			name:     "Complex nested structure",
			input:    []interface{}{
				"level1",
				[]interface{}{
					"level2",
					[]interface{}{"level3", int64(42)},
					nil,
				},
				int64(999),
			},
			expected: "*3\r\n$6\r\nlevel1\r\n*3\r\n$6\r\nlevel2\r\n*2\r\n$6\r\nlevel3\r\n:42\r\n$-1\r\n:999\r\n",
		},
		{
			name:     "Array with unsupported type (fallback to string)",
			input:    []interface{}{"test", 3.14},
			expected: "*2\r\n$4\r\ntest\r\n$4\r\n3.14\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeArrayMixed(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestEncodeElementTypes(t *testing.T) {
	t.Run("String element", func(t *testing.T) {
		result := encodeElement("hello")
		expected := "$5\r\nhello\r\n"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("String pointer element", func(t *testing.T) {
		s := "world"
		result := encodeElement(&s)
		expected := "$5\r\nworld\r\n"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("Nil element", func(t *testing.T) {
		result := encodeElement(nil)
		expected := "$-1\r\n"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("Int64 element", func(t *testing.T) {
		result := encodeElement(int64(42))
		expected := ":42\r\n"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("Int element", func(t *testing.T) {
		result := encodeElement(123)
		expected := ":123\r\n"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("Nested array element", func(t *testing.T) {
		result := encodeElement([]interface{}{"a", "b"})
		expected := "*2\r\n$1\r\na\r\n$1\r\nb\r\n"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("String slice element", func(t *testing.T) {
		result := encodeElement([]string{"x", "y"})
		expected := "*2\r\n$1\r\nx\r\n$1\r\ny\r\n"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("Unsupported type fallback", func(t *testing.T) {
		result := encodeElement(3.14159)
		expected := "$7\r\n3.14159\r\n"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})
}

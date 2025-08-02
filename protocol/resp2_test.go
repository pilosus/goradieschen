package protocol

import (
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
			cmd, args, err := DecodeCommand(tt.input)

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
		cmd, args, err := DecodeCommand(input)
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
		cmd, args, err := DecodeCommand(input)
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
		cmd, args, err := DecodeCommand(input)
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

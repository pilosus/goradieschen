package protocol

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// DecodeCommand decodes a RESP2 command from a bufio.Reader into the command name and its arguments.
func DecodeCommand(r *bufio.Reader) (string, []string, error) {

	line, err := readLine(r)
	if err != nil {
		return "", nil, err
	}

	if !strings.HasPrefix(line, "*") {
		return "", nil, fmt.Errorf("expected array (*), got: %q", line)
	}

	count, err := strconv.Atoi(line[1:])
	if err != nil {
		return "", nil, fmt.Errorf("invalid array length: %w", err)
	}

	// TODO -1 is nil in RESP2, handle this case
	if count < 1 {
		return "", nil, errors.New("command must contain at least one element")
	}

	parts := make([]string, count)
	for i := 0; i < count; i++ {
		// Expect $<length>
		line, err := readLine(r)
		if err != nil {
			return "", nil, err
		}
		if !strings.HasPrefix(line, "$") {
			return "", nil, fmt.Errorf("expected bulk string ($), got: %q", line)
		}
		length, err := strconv.Atoi(line[1:])
		if err != nil {
			return "", nil, fmt.Errorf("invalid bulk string length: %w", err)
		}
		buf := make([]byte, length+2) // +2 for \r\n
		if _, err := io.ReadFull(r, buf); err != nil {
			return "", nil, err
		}
		parts[i] = string(buf[:length]) // drop \r\n
	}
	cmd := parts[0]
	args := parts[1:]
	return cmd, args, nil
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(line, "\r\n"), nil
}

// EncodeSimpleString encodes a simple string response (+OK\r\n)
func EncodeSimpleString(s string) string {
	return "+" + s + "\r\n"
}

// EncodeError encodes an error response (-ERR message\r\n)
func EncodeError(err string) string {
	return "-" + err + "\r\n"
}

// EncodeInteger encodes an integer response (:123\r\n)
func EncodeInteger(n int64) string {
	return ":" + strconv.FormatInt(n, 10) + "\r\n"
}

// EncodeBulkString encodes a bulk string response ($5\r\nhello\r\n)
// Returns "$-1\r\n" for nil values
func EncodeBulkString(s *string) string {
	if s == nil {
		return "$-1\r\n"
	}
	return "$" + strconv.Itoa(len(*s)) + "\r\n" + *s + "\r\n"
}

// EncodeArrayMixed encodes an array with mixed element types
// Supports: string, *string, int64, []interface{}, nil
// Returns "*-1\r\n" for nil arrays
func EncodeArrayMixed(elements []interface{}) string {
	if elements == nil {
		return "*-1\r\n"
	}
	result := "*" + strconv.Itoa(len(elements)) + "\r\n"
	for _, element := range elements {
		result += encodeElement(element)
	}
	return result
}

// encodeElement encodes a single element based on its type
func encodeElement(element interface{}) string {
	switch v := element.(type) {
	case nil:
		return EncodeNullBulkString()
	case string:
		return EncodeBulkString(&v)
	case *string:
		return EncodeBulkString(v)
	case int64:
		return EncodeInteger(v)
	case int:
		return EncodeInteger(int64(v))
	case []interface{}:
		return EncodeArrayMixed(v)
	case []string:
		// Convert []string to []interface{} for recursive handling
		converted := make([]interface{}, len(v))
		for i, s := range v {
			converted[i] = s
		}
		return EncodeArrayMixed(converted)
	default:
		// Fallback: convert to string
		str := fmt.Sprintf("%v", v)
		return EncodeBulkString(&str)
	}
}

// EncodeArray encodes an array of strings (convenience function)
// Returns "*-1\r\n" for nil arrays
func EncodeArray(elements []string) string {
	if elements == nil {
		return "*-1\r\n"
	}

	result := "*" + strconv.Itoa(len(elements)) + "\r\n"
	for _, element := range elements {
		result += EncodeBulkString(&element)
	}
	return result
}

// EncodeNullBulkString encodes a null bulk string ($-1\r\n)
func EncodeNullBulkString() string {
	return "$-1\r\n"
}

// EncodeNullArray encodes a null array (*-1\r\n)
func EncodeNullArray() string {
	return "*-1\r\n"
}

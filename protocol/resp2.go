package protocol

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// DecodeCommand decodes a RESP2 command string from a client into the command name and its arguments.
func DecodeCommand(encodedCommand string) (string, []string, error) {
	r := bufio.NewReader(strings.NewReader(encodedCommand))

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

package protocol

import (
	"bufio"
	"github.com/pilosus/goradieschen/store"
	"github.com/pilosus/goradieschen/ttlstore"
	"strconv"
	"strings"
	"time"
)

const GenericErrorPrefix = "ERR"
const ReturnOK = "OK"

func ParseCommand(reader *bufio.Reader, store *store.Store, ttl *ttlstore.TTLStore) string {
	cmd, cmdArgs, err := DecodeCommand(reader)
	if err != nil {
		return EncodeError(GenericErrorPrefix + " " + err.Error())
	}

	switch strings.ToUpper(cmd) {
	case "SET":
		if len(cmdArgs) != 2 {
			return EncodeError(GenericErrorPrefix + " usage: SET key value")
		}
		store.Set(cmdArgs[0], cmdArgs[1])
		return EncodeSimpleString(ReturnOK)
	case "GET":
		if len(cmdArgs) != 1 {
			return EncodeError(GenericErrorPrefix + " usage: GET key")
		}
		val, ok := store.Get(cmdArgs[0])
		if !ok {
			return EncodeNullBulkString()
		}
		return EncodeBulkString(&val)
	case "DEL":
		if len(cmdArgs) != 1 {
			return EncodeError(GenericErrorPrefix + " usage: DEL key")
		}
		deleted := store.Delete(cmdArgs[0])
		if deleted {
			return EncodeSimpleString(ReturnOK)
		}
		return EncodeNullBulkString()
	case "KEYS":
		if len(cmdArgs) != 1 {
			return EncodeError(GenericErrorPrefix + " usage: KEYS pattern")
		}
		val, ok := store.Match(cmdArgs[0])
		if !ok {
			return EncodeNullBulkString()
		}
		return EncodeArray(val)
	case "EXPIRE":
		if len(cmdArgs) != 2 {
			return EncodeError(GenericErrorPrefix + " usage: EXPIRE key seconds")
		}
		seconds, err := strconv.Atoi(cmdArgs[1])
		if err != nil || seconds < 0 {
			return EncodeError(GenericErrorPrefix + " invalid seconds value: " + cmdArgs[1])
		}
		_, ok := store.Get(cmdArgs[0])
		// If the key does not exist, no need to set TTL
		if !ok {
			return EncodeInteger(0)
		}
		expiresAt := time.Now().Add(time.Duration(seconds) * time.Second)
		ttl.SetTTL(cmdArgs[0], expiresAt)
		return EncodeInteger(1)
	case "TTL":
		if len(cmdArgs) != 1 {
			return EncodeError(GenericErrorPrefix + " usage: TTL key")
		}
		_, ok := store.Get(cmdArgs[0])
		if !ok {
			return EncodeInteger(-2) // Key does not exist
		}
		expiresAt, ok := ttl.GetTTL(cmdArgs[0])
		if !ok {
			return EncodeInteger(-1) // Key exists but has no TTL set
		}
		remaining := expiresAt.Sub(time.Now()).Seconds()
		if remaining < 0 {
			return EncodeInteger(0) // Key has expired
		}
		return EncodeInteger(int64(remaining))
	case "FLUSHALL":
		if len(cmdArgs) != 0 {
			return EncodeError(GenericErrorPrefix + " usage: FLUSHALL")
		}
		store.FlushAll()
		ttl.FlushAll()
		return EncodeSimpleString(ReturnOK)
	case "PING":
		return "PONG"
	case "COMMAND":
		if len(cmdArgs) != 0 {
			return EncodeError(GenericErrorPrefix + " usage: COMMAND")
		}
		commands := []interface{}{
			[]interface{}{"SET", int64(3), []interface{}{"write"}, int64(1), int64(1), int64(1)},
			[]interface{}{"GET", int64(2), []interface{}{"readonly"}, int64(1), int64(1), int64(1)},
			[]interface{}{"DEL", int64(2), []interface{}{"write"}, int64(1), int64(1), int64(1)},
			[]interface{}{"KEYS", int64(2), []interface{}{"readonly"}, int64(1), int64(1), int64(1)},
			[]interface{}{"EXPIRE", int64(3), []interface{}{"write"}, int64(1), int64(1), int64(1)},
			[]interface{}{"TTL", int64(2), []interface{}{"readonly"}, int64(1), int64(1), int64(1)},
			[]interface{}{"FLUSHALL", int64(1), []interface{}{"write"}, int64(0), int64(0), int64(0)},
			[]interface{}{"PING", int64(1), []interface{}{"stale", "fast"}, int64(0), int64(0), int64(0)},
			[]interface{}{"COMMAND", int64(1), []interface{}{"readonly"}, int64(0), int64(0), int64(0)},
		}
		return EncodeArrayMixed(commands)
	default:
		return EncodeError(GenericErrorPrefix + " unknown command: " + cmd)
	}
}

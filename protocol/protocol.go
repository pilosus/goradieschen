package protocol

import (
	"github.com/pilosus/goradieschen/store"
	"github.com/pilosus/goradieschen/ttlstore"
	"strconv"
	"strings"
	"time"
)

const GenericErrorPrefix = "ERR"

const ReturnOK = "OK"
const ReturnNil = "nil"

func ParseCommand(command string, store *store.Store, ttl *ttlstore.TTLStore) string {
	parts := strings.Fields(command)

	if len(parts) == 0 {
		return GenericErrorPrefix + " empty command"
	}

	switch strings.ToUpper(parts[0]) {
	case "SET":
		if len(parts) != 3 {
			return GenericErrorPrefix + " usage: SET key value"
		}
		store.Set(parts[1], parts[2])
		return ReturnOK
	case "GET":
		if len(parts) != 2 {
			return GenericErrorPrefix + " usage: GET key"
		}
		val, ok := store.Get(parts[1])
		if !ok {
			return ReturnNil
		}
		return val
	case "DEL":
		if len(parts) != 2 {
			return GenericErrorPrefix + " usage: DEL key"
		}
		deleted := store.Delete(parts[1])
		if deleted {
			return ReturnOK
		}
		return ReturnNil
	case "KEYS":
		if len(parts) != 2 {
			return GenericErrorPrefix + " usage: KEYS pattern"
		}
		val, ok := store.Match(parts[1])
		if !ok {
			return ReturnNil
		}
		return val
	case "EXPIRE":
		if len(parts) != 3 {
			return GenericErrorPrefix + " usage: EXPIRE key seconds"
		}
		seconds, err := strconv.Atoi(parts[2])
		if err != nil || seconds < 0 {
			return GenericErrorPrefix + " invalid seconds value: " + parts[2]
		}
		_, ok := store.Get(parts[1])
		// If the key does not exist, no need to set TTL
		if !ok {
			return "0"
		}
		expiresAt := time.Now().Add(time.Duration(seconds) * time.Second)
		ttl.SetTTL(parts[1], expiresAt)
		return "1"
	case "TTL":
		if len(parts) != 2 {
			return GenericErrorPrefix + " usage: TTL key"
		}
		_, ok := store.Get(parts[1])
		if !ok {
			return "-2" // Key does not exist
		}
		expiresAt, ok := ttl.GetTTL(parts[1])
		if !ok {
			return "-1" // Key exists but has no TTL set
		}
		remaining := expiresAt.Sub(time.Now()).Seconds()
		if remaining < 0 {
			return "0" // Key has expired
		}
		return strconv.FormatFloat(remaining, 'f', 0, 64) // Return remaining seconds as integer
	case "FLUSHALL":
		if len(parts) != 1 {
			return GenericErrorPrefix + " usage: FLUSHALL"
		}
		store.FlushAll()
		ttl.FlushAll()
		return ReturnOK
	case "PING":
		return "PONG"
	default:
		return GenericErrorPrefix + " unknown command: " + parts[0]
	}
}

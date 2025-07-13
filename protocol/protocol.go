package protocol

import (
	"github.com/pilosus/goradieschen/store"
	"strings"
)

const GenericErrorPrefix = "ERR"

const ReturnOK = "OK"
const ReturnNil = "nil"

func ParseCommand(command string, store *store.Store) string {
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
	default:
		return GenericErrorPrefix + " unknown command: " + parts[0]
	}
}

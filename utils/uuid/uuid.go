package uuid

import "time"

func GenerateUUID() uint64 {
	return uint64(time.Now().UnixNano())
}

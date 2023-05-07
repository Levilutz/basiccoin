package p2p

import (
	"sync"

	"github.com/google/uuid"
)

var runtimeID *uuid.UUID
var mu sync.Mutex

func GetRuntimeID() string {
	// Get the ID uniquely identifying this runtime
	mu.Lock()
	defer mu.Unlock()
	if runtimeID == nil {
		rid := uuid.New()
		runtimeID = &rid
	}
	return runtimeID.String()
}

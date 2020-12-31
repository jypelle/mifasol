package tool

import (
	"github.com/oklog/ulid/v2"
	"math/rand"
	"sync"
	"time"
)

var entropy = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
var entropyMutex sync.Mutex

func CreateUlid() string {
	entropyMutex.Lock()
	uLID := ulid.MustNew(ulid.Timestamp(time.Now().UTC()), entropy)
	entropyMutex.Unlock()

	return uLID.String()
}

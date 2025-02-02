package testtooling

import (
	"sync"

	"github.com/brianvoe/gofakeit/v6"
)

var (
	faker      = gofakeit.New(0)
	fakerMutex = &sync.Mutex{}
)

func FakeUUID() string {
	fakerMutex.Lock()
	defer fakerMutex.Unlock()
	return faker.UUID()
}

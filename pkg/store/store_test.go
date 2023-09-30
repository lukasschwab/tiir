package store

import (
	"fmt"
	"io"
	"log"
	"testing"

	"github.com/sirupsen/logrus"
)

// TODO: move to test_store.go
var AllStoreGenerators = map[string]storeBuilder{
	"file":  BuildFile,
	"mysql": BuildMySQL,
}

func BenchmarkStore(b *testing.B) {
	log.SetOutput(io.Discard)
	logrus.SetOutput(io.Discard)

	initialLoads := []int{0, 10, 100, 1_000}
	for name, generator := range AllStoreGenerators {
		for _, load := range initialLoads {
			b.Run(fmt.Sprintf("%s:%d", name, load), func(b *testing.B) {
				s := generator(b)
				// Initial load.
				b.Logf("Initializing load")
				for i := 0; i < load; i++ {
					s.Upsert(randomText(b))
				}
				b.Logf("Benching")
				b.ResetTimer()
				for n := 0; n < b.N; n++ {
					b.StopTimer()
					marginalText := randomText(b)
					b.StartTimer()
					s.Upsert(marginalText)
				}

				b.StopTimer()
				s.Close()
				b.StartTimer()
			})
		}
	}
}

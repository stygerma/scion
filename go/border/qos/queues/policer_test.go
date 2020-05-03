package queues

import (
	"fmt"
	"testing"
	"time"
)

// As these tests are based on real time they might need some big limit
// to succeed. Usually it's rate / 1000 * 2 seconds.
func TestBasic(t *testing.T) {

	tb := TokenBucket{}

	var tests = []struct {
		name          string
		rate          int
		takeRate      int
		rateLimit     time.Duration
		shouldSucceed bool
	}{
		{"Basic Test", 1000, 1, time.Millisecond, true},
		{"TB take rate is double", 1000, 2, time.Millisecond, false},
		{"TB rate off by one", 999, 1, time.Millisecond, false},
		// {"What should succeed", 6250000, 6250, time.Millisecond, true},
		// {"What should succeed", 6250000, 62500, time.Millisecond, false},
		// {"What should succeed", 6250000, 31250, time.Millisecond, false},
		// {"What should succeed", 6250000, 15625, time.Millisecond, false},
		// {"What should succeed", 6250000, 7812, time.Millisecond, false},
	}

	for _, tt := range tests {
		tb.Init(tt.rate)
		rateLimit := tt.rateLimit
		testEnd := time.Second * 2
		totalTake := 0
		allowedTake := 0
		succ := true

		testEndTicker := time.Tick(testEnd)

		ticker := time.NewTicker(rateLimit)
		defer ticker.Stop()
		for range ticker.C {
			totalTake += tt.takeRate
			if !tb.Take(tt.takeRate) {
				succ = false
			} else {
				allowedTake += tt.takeRate
			}
			select {
			case <-testEndTicker:
				if succ != tt.shouldSucceed {
					fmt.Println("Tried to take", totalTake, "was allowed", allowedTake)
					t.Errorf("Test %s has failed", tt.name)
				}
			default:
				continue
			}
			break
		}
	}
}

func BenchmarkRefill(b *testing.B) {

	initialVal := 100000000000

	tb := TokenBucket{}
	tb.Init(initialVal)
	tb.ForceTake(initialVal)

	for n := 0; n < b.N; n++ {
		tb.refill()
		tb.ForceTake(10000)
	}
}

func BenchmarkAvailable(b *testing.B) {

	tb := TokenBucket{}
	tb.Init(10000)

	for n := 0; n < b.N; n++ {
		tb.Available(100)
	}
}

func BenchmarkTake(b *testing.B) {

	tb := TokenBucket{}
	tb.Init(10000)

	benches := []struct {
		name    string
		takeVal int
	}{
		{"1", 1},
		{"10", 10},
		{"100", 100},
		{"100", 1000},
		{"10000", 10000},
	}

	for _, bench := range benches {
		b.Run(
			bench.name,
			func(b *testing.B) {
				for n := 0; n < b.N; n++ {
					tb.Take(bench.takeVal)
				}
			},
		)

	}

}

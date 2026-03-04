package simulation

import (
	"math/rand/v2"
	"time"
)

// AuthDelay returns a realistic delay (50-200ms) to simulate auth processing.
func AuthDelay() time.Duration {
	return time.Duration(50+rand.IntN(151)) * time.Millisecond
}

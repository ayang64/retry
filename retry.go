package retry

import (
	"context"
	"iter"
	"math/rand"
	"time"
)

type Backoff interface {
	Delay(int) time.Duration
}

// Decay implements exponential decay.  I is the initial time and H is the
// iteration that the half-life falls on.
type Decay struct {
	I time.Duration // initial decay
	H int           // iteration at which the decay's half-life will be reached
}

func (d *Decay) Delay(n int) time.Duration {
	return d.I >> (n / d.H)
}

// Jitter is a backoff strategy the augments other staegies by adding random
// jitter to the Delay() result.
//
// The formula for jitter applied is ±(j/2) where j is a random number between
// 0 and J.
type Jitter struct {
	J time.Duration // Amount of jitter to apply.
	B Backoff       // Backoff strategy to apply jitter to.
}

func (j *Jitter) Delay(n int) time.Duration {
	return j.B.Delay(n) + time.Duration(rand.Int63n(int64(j.J))) - j.J/2
}

// Exponential encodes the amount to back off exponentially.  Its value is
// doubled every iteration.
type Exponential time.Duration

func (e Exponential) Delay(n int) time.Duration {
	return time.Duration(e) * 1 << n
}

// Constant applies a single constant delay to every iteration.
type Constant time.Duration

func (c Constant) Delay(n int) time.Duration {
	return time.Duration(c)
}

// Linear increases the delay by a fixed amount every iteration.
type Linear time.Duration

func (l Linear) Delay(n int) time.Duration {
	return time.Duration(l) * time.Duration(n+1)
}

// Attempt returns an iterator over retry attempts using the provided Backoff
// strategy.  Each iteration yields the attempt index and the delay duration
// before the next attempt.
//
// The caller is responsible for executing the operation, handling errors, and
// deciding when to stop based on context cancellation, attempt count, or delay
// size.
//
// Example:
//
//	for i, delay := range retry.Attempt(ctx, retry.Linear(100*time.Millisecond)) {
//	    if err := doSomething(); err == nil {
//	        break
//	    }
//	    if delay > 2*time.Second {
//	        break // give up if delay exceeds threshold
//	    }
//	}
func Attempt(ctx context.Context, b Backoff) iter.Seq2[int, time.Duration] {
	return func(yield func(int, time.Duration) bool) {
		for i := 0; ; i++ {
			if ctx.Err() != nil {
				return
			}
			d := b.Delay(i)
			if !yield(i, d) {
				return
			}

			if d == 0 {
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-time.Tick(d):
			}
		}
	}
}

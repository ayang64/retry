package retry

import (
	"context"
	"testing"
	"time"
)

func TestAttempt(t *testing.T) {
	for i, d := range Attempt(context.TODO(), &Decay{I: 500 * time.Millisecond, H: 7}) {
		t.Logf("%d, %s: hello!", i, d)
	}

	for i, d := range Attempt(context.TODO(), Exponential(time.Second*3)) {
		if d > time.Second*25 {
			break
		}
		t.Logf("%d, %s: hello!", i, d)
	}

	for i := range Attempt(context.TODO(), Constant(time.Second)) {
		if i > 9 {
			break
		}
		t.Logf("%d: hello!", i)
	}

	j := Jitter{
		J: time.Duration(500 * time.Millisecond),
		B: Constant(1 * time.Second),
	}

	for i, d := range Attempt(context.TODO(), &j) {
		if i > 9 {
			break
		}
		t.Logf("%d, %s: with jitter!", i, d)
	}
}

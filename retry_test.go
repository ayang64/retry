package retry

import (
	"context"
	"testing"
	"time"
)

func eq[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range len(a) {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestConstant(t *testing.T) {
	tests := map[string]struct {
		i        int
		backoff  Constant
		expected []time.Duration
	}{
		"0 delay": {
			i:        5,
			backoff:  Constant(0),
			expected: []time.Duration{0, 0, 0, 0, 0},
		},
		"10 second": {
			i:        5,
			backoff:  Constant(10 * time.Second),
			expected: []time.Duration{10 * time.Second, 10 * time.Second, 10 * time.Second, 10 * time.Second, 10 * time.Second},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var got []time.Duration
			for i := range test.i {
				got = append(got, test.backoff.Delay(i))
			}

			if !eq(got, test.expected) {
				t.Fatalf("got intervals %#v; expected %#v", got, test.expected)
			}
		})
	}
}

func TestLinear(t *testing.T) {
	tests := map[string]struct {
		i        int
		backoff  Linear
		expected []time.Duration
	}{
		"0 delay": {
			i:        5,
			backoff:  Linear(0),
			expected: []time.Duration{0, 0, 0, 0, 0},
		},
		"10 second": {
			i:        5,
			backoff:  Linear(10 * time.Second),
			expected: []time.Duration{10 * time.Second, 20 * time.Second, 30 * time.Second, 40 * time.Second, 50 * time.Second},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var got []time.Duration
			for i := range test.i {
				got = append(got, test.backoff.Delay(i))
			}

			if !eq(got, test.expected) {
				t.Fatalf("got intervals %#v; expected %#v", got, test.expected)
			}
		})
	}
}

func TestExponential(t *testing.T) {
	tests := map[string]struct {
		i        int
		backoff  Exponential
		expected []time.Duration
	}{
		"0 delay": {
			i:        5,
			backoff:  Exponential(0),
			expected: []time.Duration{0, 0, 0, 0, 0},
		},
		"10 second": {
			i:        5,
			backoff:  Exponential(10 * time.Second),
			expected: []time.Duration{10 * time.Second, 20 * time.Second, 40 * time.Second, 80 * time.Second, 160 * time.Second},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var got []time.Duration
			for i := range test.i {
				got = append(got, test.backoff.Delay(i))
			}

			if !eq(got, test.expected) {
				t.Fatalf("got intervals %#v; expected %#v", got, test.expected)
			}
		})
	}
}

func TestDecay(t *testing.T) {
	tests := map[string]struct {
		i        int
		backoff  Decay
		expected []time.Duration
	}{
		"0 delay": {
			i:        10,
			backoff:  Decay{I: 0, H: 3},
			expected: []time.Duration{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		"10 second": {
			i:        10,
			backoff:  Decay{I: time.Second * 10, H: 2},
			expected: []time.Duration{10 * time.Second, 10 * time.Second, 5 * time.Second, 5 * time.Second, 2500 * time.Millisecond, 2500 * time.Millisecond, 1250 * time.Millisecond, 1250 * time.Millisecond, 625 * time.Millisecond, 625 * time.Millisecond},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var got []time.Duration
			for i := range test.i {
				got = append(got, test.backoff.Delay(i))
			}

			if !eq(got, test.expected) {
				t.Fatalf("got intervals %#v; expected %#v", got, test.expected)
			}
		})
	}
}

func TestAttempt(t *testing.T) {
	t.Run("constant time, 10 iterations", func(t *testing.T) {
		backoff := Constant(10 * time.Microsecond)
		var interval []time.Duration
		for i, d := range Attempt(t.Context(), backoff) {
			interval = append(interval, d)
			if i == 9 {
				break
			}
		}
		if got, expected := len(interval), 10; got != expected {
			t.Fatalf("got %d results; expected %d", got, expected)
		}
	})

	// a 0 delay causes loop to only execute once
	t.Run("0 delay", func(t *testing.T) {
		backoff := Constant(0)
		var interval []time.Duration
		for i, d := range Attempt(t.Context(), backoff) {
			interval = append(interval, d)
			if i == 9 {
				break
			}
		}
		if got, expected := len(interval), 1; got != expected {
			t.Fatalf("got %d results; expected %d", got, expected)
		}
	})

	t.Run("cancelled context", func(t *testing.T) {
		backoff := Constant(2 * time.Second)
		ctx, stop := context.WithTimeout(t.Context(), 2*time.Second)
		defer stop()
		var interval []time.Duration
		for i, d := range Attempt(ctx, backoff) {
			interval = append(interval, d)
			if i == 9 {
				break
			}
		}
		if got, expected := len(interval), 2; got != expected {
			t.Fatalf("got %d results; expected %d", got, expected)
		}
	})
}

# Retry - Iterator-Based Retry Logic

This package provides retry behavior via **composable, transparent iterators**.

Now that iterators are availalable in Go, we can use them to provide a flexible
stream of `(attempt, delay)` pairs. The `delay` `time.Duraiton` value is
yielded to the body of the loop and all control over how retries are managed is
done using built-in operators and statements like continue, break, return,
panic etc.

## Why?

- Transparent: nothing is hidden — you run the loop.
- Composable: use any logic to decide when to stop.
- Context-aware: cancellation is built in.
- Testable: easy to verify retry schedules without sleeping.

## Core Concepts

### `Backoff`
A `Backoff` strategy calculates the delay for each retry attempt:
```go
type Backoff interface {
    Delay(attempt int) time.Duration
}
```

### Built-in Strategies
- `Constant`: always returns the same delay
- `Linear`: grows linearly with attempts (delay * attempt)
- `Exponential`: doubles delay each time
- `Jitter`: adds randomness to any strategy

### Retry Loop
Use `retry.Attempt` to generate `(attempt, delay)` values:

```go
for i, d := range retry.Attempt(ctx, backoff) {
    // stop retries if we've exceeded 5 iterations or 2 seconds of delay
    if i >= 5 || d > 2*time.Second {
        log.Println("giving up")
        break
    }

    err, someVal := doSomething()

    // on error, we can simply continue and most likely log the error
    if err != nil {
        log.Printf("failed on attempt %d with %v", i, err)
        continue
    }

    // at this point, we're within our retry conditions (number of iterations
    // and delay) and // no error has occured so we can return our value, thus
    // exiting the retry loop.
    return someVal, nil
}
```

## Example: Exponential Backoff with Jitter
```go
ctx := context.Background()

backoff := &retry.Jitter{
    J: 100 * time.Millisecond,
    B: retry.Exponential(time.Second * 1),
}

for i, delay := range retry.Attempt(ctx, backoff) {
    // stop retries if we've exceeded 2 seconds of delay
    if delay > 2*time.Second {
        break
    }

    // on error, we can simply continue and most likely log the error
    if err := makeRequest(); err != nil {
        continue
    }

    return
}
```

## No Opinionated Wrappers

There’s no `Do()` function that wraps your retry logic — this package gives you
just the iterator. This is by design.

You decide:
- When to stop
- What to retry
- How to handle each error

## License
BSD

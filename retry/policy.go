package retry

import (
	"context"
	"math"
	"time"
)

// Check is a func that defines a policy for processing retries based on
// a given error and an optional result associated with it.
// Returns true to continue retrying.
type Check func(err error, res interface{}) (bool, error)

// Backoff is a func that defines the amount of time to wait beatwen retries.
type Backoff func(min, max time.Duration, attemptNum int) time.Duration

// Callback accepts the boolean 'retrying' variable for the optional logical
// extension on the callback side. The 'res' return value is an optional
// untyped result for the check verification that can be useful in custom
// defined policy, and is also an actual error of the main action.
type Callback func(retrying bool) (res interface{}, err error)

// Policy is a main struct that defines retrial options and policy.
type Policy struct {
	WaitMin  time.Duration
	WaitMax  time.Duration
	RetryMax int
	Check    Check
	Backoff  Backoff
}

// DefaultPolicy is default retry policy configuration.
func DefaultPolicy() Policy {
	const (
		defWaitMin  = 1 * time.Second
		defWaitMax  = 30 * time.Second
		defRetryMax = 4
	)

	return Policy{
		WaitMin:  defWaitMin,
		WaitMax:  defWaitMax,
		RetryMax: defRetryMax,
	}
}

// Do executes the provided callback with the given retry policy. It will
// stop retrying when the context is canceled, validation completed or the
// number of attempts exceeded. The first invoke will always be performed.
func (p Policy) Do(ctx context.Context, cb Callback) (count int, err error) {
	var (
		res        interface{}
		attemptNum int
	)

	for retrying := false; ; attemptNum++ {
		res, err = cb(retrying)

		con, checkErr := p.check(err, res)
		if checkErr != nil {
			return attemptNum, checkErr
		}

		if retrying = p.RetryMax > attemptNum; !con || !retrying {
			break
		}

		wait := p.backoff(p.WaitMin, p.WaitMax, attemptNum)

		select {
		case <-ctx.Done():
			return attemptNum, ctx.Err()
		case <-time.After(wait):
		}
	}

	return attemptNum, err
}

func (p Policy) check(err error, res interface{}) (bool, error) {
	if p.Check == nil {
		return err != nil, nil
	}

	return p.Check(err, res)
}

func (p Policy) backoff(min, max time.Duration, attemptNum int) time.Duration {
	if p.Backoff == nil {
		return ExponentialBackoff(min, max, attemptNum)
	}

	return p.Backoff(min, max, attemptNum)
}

// ExponentialBackoff is a callback for exponential backoff based on
// the attempt number and limited by the provided min and max durations.
func ExponentialBackoff(min, max time.Duration, attemptNum int) time.Duration {
	const base = 2

	exp := math.Pow(base, float64(attemptNum)) * float64(min)
	dur := time.Duration(exp)

	if float64(dur) != exp || dur > max {
		return max
	}

	return dur
}

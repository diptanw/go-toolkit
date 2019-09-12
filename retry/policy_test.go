package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ExamplePolicy_Do() {
	policy := Policy{
		RetryMax: 3,
	}

	n, _ := policy.Do(context.TODO(), func(retrying bool) (interface{}, error) {
		if !retrying {
			retErr := errors.New("error")
			fmt.Printf("cb: %s\n", retErr)
			return nil, retErr
		}

		fmt.Println("cb: recovered!")
		return nil, nil
	})

	fmt.Printf("retried: %d", n)

	// Output:
	// cb: error
	// cb: recovered!
	// retried: 1
}

func TestDefaultPolicy(t *testing.T) {
	p := DefaultPolicy()

	assert.Equal(t, p.WaitMin, time.Second)
	assert.Equal(t, p.WaitMax, 30*time.Second)
	assert.Equal(t, p.RetryMax, 4)
}

func TestPolicy_Do_Attempts(t *testing.T) {
	tests := map[string]struct {
		giveCheck Check
		giveMax   int
		wantNum   int
		wantErr   error
	}{
		"success": {
			nil,
			10,
			0,
			nil,
		},
		"exceeded": {
			func(err error, res interface{}) (bool, error) {
				return true, nil
			},
			10,
			10,
			nil,
		},
		"check error": {
			func(err error, res interface{}) (bool, error) {
				return false, assert.AnError
			},
			10,
			0,
			assert.AnError,
		},
	}

	for name, test := range tests {
		tc := test

		t.Run(name, func(t *testing.T) {
			p := Policy{
				RetryMax: tc.giveMax,
				Check:    tc.giveCheck,
			}

			num, err := p.Do(context.TODO(), func(retrying bool) (res interface{}, err error) {
				return nil, nil
			})

			assert.Equal(t, tc.wantNum, num)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestPolicy_Do_ContextCancelled(t *testing.T) {
	p := Policy{
		RetryMax: 10,
		WaitMin:  time.Minute,
		WaitMax:  time.Minute,
	}

	ctx, cancel := context.WithCancel(context.TODO())
	cancel()

	_, err := p.Do(ctx, func(retrying bool) (res interface{}, err error) {
		return nil, assert.AnError
	})

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestPolicy_Do_WaitMin(t *testing.T) {
	const wait = 10 * time.Millisecond

	p := Policy{
		WaitMin:  wait,
		WaitMax:  wait,
		RetryMax: 2,
		Backoff: func(min, max time.Duration, attemptNum int) time.Duration {
			return wait
		},
	}

	now := time.Now()

	num, err := p.Do(context.TODO(), func(retrying bool) (res interface{}, err error) {
		return nil, assert.AnError
	})

	assert.GreaterOrEqual(t, int(time.Since(now)), p.RetryMax*int(p.WaitMin))
	assert.Error(t, err)
	assert.Equal(t, p.RetryMax, num)
}

func TestPolicy_Do_Retrying(t *testing.T) {
	p := Policy{
		RetryMax: 1,
	}

	var count int

	_, err := p.Do(context.TODO(), func(retrying bool) (res interface{}, err error) {
		require.Equal(t, count > 0, retrying)
		count++
		return nil, assert.AnError
	})

	assert.Equal(t, 2, count)
	assert.Error(t, err)
}

func TestPolicy_Check(t *testing.T) {
	tests := map[string]struct {
		giveCheck Check
		wantRes   bool
		wantErr   error
	}{
		"default if not set": {
			nil,
			true,
			nil,
		},
		"check is set": {
			func(err error, res interface{}) (bool, error) {
				return false, nil
			},
			false,
			nil,
		},
		"check error": {
			func(err error, res interface{}) (bool, error) {
				return false, assert.AnError
			},
			false,
			assert.AnError,
		},
	}

	for name, test := range tests {
		tc := test

		t.Run(name, func(t *testing.T) {
			p := Policy{
				Check: tc.giveCheck,
			}

			res, err := p.check(assert.AnError, nil)

			assert.Equal(t, tc.wantRes, res)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestPolicy_Backoff(t *testing.T) {
	tests := map[string]struct {
		giveBackoff Backoff
		wantDur     time.Duration
	}{
		"default exponential if not set": {
			nil,
			time.Second,
		},
		"backoff set with a fixed duration": {
			func(min, max time.Duration, attemptNum int) time.Duration {
				return 2 * time.Second
			},
			2 * time.Second,
		},
	}

	for name, test := range tests {
		tc := test

		t.Run(name, func(t *testing.T) {
			p := Policy{
				Backoff: tc.giveBackoff,
			}

			dur := p.backoff(time.Second, time.Minute, 0)
			assert.Equal(t, tc.wantDur, dur)
		})
	}
}

func TestExponentialBackoff(t *testing.T) {
	tests := map[string]struct {
		giveMin time.Duration
		giveMax time.Duration
		giveNum int
		wantDur time.Duration
	}{
		"0": {
			time.Second,
			5 * time.Minute,
			0,
			time.Second,
		},
		"1": {
			time.Second,
			5 * time.Minute,
			1,
			2 * time.Second,
		},
		"2": {
			time.Second,
			5 * time.Minute,
			2,
			4 * time.Second,
		},
		"8": {
			time.Second,
			5 * time.Minute,
			8,
			256 * time.Second,
		},
		"32": {
			time.Second,
			5 * time.Minute,
			32,
			5 * time.Minute,
		},
		"64": {
			time.Second,
			5 * time.Minute,
			64,
			5 * time.Minute,
		},
	}

	for name, test := range tests {
		tc := test

		t.Run(name, func(t *testing.T) {
			dur := ExponentialBackoff(tc.giveMin, tc.giveMax, tc.giveNum)
			assert.Equal(t, tc.wantDur, dur)
		})
	}
}

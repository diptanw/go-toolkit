package message

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewConsumer(t *testing.T) {
	rc := NewConsumer()
	assert.NotNil(t, rc)
}

func TestSubscription_Close(t *testing.T) {
	var canceled bool

	s := Subscription{
		errCh: make(chan error),
		cancel: func() {
			canceled = true
		},
	}

	s.Close()

	assert.True(t, canceled)
}

func TestSubscription_Errors(t *testing.T) {
	errCh := make(chan error)

	s := Subscription{
		errCh: errCh,
	}

	ch := s.Errors()

	assert.EqualValues(t, errCh, ch)
}

func TestConsumer_Subscribe(t *testing.T) {
	c := NewConsumer()

	tests := map[string]struct {
		giveCtx      context.Context
		giveReceiver Receiver
		giveMsgCh    chan Message
		wantError    string
	}{
		"ctx is missing": {
			nil,
			&fakeTopicReceiver{},
			make(chan Message),
			"ctx must not be nil",
		},
		"rc is missing": {
			context.TODO(),
			nil,
			make(chan Message),
			"rc must not be nil",
		},
		"msgCh is missing": {
			context.TODO(),
			&fakeTopicReceiver{},
			nil,
			"msgCh must not be nil",
		},
		"success": {
			context.TODO(),
			&fakeTopicReceiver{},
			make(chan Message),
			"",
		},
	}

	for name, test := range tests {
		tc := test

		t.Run(name, func(t *testing.T) {
			sub, err := c.Subscribe(tc.giveCtx, tc.giveReceiver, tc.giveMsgCh)

			if tc.wantError != "" {
				assert.EqualError(t, err, tc.wantError)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, c.subs)
				assert.NotNil(t, sub.cancel)
				assert.NotNil(t, sub.errCh)
			}
		})
	}
}

func TestConsumer_Subscribe_receivesMsg(t *testing.T) {
	c := Consumer{}
	ch := make(chan Message)
	rc := fakeTopicReceiver{
		retMsg: []Message{
			{
				Data: []byte{},
			},
		},
	}

	c.Subscribe(context.TODO(), &rc, ch)

	for m := range ch {
		assert.Equal(t, rc.retMsg[0], m)
		break
	}
}

func TestConsumer_Subscribe_receivesErr(t *testing.T) {
	c := Consumer{}
	rc := fakeTopicReceiver{
		retErr: assert.AnError,
	}

	sub, _ := c.Subscribe(context.TODO(), &rc, make(chan Message))
	err := <-sub.Errors()

	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

func TestConsumer_Subscribe_subCancelled(t *testing.T) {
	c := Consumer{}
	sub, _ := c.Subscribe(context.TODO(), &fakeTopicReceiver{}, make(chan Message))

	wg := sync.WaitGroup{}

	wg.Add(1)

	go func() {
		wg.Done()

		err := <-sub.Errors()

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	}()

	wg.Add(1)

	go func() {
		wg.Done()
		sub.Close()
	}()

	wg.Wait()
}

func TestConsumer_Close(t *testing.T) {
	c := Consumer{}
	sub, _ := c.Subscribe(context.TODO(), &fakeTopicReceiver{}, make(chan Message))

	wg := sync.WaitGroup{}

	wg.Add(1)

	go func() {
		wg.Done()

		err := <-sub.Errors()

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	}()

	wg.Add(1)

	go func() {
		wg.Done()
		c.Close()
	}()
}

type fakeTopicReceiver struct {
	retMsg []Message
	retErr error
}

func (rc *fakeTopicReceiver) Receive(ctx context.Context) ([]Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		if rc.retMsg == nil && rc.retErr == nil {
			time.Sleep(time.Minute)
		}

		return rc.retMsg, rc.retErr
	}
}

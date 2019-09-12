package message

import (
	"context"
	"errors"
)

// Message defines the structure of a message to be received/published.
type Message struct {
	// ID is unique message identifier.
	ID string
	// Data is the actual event data received or to be sent.
	Data []byte

	// Passing false indicates processing failed.
	// A non-nil return indicates that the peer has not accepted
	// the acknowledge event. Can be cancelled by context.
	Ack func(ctx context.Context, ok bool) error
}

// Receiver in an interface that defines the topic consumer methods.
type Receiver interface {
	Receive(ctx context.Context) ([]Message, error)
}

// Consumer is wrapper for the message transport abstraction.
type Consumer struct {
	subs []Subscription
}

// NewConsumer creates new instance of Consumer struct.
func NewConsumer() *Consumer {
	return &Consumer{}
}

// Subscribe subscribes the handler for a given topic.
// Returns a subscription for further cancellation and error handling.
func (c *Consumer) Subscribe(ctx context.Context, rc Receiver, msgCh chan<- Message) (Subscription, error) {
	if ctx == nil {
		return Subscription{}, errors.New("ctx must not be nil")
	}

	if rc == nil {
		return Subscription{}, errors.New("rc must not be nil")
	}

	if msgCh == nil {
		return Subscription{}, errors.New("msgCh must not be nil")
	}

	ctx, cancel := context.WithCancel(ctx)
	errCh := connect(ctx, rc, msgCh)

	sub := Subscription{
		errCh:  errCh,
		cancel: cancel,
	}

	c.subs = append(c.subs, sub)

	return sub, nil
}

func connect(ctx context.Context, rc Receiver, msgCh chan<- Message) chan error {
	errCh := make(chan error)

	go func() {
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
				ms, err := rc.Receive(ctx)
				if err != nil {
					errCh <- err
					return
				}

				for _, m := range ms {
					msgCh <- m
				}
			}
		}
	}()

	return errCh
}

// Close cancels all active subscriptions.
func (c *Consumer) Close() {
	for _, s := range c.subs {
		s.Close()
	}

	c.subs = nil
}

// Subscription is a struct with the registration data.
type Subscription struct {
	errCh  chan error
	cancel context.CancelFunc
}

// Close cancels the subscription by invoking cancelation callback.
func (s Subscription) Close() {
	if s.cancel != nil {
		s.cancel()
	}
}

// Errors returns error channel for subscription error handling.
func (s Subscription) Errors() <-chan error {
	return s.errCh
}

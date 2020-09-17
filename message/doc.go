/*
Package message provides types for asynchronous event consumption
based on Go channels. Topic receiver is pluggable and it abstracts the
technical integrity part.

SQS topic receiver currently support receiving SQS messages performing
a long polling request to the message queue.

Basic Usage

	sqs := sqs.New(session.New())
	url, _ := awssqs.GetQueueURL(sqs, "sqs-test")
	rc := awssqs.NewSQSReceiver(sqs, url)

	msgCh := make(chan message.Message)

	c := message.NewConsumer()

	c.Subscribe(context.Background(), rc, msgCh)

	p := worker.NewPool(10)

	for m := range msgCh {
		mc := m

		// Distribute messages to async workers
		p.Enqueue(func(_ context.Context) {
			// Do something with mc.Data
			mc.Ack(context.Background(), true)
		})
	}
*/
package message

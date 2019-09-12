/*
Package retry provides types for resilient retry policy, as well as
HTTP RoundTripper for the common status/error responses

Policy types can only be used at the application level and should be
loosely coupled with other shared modules that depend on it. The
extracted abstraction for the Policy.Do method and Callback is the only
behavior needed for the most common use cases.

Basic Usage

	policy := retry.Policy{
		RetryMax: 3,
		WaitMin: 1 * time.Second,
		WaitMax: 10 * time.Second
		Backoff: ExponentialBackoff
	}

	n, _ := policy.Do(context.TODO(), func(retrying bool) (interface{}, error) {
		if retrying {
			// This is another attempt, let's do it another way
			err := do something else...
			return nil, err
		}

		err := do something...
		if err != nil {
			return nil, err
		}

		return res, nil
	})

HTTPClient uses the standard RoundTripper, enhanced by the policy wrapper,
which allows resiliency for failover requests with the ability to provide
a custom policy. It can be used exact same way as the standard http.Client.

	client := retry.NewHTTPClient(Policy{RetryMax: 3})
	res, err := client.Get(ts.URL)
	if err != nil {
		fmt.Println("total disaster...")
	}

	defer res.Body.Close()

WithPolicy will return a new copy of http.Client decorated with retryable
round tripper for a given policy.

	client := http.DefaultClient
	client = retry.WithPolicy(client, policy)

*/
package retry

package utils

import (
	"time"
)

type Retryable func() error

// The Retry function retries a given function a specified number of times with a delay between each
// attempt.
func Retry(retryFunc Retryable, attempts int, delay time.Duration) (err error) {

	for i := 0; i < attempts; i++ {
		err = retryFunc()
		if err == nil {
			return
		}
		// delay
		<-time.After(delay)
	}
	return
}

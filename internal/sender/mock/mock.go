package mockSender

import (
	"context"
	"sync"
)

type mockEmailSender struct {
}

func NewMockEmailSender(ctx context.Context, wg *sync.WaitGroup) (*mockEmailSender, error) {

	go func() {
		defer wg.Done()
		<-ctx.Done()
	}()

	return &mockEmailSender{}, nil
}

func (ms *mockEmailSender) Send(email, msg string) error {
	return nil
}

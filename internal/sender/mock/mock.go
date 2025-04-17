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

func (ms *mockEmailSender) Send(guid, msg string) error {
	//поиск user email по guid и отправка msg
	return nil
}

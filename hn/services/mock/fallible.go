package mock

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/bensadeh/circumflex/item"
)

// timeoutError satisfies net.Error with Timeout() returning true,
// simulating a real HTTP client timeout.
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "Client.Timeout exceeded while awaiting headers" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return false }

type FallibleService struct {
	mock Service
}

func NewFallibleService() FallibleService {
	return FallibleService{mock: Service{}}
}

func (f FallibleService) FetchItems(ctx context.Context, itemsToFetch int, category string) ([]*item.Story, error) {
	return f.mock.FetchItems(ctx, itemsToFetch, category)
}

func (f FallibleService) FetchComments(_ context.Context, _ int, _ func(fetched, total int)) (*item.Story, error) {
	n, _ := rand.Int(rand.Reader, big.NewInt(3))
	time.Sleep(time.Duration(1+n.Int64()) * time.Second)

	return nil, randomError()
}

func (f FallibleService) FetchItem(_ context.Context, _ int) (*item.Story, error) {
	return nil, randomError()
}

func randomError() error {
	errors := []error{
		&timeoutError{},
		fmt.Errorf("server returned status 403"),
		fmt.Errorf("server returned status 500"),
		fmt.Errorf("unexpected response from server"),
	}

	n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(errors))))

	return errors[n.Int64()]
}

package sharedstore

import (
	"context"
	"fmt"
	"time"
)

// Setter is used while the thread has the lock to prepare the data.
// Once the data is ready, the underlying Store will unlock the key and broadcast to other threads.
type Setter interface {
	Done(ctx context.Context, data interface{}, ttl time.Duration) error
}

type setter struct {
	key   string
	store Store
}

func (s *setter) Done(ctx context.Context, data interface{}, ttl time.Duration) error {
	err := s.store.setData(ctx, s.key, data, ttl)
	// The error can be overwritten below, so log it here just in case.
	if err != nil {
		log(ctx, err).Warn("unable to set data")
	}

	// Proceed even in case of error so threads get unlocked
	unlockErr := s.store.unlock(ctx, s.key)
	if unlockErr != nil {
		err = fmt.Errorf("unable to delete item lock: %w", unlockErr)
	}

	s.store.broadcast(ctx, s.key)

	return err
}

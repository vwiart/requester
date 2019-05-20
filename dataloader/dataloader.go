package dataloader

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vwiart/requester/client/users"
)

type result struct {
	user users.User
	err  error
}

type batch struct {
	mutex     sync.Mutex
	client    users.Client
	isRunning bool
	userIDs   map[int][]chan result
}

func (b *batch) register(userID int) chan result {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.userIDs == nil {
		b.userIDs = make(map[int][]chan result)
	}

	ch := make(chan result)
	results := b.userIDs[userID]
	if results == nil {
		results = make([]chan result, 0)
	}

	results = append(results, ch)
	b.userIDs[userID] = results
	fmt.Printf("registered: %d %d\n", userID, len(results))

	return ch
}

func (b *batch) load() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	var userIDs []int
	for k := range b.userIDs {
		userIDs = append(userIDs, k)
	}

	users, err := b.client.List(context.Background(), userIDs...)
	if err != nil {
		return err
	}

	for _, user := range users {
		chs := b.userIDs[user.ID]
		for _, ch := range chs {
			ch <- result{user: user}
		}
	}
	b.userIDs = nil
	return nil
}

type UserDataloader struct {
	batch *batch
}

func NewUserDataloader(client users.Client) UserDataloader {
	batch := &batch{client: client}

	go func() {
		for {
			select {
			case <-time.After(100 * time.Millisecond):
				batch.load()
			}
		}
	}()

	return UserDataloader{batch: batch}
}

func (udl UserDataloader) Load(ctx context.Context, userID int) (users.User, error) {
	ch := udl.batch.register(userID)

	res := <-ch
	if res.err != nil {
		return users.User{}, res.err
	}

	return res.user, nil
}

package users

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vwiart/requester/requester"
)

type Client interface {
	Get(ctx context.Context, userID int) (User, error)
	List(ctx context.Context, userID ...int) ([]User, error)
}

type userClient struct {
	requester requester.Client
}

func New(c requester.Client) Client {
	return userClient{requester: c}
}

type User struct {
	ID int `json:"id"`
}

func (c userClient) Get(ctx context.Context, userID int) (User, error) {
	resp, err := c.requester.Do(ctx,
		requester.Method(http.MethodGet),
		requester.URL(fmt.Sprintf("http://localhost:8080/users/%d", userID)),
	)
	if err != nil {
		return User{}, err
	}
	defer resp.Close()

	var buffer User
	if err := json.NewDecoder(resp).Decode(&buffer); err != nil {
		return User{}, err
	}

	return buffer, nil
}

func (c userClient) List(ctx context.Context, userIDs ...int) ([]User, error) {
	type request struct {
		UserIDs []int `json:"id"`
	}

	payload, err := json.Marshal(request{UserIDs: userIDs})
	if err != nil {
		return nil, err
	}

	resp, err := c.requester.Do(ctx,
		requester.Method(http.MethodGet),
		requester.URL("http://localhost:8080/users"),
		requester.Body(bytes.NewReader(payload)),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	type response struct {
		UserIDs []int `json:"id"`
	}

	var buffer response
	if err := json.NewDecoder(resp).Decode(&buffer); err != nil {
		return nil, err
	}

	user := make([]User, len(buffer.UserIDs))
	for i, id := range buffer.UserIDs {
		user[i] = User{id}
	}

	return user, nil
}

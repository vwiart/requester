package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/vwiart/requester/client/users"
	"github.com/vwiart/requester/dataloader"
	"github.com/vwiart/requester/requester"
)

type response struct {
	ID int `json:"id"`
}

func request(dl dataloader.UserDataloader) error {
	start := time.Now().UTC()
	user, err := dl.Load(context.Background(), rand.Intn(10))
	if err != nil {
		fmt.Printf("err=%+v\n", err)
		return err
	}
	fmt.Printf("user=%+v (Elapsed: %v)\n", user, time.Since(start))

	return nil
}

func main() {
	go func() {
		counter := 0
		mux := http.NewServeMux()
		mux.HandleFunc("/users/1", func(w http.ResponseWriter, r *http.Request) {
			counter++
			fmt.Println(counter)
			json.NewEncoder(w).Encode(response{ID: 1})
		})

		mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
			io.Copy(w, r.Body)
			fmt.Println("load-many")
		})

		if err := http.ListenAndServe(":8080", mux); err != nil {
			fmt.Printf("error: %+v\n", err)
		}
	}()

	requester := requester.New()
	userClient := users.New(requester)
	userDataloader := dataloader.NewUserDataloader(userClient)

	for {
		select {
		case <-time.After(1 * time.Millisecond):
			go func() {
				if err := request(userDataloader); err != nil {
					fmt.Printf("error: %+v\n", err)
				}
			}()

		}
	}
}

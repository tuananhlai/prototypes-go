package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

const (
	baseURL = "http://localhost:8081/value"
)

func main() {
	client := &http.Client{}

	var wg sync.WaitGroup

	for range 100 {
		wg.Go(func() {
			currentValue, err := getValue(client)
			if err != nil {
				log.Printf("error getting current value: %v", err)
				return
			}

			putValueReqBody, _ := json.Marshal(UpdateCounterRequest{
				Value: currentValue + 1,
			})
			putValueReq, err := http.NewRequest("PUT", baseURL, bytes.NewReader(putValueReqBody))
			if err != nil {
				log.Printf("error creating put value request: %v", err)
				return
			}

			_, err = client.Do(putValueReq)
			if err != nil {
				log.Printf("error updating value: %v", err)
			}
		})
	}
	wg.Wait()

	finalValue, err := getValue(client)
	if err != nil {
		log.Printf("error getting final value: %v", err)
		return
	}
	log.Printf("Final value: %d", finalValue)
}

func getValue(client *http.Client) (int64, error) {
	res, err := client.Get(baseURL)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	var resBody GetCounterResponse
	if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		return 0, err
	}

	return resBody.Value, nil
}

type UpdateCounterRequest struct {
	Value int64 `json:"value"`
}

type GetCounterResponse struct {
	Value int64 `json:"value"`
}

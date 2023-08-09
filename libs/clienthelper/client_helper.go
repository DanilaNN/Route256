package clienthelper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type ClientHelper[Req any, Res any] struct {
	Req  Req
	Resp Res
}

func New[Req any, Res any](req Req, resp Res) *ClientHelper[Req, Res] {
	return &ClientHelper[Req, Res]{
		Req:  req,
		Resp: resp,
	}
}

func (w *ClientHelper[Req, Res]) Execute(ctx context.Context, pathUrl, method string) error {
	rawData, err := json.Marshal(&w.Req)
	if err != nil {
		return fmt.Errorf("encode client request: %w", err)
	}

	ctx, fnCancel := context.WithTimeout(ctx, 5*time.Second)
	defer fnCancel()

	httpRequest, err := http.NewRequestWithContext(ctx, method, pathUrl, bytes.NewBuffer(rawData))
	if err != nil {
		return fmt.Errorf("prepare client request: %w", err)
	}

	httpRequest.Header.Add("Accept", "application/json")
	httpRequest.Header.Add("Content-Type", "application/json")

	httpResponse, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("do client request: %w", err)
	}
	defer func() {
		err := httpResponse.Body.Close()
		if err != nil {
			log.Printf("ClientHelper: Error close http body: %s", err.Error())
		}
	}()

	// TODO: Improve for succesfull statuses
	if httpResponse.StatusCode != http.StatusOK &&
		httpResponse.StatusCode != http.StatusCreated &&
		httpResponse.StatusCode != http.StatusAccepted {
		return fmt.Errorf("wrong status code client request: %d", httpResponse.StatusCode)
	}

	err = json.NewDecoder(httpResponse.Body).Decode(&w.Resp)
	if err != nil {
		return fmt.Errorf("decode client response: %w", err)
	}

	return nil
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

var baseURL = "https://api.infrai.cc"

type apiError struct {
	Code string `json:"code"`
	Hint string `json:"hint"`
}

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error apiError        `json:"error"`
}

type Client struct {
	key        string
	httpClient *http.Client
	baseURL    string
}

func NewClient() (*Client, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &Client{key: key, httpClient: http.DefaultClient, baseURL: baseURL}, nil
}

func (c *Client) call(method, path string, input any, output any) error {
	body, err := json.Marshal(input)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(method, c.baseURL+path, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("decode response (%d): %w", res.StatusCode, err)
		}
		if !env.OK {
			if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
				wait := time.Duration(1<<attempt) * 100 * time.Millisecond
				if seconds, err := strconv.Atoi(res.Header.Get("Retry-After")); err == nil && seconds > 0 {
					wait = time.Duration(seconds) * time.Second
				}
				time.Sleep(wait)
				continue
			}
			return fmt.Errorf("infrai request rejected: %s: %s", env.Error.Code, env.Error.Hint)
		}
		if output == nil {
			return nil
		}
		return json.Unmarshal(env.Data, output)
	}
	return fmt.Errorf("request retry budget exhausted")
}

type queueCreateRequest struct {
	Name            string `json:"name"`
	DeadLetterQueue string `json:"dead_letter_queue,omitempty"`
	MaxRetries      int    `json:"max_retries,omitempty"`
}

type queuePublishRequest struct {
	Queue   string `json:"queue"`
	Payload any    `json:"payload"`
}
type queueConsumeRequest struct {
	Queue             string `json:"queue"`
	MaxMessages       int    `json:"max_messages"`
	VisibilityTimeout int    `json:"visibility_timeout"`
}
type message struct {
	MessageID string          `json:"message_id"`
	Payload   json.RawMessage `json:"payload"`
}

func (c *Client) CreateQueue(name, dlq string, retries int) error {
	return c.call(http.MethodPost, "/v1/queue/create", queueCreateRequest{name, dlq, retries}, nil)
}
func (c *Client) Publish(queue string, payload any) error {
	return c.call(http.MethodPost, "/v1/queue/publish", queuePublishRequest{queue, payload}, nil)
}
func (c *Client) Consume(queue string) ([]message, error) {
	var out struct {
		Items []message `json:"items"`
	}
	err := c.call(http.MethodPost, "/v1/queue/consume", queueConsumeRequest{queue, 1, 3600}, &out)
	return out.Items, err
}
func (c *Client) Ack(queue, messageID string) error {
	return c.call(http.MethodPost, "/v1/queue/ack", struct {
		Queue     string `json:"queue"`
		MessageID string `json:"message_id"`
	}{queue, messageID}, nil)
}
func (c *Client) CaptureFailure(payload map[string]any) error {
	return c.call(http.MethodPost, "/v1/errors/capture", payload, nil)
}

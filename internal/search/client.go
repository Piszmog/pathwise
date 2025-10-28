package search

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

var ErrSearchFailed = errors.New("search failed")

type Client struct {
	c       *http.Client
	baseURL string
	logger  *slog.Logger
}

func NewClient(logger *slog.Logger, client *http.Client, baseURL string) *Client {
	return &Client{
		c:       client,
		baseURL: baseURL,
		logger:  logger,
	}
}

func (c *Client) Search(ctx context.Context, searchReq Request) ([]JobListing, error) {
	u, err := url.JoinPath(c.baseURL, "/api/v1/search")
	if err != nil {
		return nil, fmt.Errorf("failed to create url: %w", err)
	}

	b, err := json.Marshal(searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("failed to build new request: %w", err)
	}

	id := uuid.New().String()
	req.Header.Add("X-Correlation-ID", id)

	c.logger.DebugContext(ctx, "searching for job listings", "request", searchReq, "url", u)
	resp, err := c.c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call search api for %s: %w", id, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.ErrorContext(ctx, "failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		var errResp Error
		err = json.NewDecoder(resp.Body).Decode(&errResp)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal error response for %s: %w", id, err)
		}

		return nil, fmt.Errorf("%w for %s: %s", ErrSearchFailed, id, errResp.Message)
	}

	var searchResp Response
	err = json.NewDecoder(resp.Body).Decode(&searchResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response for %s: %w", id, err)
	}

	return searchResp.JobListings, nil
}

package github

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const userAgent = "Heatdot/1.0"

func FetchProfileHTML(ctx context.Context, profileURL string) ([]byte, error) {
	if profileURL == "" {
		return nil, fmt.Errorf("profile URL is empty")
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, profileURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", userAgent)

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("unexpected status: %s", resp.Status)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		return body, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("request failed")
	}
	return nil, lastErr
}

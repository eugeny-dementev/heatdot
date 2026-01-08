package github

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const userAgent = "Heatdot/1.0"

func FetchProfileHTML(ctx context.Context, profileURL string) ([]byte, error) {
	if profileURL == "" {
		return nil, fmt.Errorf("profile URL is empty")
	}

	if contribURL, ok := contributionsURL(profileURL); ok {
		body, err := fetchURL(ctx, contribURL)
		if err == nil {
			return body, nil
		}
	}

	body, err := fetchURL(ctx, profileURL)
	if err == nil {
		return body, nil
	}
	return nil, err
}

func fetchURL(ctx context.Context, target string) ([]byte, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
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

func contributionsURL(profileURL string) (string, bool) {
	parsed, err := url.Parse(profileURL)
	if err != nil {
		return "", false
	}
	if parsed.Scheme == "" {
		parsed, err = url.Parse("https://" + profileURL)
		if err != nil {
			return "", false
		}
	}

	path := strings.Trim(parsed.Path, "/")
	if path == "" {
		return "", false
	}
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		return "", false
	}
	user := parts[0]

	contrib := &url.URL{
		Scheme: parsed.Scheme,
		Host:   parsed.Host,
		Path:   "/users/" + user + "/contributions",
	}
	return contrib.String(), true
}

// Package testutil provides helpers for catalog CI scripts and test harnesses.
// It is not imported by the failure-mode servers under failures/ — those are
// single-file programs that depend only on the standard library.
package testutil

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// WaitReady polls url with GET until it returns a 2xx status or ctx is
// cancelled. It sleeps interval between attempts.
//
// CI uses this to block until a failure-mode server has bound its port and
// answered its /health endpoint before running clrsig against it.
func WaitReady(ctx context.Context, url string, interval time.Duration) error {
	client := &http.Client{Timeout: 2 * time.Second}
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("build request: %w", err)
		}
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("server at %s not ready: %w", url, ctx.Err())
		case <-time.After(interval):
		}
	}
}

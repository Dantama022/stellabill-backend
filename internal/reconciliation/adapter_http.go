package reconciliation

import (
<<<<<<< HEAD
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"stellabill-backend/internal/httpclient"
=======
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"

    "go.uber.org/zap"

    "stellarbill-backend/internal/httpclient"
>>>>>>> upstream/main
)

// HTTPAdapter fetches snapshots from a configured HTTP endpoint.
type HTTPAdapter struct {
	Client *httpclient.Client
	URL    string
	// Optional Authorization header value (e.g., Bearer <token>)
	AuthHeader string
}

// NewHTTPAdapter creates an adapter that will GET snapshots from url.
<<<<<<< HEAD
func NewHTTPAdapter(url string, authHeader string) *HTTPAdapter {
	return &HTTPAdapter{Client: httpclient.NewClient(), URL: url, AuthHeader: authHeader}
=======
func NewHTTPAdapter(urlStr string, authHeader string, logger *zap.Logger) *HTTPAdapter {
    u, err := url.Parse(urlStr)
    host := "unknown"
    if err == nil && u.Host != "" {
        host = u.Host
    }
    return &HTTPAdapter{Client: httpclient.NewClient(host, logger), URL: urlStr, AuthHeader: authHeader}
>>>>>>> upstream/main
}

// FetchSnapshots implements Adapter.
func (h *HTTPAdapter) FetchSnapshots(ctx context.Context) ([]Snapshot, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.URL, nil)
	if err != nil {
		return nil, err
	}
	if h.AuthHeader != "" {
		req.Header.Set("Authorization", h.AuthHeader)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := h.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resp.Body != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var snaps []Snapshot
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&snaps); err != nil {
		return nil, err
	}
	return snaps, nil
}


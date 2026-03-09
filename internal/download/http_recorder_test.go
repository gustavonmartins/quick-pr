package download

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func init() {
	// Replace the default HTTP client with a recording/caching transport
	HTTPClient = &http.Client{
		Transport: &RecordingTransport{
			cacheDir: "../../testdata/cache",
			real:     http.DefaultTransport,
		},
	}
}

// RecordingTransport is an http.RoundTripper that caches HTTP responses.
type RecordingTransport struct {
	cacheDir string
	real     http.RoundTripper
}

func (r *RecordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cacheFile := r.cacheFilePath(req)

	// Try cache first
	if data, err := os.ReadFile(cacheFile); err == nil {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(data)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	}

	// Cache MISS - call real HTTP
	resp, err := r.real.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if closeErr := resp.Body.Close(); closeErr != nil {
		return nil, closeErr
	}
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		if err := os.MkdirAll(filepath.Dir(cacheFile), 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(cacheFile, body, 0644); err != nil {
			return nil, err
		}
	}

	resp.Body = io.NopCloser(bytes.NewReader(body))
	return resp, nil
}

func (r *RecordingTransport) cacheFilePath(req *http.Request) string {
	key := req.Method + "_" + req.URL.String()
	hash := sha256.Sum256([]byte(key))
	filename := fmt.Sprintf("%x.json", hash[:8])
	return filepath.Join(r.cacheDir, filename)
}

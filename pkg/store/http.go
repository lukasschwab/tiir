package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/lukasschwab/tiir/pkg/text"
)

var (
	errNotFound = errors.New("not found")
)

func UseHTTP(baseURL, apiSecret string) (Store, error) {
	url, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	return &httpStore{baseURL: url, apiSecret: apiSecret}, nil
}

// httpStore implements Store for a remote cmd/server process.
//
// NOTE: if we had a proto-defined service, this would probably wrap that (and
// the generated code would determine the inner values necessary to specify and
// connect). As it is, this mirrors the routes and renderers exposed by
// cmd/server; these are hidden dependencies!
type httpStore struct {
	baseURL   *url.URL
	apiSecret string
}

// newRequest wraps http.NewRequest for requests rooted at h.baseURL.
func (h *httpStore) newRequest(method string, body io.Reader, path ...string) (*http.Request, error) {
	requestURL := h.baseURL.JoinPath(path...).String()
	req, err := http.NewRequest(method, requestURL, body)
	if err != nil {
		return req, err
	}
	req.Header.Add(fiber.HeaderContentType, "application/json")
	if h.apiSecret != "" {
		req.Header.Add(fiber.HeaderAuthorization, fmt.Sprintf("Bearer %s", h.apiSecret))
	}
	return req, nil
}

// Read implements Store.
func (h *httpStore) Read(id string) (*text.Text, error) {
	result := new(text.Text)
	if req, err := h.newRequest(http.MethodGet, nil, "texts", id); err != nil {
		return nil, fmt.Errorf("error building request: %w", err)
	} else if resp, err := http.DefaultClient.Do(req); err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	} else if resp.StatusCode == http.StatusNotFound {
		return nil, errNotFound
	} else if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	} else if err := result.Validate(); err != nil {
		return nil, fmt.Errorf("result is invalid text: %w", err)
	}
	return result, nil
}

// Upsert implements Store. It reads before writing to decide whether to call
// the server's POST route or its PATCH route, since cmd/server doesn't expose
// an upsert route.
func (h *httpStore) Upsert(t *text.Text) (*text.Text, error) {
	marshaled, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("error encoding text: %w", err)
	}
	body := bytes.NewReader(marshaled)

	var method string
	var path []string
	if _, err := h.Read(t.ID); errors.Is(err, errNotFound) {
		// The record doesn't exist; create it.
		method, path = http.MethodPost, []string{"texts"}
	} else if err != nil {
		return nil, fmt.Errorf("unexpected error: %w", err)
	} else {
		// It exists; update it.
		method, path = http.MethodPatch, []string{"texts", t.ID}
	}

	result := new(text.Text)
	if req, err := h.newRequest(method, body, path...); err != nil {
		return nil, fmt.Errorf("error building request: %w", err)
	} else if resp, err := http.DefaultClient.Do(req); err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	} else if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	} else if err := result.Validate(); err != nil {
		return nil, fmt.Errorf("result is invalid text: %w", err)
	}
	return result, nil
}

// Delete implements Store.
func (h *httpStore) Delete(id string) (*text.Text, error) {
	result := new(text.Text)
	if req, err := h.newRequest(http.MethodDelete, nil, "texts", id); err != nil {
		return nil, fmt.Errorf("error building request: %w", err)
	} else if resp, err := http.DefaultClient.Do(req); err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	} else if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	} else if err := result.Validate(); err != nil {
		return nil, fmt.Errorf("result is invalid text: %w", err)
	}
	return result, nil
}

// List implements Store. It re-sorts the response accoding to c and d.
func (h *httpStore) List(c text.Comparator, d text.Direction) ([]*text.Text, error) {
	req, err := h.newRequest(http.MethodGet, nil, "texts")
	if err != nil {
		return nil, fmt.Errorf("error building request: %w", err)
	}
	req.Header.Add("Accept-Encoding", "application/json")
	q := req.URL.Query()
	q.Add("format", "application/json")
	req.URL.RawQuery = q.Encode()

	var result []*text.Text
	if resp, err := http.DefaultClient.Do(req); err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
		// TODO: handle "not found" status.
	} else if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	text.Sort(result).By(c, d)
	return result, nil
}

// Close implements Store.
func (h *httpStore) Close() error {
	return nil
}

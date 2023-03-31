package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lukasschwab/tiir/pkg/text"
)

func UseHTTP(baseURL string) Store {
	return &httpStore{baseURL: baseURL}
}

// httpStore implements Store for a remote cmd/server process.
//
// NOTE: if we had a proto-defined service, this would probably wrap that (and
// the generated code would determine the inner values necessary to specify and
// connect). As it is, this mirrors the routes and renderers exposed by
// cmd/server; these are hidden dependencies!
type httpStore struct {
	baseURL string
}

// Read implements Store.
func (h *httpStore) Read(id string) (*text.Text, error) {
	result := new(text.Text)
	if resp, err := http.Get(fmt.Sprintf("%s/texts/%s", h.baseURL, id)); err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
		// TODO: handle "not found" status.
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

	var method, url string
	if _, err := h.Read(t.ID); err != nil {
		// The record presumably doesn't exist; create it.
		// TODO: be more selective about the errors accepted here.
		method, url = http.MethodPost, fmt.Sprintf("%s/texts", h.baseURL)
	} else {
		// It exists; update it.
		method, url = http.MethodPatch, fmt.Sprintf("%s/texts/%s", h.baseURL, t.ID)
	}

	result := new(text.Text)
	if req, err := http.NewRequest(method, url, body); err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
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
	if req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/texts/%s", h.baseURL, id), nil); err != nil {
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
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/texts", h.baseURL), nil)
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

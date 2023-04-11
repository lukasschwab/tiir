package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/lukasschwab/tiir/pkg/text"
)

var (
	errNotFound = errors.New("not found")
)

// UseHTTP requests to a remote [github.com/lukasschwab/tiir/cmd/server]
// instance (hosted at baseURL, accepting secret apiSecret) to read and write
// texts.
func UseHTTP(baseURL, apiSecret string) (Interface, error) {
	url, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	return &HTTP{baseURL: url, apiSecret: apiSecret}, nil
}

// HTTP implements [Interface] for a remote cmd/server process. See [UseHTTP].
//
// NOTE: if we had a proto-defined service, this would probably wrap that (and
// the generated code would determine the inner values necessary to specify and
// connect). As it is, this mirrors the routes and renderers exposed by
// [github.com/lukasschwab/tiir/cmd/server]; these are hidden dependencies!
type HTTP struct {
	baseURL   *url.URL
	apiSecret string
}

// newRequest wraps http.NewRequest for requests rooted at h.baseURL.
func (h *HTTP) newRequest(method string, body io.Reader, path ...string) (*http.Request, error) {
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

func checkStatus(resp *http.Response) error {
	if resp.StatusCode/100 != 2 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body: %v", err)
		}
		return fmt.Errorf("server responded %d: %s", resp.StatusCode, body)
	}
	return nil
}

// Read implements [Interface].
func (h *HTTP) Read(id string) (*text.Text, error) {
	req, err := h.newRequest(http.MethodGet, nil, "texts", id)
	if err != nil {
		return nil, fmt.Errorf("error building request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	result := new(text.Text)
	if resp.StatusCode == http.StatusNotFound {
		return nil, errNotFound
	} else if err := checkStatus(resp); err != nil {
		return nil, err
	} else if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	} else if err := result.Validate(); err != nil {
		return nil, fmt.Errorf("result is invalid text: %w", err)
	}
	return result, nil
}

// Upsert implements [Interface]. It reads before writing to decide whether to call
// the server's POST route or its PATCH route, since cmd/server doesn't expose
// an upsert route.
func (h *HTTP) Upsert(t *text.Text) (*text.Text, error) {
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

	req, err := h.newRequest(method, body, path...)
	if err != nil {
		return nil, fmt.Errorf("error building request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	result := new(text.Text)
	if err := checkStatus(resp); err != nil {
		return nil, err
	} else if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	} else if err := result.Validate(); err != nil {
		return nil, fmt.Errorf("result is invalid text: %w", err)
	}
	return result, nil
}

// Delete implements [Interface].
func (h *HTTP) Delete(id string) (*text.Text, error) {
	req, err := h.newRequest(http.MethodDelete, nil, "texts", id)
	if err != nil {
		return nil, fmt.Errorf("error building request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	result := new(text.Text)
	if err := checkStatus(resp); err != nil {
		return nil, err
	} else if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	} else if err := result.Validate(); err != nil {
		return nil, fmt.Errorf("result is invalid text: %w", err)
	}
	return result, nil
}

// List implements [Interface]. It re-sorts the response accoding to c and d.
func (h *HTTP) List(c text.Comparator, d text.Direction) ([]*text.Text, error) {
	req, err := h.newRequest(http.MethodGet, nil, "texts")
	if err != nil {
		return nil, fmt.Errorf("error building request: %w", err)
	}
	// Don't want the default HTML representation.
	req.Header.Add("Accept-Encoding", "application/json")
	q := req.URL.Query()
	q.Add("format", "application/json")
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	var result []*text.Text
	if err := checkStatus(resp); err != nil {
		return nil, err
	} else if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	text.Sort(result).By(c, d)
	return result, nil
}

// Close implements [Interface].
func (h *HTTP) Close() error {
	return nil
}

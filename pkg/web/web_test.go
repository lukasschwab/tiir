package web_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lukasschwab/tiir/pkg/text"
	"github.com/lukasschwab/tiir/pkg/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		file, err := os.Open("./tests/standard.html")
		require.NoError(t, err)
		defer file.Close()

		_, err = io.Copy(w, file)
		require.NoError(t, err)
	}))
	defer server.Close()

	initial, err := web.WebMetadata(server.URL)
	assert.NoError(t, err)

	assert.Equal(t, server.URL, initial.URL)
	assert.Equal(t, "Test Case 1", initial.Title)
	assert.Equal(t, "Author One", initial.Author)
}

func TestMetadata(t *testing.T) {
	cases, err := filepath.Glob("./tests/*.html")
	assert.NoError(t, err)

	for _, testCase := range cases {
		t.Run(testCase, func(t *testing.T) {
			file, err := os.Open(testCase)
			require.NoError(t, err)
			defer file.Close()

			initial, err := web.Metadata(file)
			require.NoError(t, err)

			expectedFile, err := os.Open(fmt.Sprintf("%s.json", strings.TrimSuffix(testCase, ".html")))
			require.NoError(t, err)

			var expected text.Text
			err = json.NewDecoder(expectedFile).Decode(&expected)
			require.NoError(t, err)

			assert.Equal(t, expected, *initial)
		})
	}
}

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

// Bodge for generating JSON text for a new HTML test file. See [TestMetadata].
func generateJSONFiles() {
	cases, err := filepath.Glob("./tests/*.html")
	if err != nil {
		panic(err)
	}

	for _, testCase := range cases {
		file, err := os.Open(testCase)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		root := strings.TrimSuffix(testCase, filepath.Ext(testCase))

		initial, err := web.Metadata(file)
		if err != nil {
			panic(err)
		}

		bytes, err := json.Marshal(initial)

		err = os.WriteFile(fmt.Sprintf("%v.json", root), bytes, 0644)
		if err != nil {
			panic(err)
		}
	}
}

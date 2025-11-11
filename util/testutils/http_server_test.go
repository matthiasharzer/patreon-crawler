package testutils_test

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/MatthiasHarzer/patreon-crawler/util/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeRequest(method string, url url.URL) (*http.Response, error) {
	client := http.Client{}
	request, err := http.NewRequest(method, url.String(), nil)
	if err != nil {
		return nil, err
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error, status code: %d, status: %s", response.StatusCode, response.Status)
	}

	return response, nil
}

func TestCreateHttpServer(t *testing.T) {
	t.Run("creates an HTTP server", func(t *testing.T) {
		serverURL, closeServer := testutils.HTTPServer(map[string]http.HandlerFunc{
			"/": func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}})
		defer closeServer()

		require.NotEmpty(t, serverURL)
		assert.Equal(t, serverURL.Scheme, "http")
	})

	t.Run("closes the HTTP server", func(t *testing.T) {
		serverURL, closeServer := testutils.HTTPServer(map[string]http.HandlerFunc{
			"/": func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}})

		response, err := makeRequest(http.MethodGet, serverURL)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		closeServer()

		response, err = makeRequest(http.MethodGet, serverURL)
		require.Error(t, err)
		assert.Nil(t, response)
	})

	t.Run("serves the expected response", func(t *testing.T) {
		serverURL, closeServer := testutils.HTTPServer(map[string]http.HandlerFunc{
			"/": func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("Hello, World!"))
				require.NoError(t, err)
			}})
		defer closeServer()

		response, err := makeRequest(http.MethodGet, serverURL)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		body, err := io.ReadAll(response.Body)
		require.NoError(t, err)
		assert.Equal(t, "Hello, World!", string(body))
	})

	t.Run("serves the expected response for a specific path", func(t *testing.T) {
		serverURL, closeServer := testutils.HTTPServer(map[string]http.HandlerFunc{
			"/hello": func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("Hello, World!"))
				require.NoError(t, err)
			},
			"/goodbye": func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("Goodbye, World!"))
				require.NoError(t, err)
			},
		})
		defer closeServer()

		response, err := makeRequest(http.MethodGet, serverURL)
		assert.Error(t, err)
		assert.Nil(t, response)

		response, err = makeRequest(http.MethodGet, *serverURL.ResolveReference(&url.URL{Path: "/hello"}))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		body, err := io.ReadAll(response.Body)
		assert.NoError(t, err)
		assert.Equal(t, "Hello, World!", string(body))

		response, err = makeRequest(http.MethodGet, *serverURL.ResolveReference(&url.URL{Path: "/goodbye"}))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		body, err = io.ReadAll(response.Body)
		assert.NoError(t, err)
		assert.Equal(t, "Goodbye, World!", string(body))
	})
}

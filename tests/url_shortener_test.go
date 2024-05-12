package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"url-shortener/tests/suite"

	"github.com/stretchr/testify/assert"

	"github.com/brianvoe/gofakeit/v6"
)

var (
	host   = "localhost"
	scheme = "http"
)

func init() {
	_, st := suite.New(&testing.T{})

	host += st.Cfg.HttpServer.Port
}

func TestUrlShortener_Save(t *testing.T) {
	u := url.URL{
		Scheme: scheme,
		Host:   host,
	}

	tests := []struct {
		name               string
		username           string
		password           string
		body               map[string]interface{}
		expectedStatusCode int
		expectedErr        bool
	}{
		{
			name:     "Normal save 1",
			username: "pasha",
			password: "1234",
			body: map[string]interface{}{
				"url":   gofakeit.URL(),
				"alias": gofakeit.Word() + "_" + gofakeit.Word(),
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name:     "Normal save 2",
			username: "pasha",
			password: "1234",
			body: map[string]interface{}{
				"url":   gofakeit.URL(),
				"alias": "",
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name:     "Normal save 3",
			username: "pasha",
			password: "1234",
			body: map[string]interface{}{
				"url":   gofakeit.URL(),
				"alias": "kaif",
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name:     "Normal save 4",
			username: "pasha",
			password: "1234",
			body: map[string]interface{}{
				"url":   "https://stepik.org/learn",
				"alias": "gmail",
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name:     "Normal save 5",
			username: "pasha",
			password: "1234",
			body: map[string]interface{}{
				"url":   "https://stepik.org/learn",
				"alias": "lms",
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name:     "Save already exist alias from other user 1",
			username: "vova",
			password: "9876",
			body: map[string]interface{}{
				"url":   "https://stepik.org/learn",
				"alias": "lms",
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name:     "Save already exist alias from other user 2",
			username: "vova",
			password: "9876",
			body: map[string]interface{}{
				"url":   "https://stepik.org/learn",
				"alias": "gmail",
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name:     "Alias already exists 1",
			username: "vova",
			password: "9876",
			body: map[string]interface{}{
				"url":   gofakeit.URL(),
				"alias": "lms",
			},
			expectedStatusCode: 400,
			expectedErr:        true,
		},
		{
			name:     "Alias already exists 2",
			username: "pasha",
			password: "1234",
			body: map[string]interface{}{
				"url":   gofakeit.URL(),
				"alias": "gmail",
			},
			expectedStatusCode: 400,
			expectedErr:        true,
		},
		{
			name:     "Invalid JSON",
			username: "pasha",
			password: "1234",
			body: map[string]interface{}{
				"url":   500,
				"alias": "",
			},
			expectedStatusCode: 400,
			expectedErr:        true,
		},
		{
			name:     "Invalid url",
			username: "pasha",
			password: "1234",
			body: map[string]interface{}{
				"url":   gofakeit.Word(),
				"alias": gofakeit.Word(),
			},
			expectedStatusCode: 400,
			expectedErr:        true,
		},
		{
			name:               "empty body",
			username:           "pasha",
			password:           "1234",
			body:               map[string]interface{}{},
			expectedStatusCode: 400,
			expectedErr:        true,
		},
	}

	var wg sync.WaitGroup

	for range 1 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					jsonBody, err := json.Marshal(tt.body)
					assert.NoError(t, err)

					req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(jsonBody))
					assert.NoError(t, err)

					req.SetBasicAuth(tt.username, tt.password)

					resp, err := http.DefaultClient.Do(req)
					assert.NoError(t, err)

					defer func() { _ = resp.Body.Close() }()

					jsonData, err := io.ReadAll(resp.Body)
					assert.NoError(t, err)

					var data map[string]interface{}
					err = json.Unmarshal(jsonData, &data)
					assert.NoError(t, err)

					assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)
					assert.Equal(t, data["status"].(string) == "Error", tt.expectedErr)
				})
			}
		}()
	}

	wg.Wait()
}

func TestUrlShortener_Redirect(t *testing.T) {
	tests := []struct {
		name               string
		username           string
		alias              string
		expectedStatusCode int
		expectedErr        bool
	}{
		{
			name:               "Normal Redirect 1",
			username:           "pasha",
			alias:              "gmail",
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name:               "Normal Redirect 2",
			username:           "pasha",
			alias:              "lms",
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name:               "Alias not found 1",
			username:           "pasha",
			alias:              gofakeit.Word(),
			expectedStatusCode: 400,
			expectedErr:        true,
		},
		{
			name:               "Alias not found 2",
			username:           "pasha",
			alias:              gofakeit.Word(),
			expectedStatusCode: 400,
			expectedErr:        true,
		},
	}

	var wg sync.WaitGroup

	for range 1 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					u := url.URL{
						Scheme: scheme,
						Host:   host,
						Path:   tt.username + "/" + tt.alias,
					}

					req, err := http.NewRequest(http.MethodGet, u.String(), nil)
					assert.NoError(t, err)

					resp, err := http.DefaultClient.Do(req)
					assert.NoError(t, err)

					defer func() { _ = resp.Body.Close() }()

					assert.NoError(t, err)

					assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)
				})
			}
		}()
	}
	wg.Wait()
}

func TestUrlShortener_Update(t *testing.T) {
	tests := []struct {
		name               string
		username           string
		password           string
		body               map[string]interface{}
		expectedStatusCode int
		expectedErr        bool
	}{
		{
			name:     "Normal update 1",
			username: "pasha",
			password: "1234",
			body: map[string]interface{}{
				"alias":     "gmail",
				"new_alias": "mail",
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name:     "Normal update 2",
			username: "pasha",
			password: "1234",
			body: map[string]interface{}{
				"alias":     "lms",
				"new_alias": "yandex_lms",
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name:     "Alias does not exists",
			username: "pasha",
			password: "1234",
			body: map[string]interface{}{
				"alias":     gofakeit.Word(),
				"new_alias": gofakeit.Word(),
			},
			expectedStatusCode: 400,
			expectedErr:        true,
		},
		{
			name:     "NewAlias already exists",
			username: "pasha",
			password: "1234",
			body: map[string]interface{}{
				"alias":     "mail",
				"new_alias": "kaif",
			},
			expectedStatusCode: 400,
			expectedErr:        true,
		},
	}

	var wg sync.WaitGroup

	for range 1 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					u := url.URL{
						Scheme: scheme,
						Host:   host,
					}

					jsonBody, err := json.Marshal(tt.body)
					assert.NoError(t, err)

					req, err := http.NewRequest(http.MethodPut, u.String(), bytes.NewBuffer(jsonBody))
					assert.NoError(t, err)

					req.SetBasicAuth(tt.username, tt.password)

					resp, err := http.DefaultClient.Do(req)
					assert.NoError(t, err)

					defer func() { _ = resp.Body.Close() }()

					jsonData, err := io.ReadAll(resp.Body)
					assert.NoError(t, err)

					var data map[string]interface{}
					assert.NoError(t, json.Unmarshal(jsonData, &data))

					assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)
					assert.Equal(t, tt.expectedErr, data["status"].(string) == "Error")
				})
			}
		}()
	}
	wg.Wait()
}

func TestUrlShortener_Delete(t *testing.T) {
	tests := []struct {
		name               string
		username           string
		password           string
		body               map[string]interface{}
		expectedStatusCode int
		expectedErr        bool
	}{
		{
			name:     "Normal delete 1",
			username: "pasha",
			password: "1234",
			body: map[string]interface{}{
				"alias": "mail",
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name:     "Normal delete 2",
			username: "pasha",
			password: "1234",
			body: map[string]interface{}{
				"alias": "yandex_lms",
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name:     "Normal delete 3",
			username: "pasha",
			password: "1234",
			body: map[string]interface{}{
				"alias": "kaif",
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name:     "Normal delete 4",
			username: "vova",
			password: "9876",
			body: map[string]interface{}{
				"alias": "lms",
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name:     "Normal delete 5",
			username: "vova",
			password: "9876",
			body: map[string]interface{}{
				"alias": "gmail",
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name:     "Alias not found 1",
			username: "pasha",
			password: "1234",
			body: map[string]interface{}{
				"alias": gofakeit.Word(),
			},
			expectedStatusCode: 400,
			expectedErr:        true,
		},
		{
			name:     "Alias not found 2",
			username: "pasha",
			password: "1234",
			body: map[string]interface{}{
				"alias": gofakeit.Word(),
			},
			expectedStatusCode: 400,
			expectedErr:        true,
		},
	}

	var wg sync.WaitGroup

	for range 1 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					u := url.URL{
						Scheme: scheme,
						Host:   host,
					}

					jsonBody, err := json.Marshal(tt.body)
					assert.NoError(t, err)

					req, err := http.NewRequest(http.MethodDelete, u.String(), bytes.NewBuffer(jsonBody))
					assert.NoError(t, err)

					req.SetBasicAuth(tt.username, tt.password)

					resp, err := http.DefaultClient.Do(req)
					assert.NoError(t, err)

					defer func() { _ = resp.Body.Close() }()

					jsonData, err := io.ReadAll(resp.Body)
					assert.NoError(t, err)

					var data map[string]interface{}
					assert.NoError(t, json.Unmarshal(jsonData, &data))

					assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)
					assert.Equal(t, tt.expectedErr, data["status"].(string) == "Error")
				})
			}
		}()
	}
	wg.Wait()
}

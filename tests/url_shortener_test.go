package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/url"
	"testing"
	"url-shortener/internal/http-server/save"
	"url-shortener/internal/lib/random"
)

const (
	host   = "localhost:8080"
	scheme = "http"
)

func TestUrlShortener_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: scheme,
		Host:   host,
	}

	e := httpexpect.Default(t, u.String())

	e.POST("/").
		WithJSON(save.Request{Url: gofakeit.URL(), Alias: random.Alias(10)}).
		WithBasicAuth("pasha", "1234").
		Expect().
		Status(200).
		JSON().Object().ContainsKey("alias")
}

func TestUrlShortener_Save(t *testing.T) {
	tests := []struct {
		name               string
		body               map[string]interface{}
		expectedStatusCode int
		expectedErr        bool
	}{
		{
			name: "Normal save 1",
			body: map[string]interface{}{
				"url":   gofakeit.URL(),
				"alias": gofakeit.Word() + "_" + gofakeit.Word(),
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name: "Normal save 2",
			body: map[string]interface{}{
				"url":   gofakeit.URL(),
				"alias": "",
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name: "Normal save 3",
			body: map[string]interface{}{
				"url":   gofakeit.URL(),
				"alias": "kaif",
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name: "Normal save 4",
			body: map[string]interface{}{
				"url":   gofakeit.URL(),
				"alias": "kaif2",
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name: "Alias already exists",
			body: map[string]interface{}{
				"url":   gofakeit.Word() + "_" + gofakeit.Word(),
				"alias": "leetcode",
			},
			expectedStatusCode: 400,
			expectedErr:        true,
		},
		{
			name: "Invalid JSON",
			body: map[string]interface{}{
				"url":   500,
				"alias": "",
			},
			expectedStatusCode: 400,
			expectedErr:        true,
		},
		{
			name: "Invalid url",
			body: map[string]interface{}{
				"url":   gofakeit.Word(),
				"alias": gofakeit.Word(),
			},
			expectedStatusCode: 400,
			expectedErr:        true,
		},
		{
			name:               "empty body",
			body:               map[string]interface{}{},
			expectedStatusCode: 400,
			expectedErr:        true,
		},
	}

	u := url.URL{
		Scheme: scheme,
		Host:   host,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(tt.body)
			assert.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(jsonBody))
			assert.NoError(t, err)

			req.SetBasicAuth("pasha", "1234")

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
}

func TestUrlShortener_Redirect(t *testing.T) {
	tests := []struct {
		name               string
		alias              string
		expectedStatusCode int
		expectedErr        bool
	}{
		{
			name:               "Normal Redirect 1",
			alias:              "hell",
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name:               "Normal Redirect 2",
			alias:              "jump2",
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name:               "Alias not found 1",
			alias:              gofakeit.Word(),
			expectedStatusCode: 400,
			expectedErr:        true,
		},
		{
			name:               "Alias not found 2",
			alias:              gofakeit.Word(),
			expectedStatusCode: 400,
			expectedErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := url.URL{
				Scheme: scheme,
				Host:   host,
				Path:   tt.alias,
			}

			req, err := http.NewRequest(http.MethodGet, u.String(), nil)
			assert.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			assert.NoError(t, err)

			assert.Equal(t, resp.StatusCode, tt.expectedStatusCode)
		})
	}
}

func TestUrlShortener_Update(t *testing.T) {
	tests := []struct {
		name               string
		body               map[string]interface{}
		expectedStatusCode int
		expectedErr        bool
	}{
		{
			name: "Normal update 1",
			body: map[string]interface{}{
				"alias":     "zzz123",
				"new_alias": "zzz",
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name: "Normal update 2",
			body: map[string]interface{}{
				"alias":     "zzz",
				"new_alias": "zzz123",
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name: "Alias does not exists",
			body: map[string]interface{}{
				"alias":     gofakeit.Word(),
				"new_alias": gofakeit.Word(),
			},
			expectedStatusCode: 400,
			expectedErr:        true,
		},
		{
			name: "NewAlias already exists",
			body: map[string]interface{}{
				"alias":     "jump2",
				"new_alias": "zzz123",
			},
			expectedStatusCode: 400,
			expectedErr:        true,
		},
	}

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

			req.SetBasicAuth("pasha", "1234")

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
}

func TestUrlShortener_Delete(t *testing.T) {
	tests := []struct {
		name               string
		body               map[string]interface{}
		expectedStatusCode int
		expectedErr        bool
	}{
		{
			name: "Normal delete 1",
			body: map[string]interface{}{
				"alias": "kaif",
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name: "Normal delete 2",
			body: map[string]interface{}{
				"alias": "kaif2",
			},
			expectedStatusCode: 200,
			expectedErr:        false,
		},
		{
			name: "Alias not found 1",
			body: map[string]interface{}{
				"alias": gofakeit.Word(),
			},
			expectedStatusCode: 400,
			expectedErr:        true,
		},
		{
			name: "Alias not found 2",
			body: map[string]interface{}{
				"alias": gofakeit.Word(),
			},
			expectedStatusCode: 400,
			expectedErr:        true,
		},
	}

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

			req.SetBasicAuth("pasha", "1234")

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			jsonData, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			var data map[string]interface{}
			assert.NoError(t, json.Unmarshal(jsonData, &data))

			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)
			if assert.Equal(t, tt.expectedErr, data["status"].(string) == "Error") {
				fmt.Println(data["error"].(string))
			}
		})
	}
}

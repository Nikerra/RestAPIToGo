package get_test

import (
	"RestApi/internal/http-server/handlers/url/get"
	"RestApi/internal/http-server/handlers/url/get/mocks"
	"RestApi/internal/storage"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetURLHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{
		{
			name:  "success",
			alias: "test_alias",
			url:   "https://google.com",
		},
		{
			name:      "Empty alias",
			alias:     "",
			respError: "field Alias is a required field",
		},
		{
			name:      "Not found",
			alias:     "test_bad_alias",
			respError: "url not found",
			mockError: storage.ErrURLNotFound,
		},
		{
			name:      "GetURL Error",
			alias:     "test_alias",
			respError: "failed to get url",
			mockError: errors.New("unexpected error"),
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlGetMock := mocks.NewURLGetter(t)

			if tc.respError == "" || tc.mockError != nil {
				urlGetMock.On(
					"GetURL", tc.alias).
					Return(tc.url, tc.mockError).
					Once()
			}

			handler := get.New(slog.New(
				slog.NewTextHandler(io.Discard, nil)), urlGetMock)
			input := fmt.Sprintf(`{"alias": "%s"}`, tc.alias)
			req, err := http.NewRequest(
				http.MethodPost, "/get-url", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			require.Equal(t, rr.Code, http.StatusOK)

			body := rr.Body.String()
			var resp get.Response
			require.NoError(t, json.Unmarshal([]byte(body), &resp))
			require.Equal(t, tc.respError, resp.Error)
		})
	}
}

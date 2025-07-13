package delete_test

import (
	"RestApi/internal/http-server/hadlers/url/delete"
	"RestApi/internal/http-server/hadlers/url/delete/mocks"
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

func TestDeleteURLHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		respError string
		mockError error
	}{
		{
			name:  "success",
			alias: "test_alias",
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
			name:      "DeleteURL Error",
			alias:     "test_alias",
			respError: "failed to delete url",
			mockError: errors.New("failed to delete url"),
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlDeleteMock := mocks.NewDeleteURL(t)

			if tc.respError == "" || tc.mockError != nil {
				urlDeleteMock.On(
					"DeleteURL", tc.alias).
					Return(tc.mockError).
					Once()
			}

			handler := delete.New(slog.New(
				slog.NewTextHandler(io.Discard, nil)), urlDeleteMock)
			input := fmt.Sprintf(`{"alias": "%s"}`, tc.alias)
			req, err := http.NewRequest(
				http.MethodDelete, "/delete-url", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			require.Equal(t, rr.Code, http.StatusOK)

			body := rr.Body.String()
			var resp delete.Response
			require.NoError(t, json.Unmarshal([]byte(body), &resp))
			require.Equal(t, tc.respError, resp.Error)
		})
	}
}

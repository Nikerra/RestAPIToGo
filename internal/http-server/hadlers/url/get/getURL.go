package get

import (
	resp "RestApi/internal/lib/api/response"
	"RestApi/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
)

type Request struct {
	Alias string `json:"alias" validate:"required"`
}

type Response struct {
	URL string `json:"url,omitempty"`
	resp.Response
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=URLGetter
type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, getter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.get.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", "error", err.Error())
			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))
		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)
			log.Error("invalid request", "error", err.Error())
			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		resUrl, err := getter.GetURL(req.Alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", slog.String("alias", req.Alias))
			render.JSON(w, r, resp.Error("url not found"))

			return
		}
		if err != nil {
			log.Error("failed to get url", "error", err.Error())
			render.JSON(w, r, resp.Error("failed to get url"))

			return
		}

		log.Info("url retrieved", slog.String("alias", req.Alias))

		render.JSON(w, r, Response{
			Response: resp.OK(),
			URL:      resUrl,
		})
	}
}

package delete

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"golang.org/x/exp/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"
)

//go:generate go run github.com/vektra/mockery/v2@v2.32.0 --name=URLGetter
type URLDelete interface {
	DeleteURL(alias string) (string, error)
}

func New(log *slog.Logger, urlDelete URLDelete) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")
			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		resURL, err := urlDelete.DeleteURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not deleted", "alias", alias)
			render.JSON(w, r, resp.Error("not deleted"))

			return
		}
		if err != nil {
			log.Error("failed to delete url", sl.Err(err))
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("delete url", slog.String("url", resURL))
		fmt.Println(http.StatusOK)
	}
}

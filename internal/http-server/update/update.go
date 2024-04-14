package update

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	httpServer "url-shortener/internal/http-server"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"

	"github.com/gin-gonic/gin"
)

type Request struct {
	Alias    string `json:"alias"`
	NewAlias string `json:"new_alias,omitempty"`
}

type Response struct {
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	NewAlias string `json:"new_alias,omitempty"`
}

type Decorator func(response *Response)

func SetStatus(status string) Decorator {
	return func(response *Response) {
		response.Status = status
	}
}

func SetError(error string) Decorator {
	return func(response *Response) {
		response.Error = error
	}
}

func SetNewAlias(newAlias string) Decorator {
	return func(response *Response) {
		response.NewAlias = newAlias
	}
}

func NewResponse(decorators ...Decorator) Response {
	var resp Response

	for _, d := range decorators {
		d(&resp)
	}

	return resp
}

func Update(log *slog.Logger, s storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "http-server.Update"
		var req Request

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Error(
				fmt.Sprintf("%s: %s", "failed to decode request", err.Error()),
				slog.String("op", op),
			)
			c.JSON(
				http.StatusBadRequest,
				NewResponse(
					SetStatus(httpServer.StatusError),
					SetError(httpServer.BadRequest),
				),
			)
			return
		}

		if req.NewAlias == "" {
			newAlias := random.Alias()
			if newAlias == "" {
				log.Error("failed to generate newAlias", slog.String("op", op))
				c.JSON(
					http.StatusInternalServerError,
					NewResponse(
						SetStatus(httpServer.StatusError),
						SetError(httpServer.InternalError),
					),
				)
				return
			}
			req.NewAlias = newAlias
		}

		username := c.GetString("username")

		log.Debug(
			"try to handle update request",
			slog.String("username", username),
			slog.String("alias", req.Alias),
			slog.String("newAlias", req.NewAlias),
			slog.String("op", op),
		)

		err := s.UpdateAlias(c, username, req.Alias, req.NewAlias)
		if err != nil {
			if errors.Is(err, storage.ErrCacheUpdate) {
				// failed to update alias in cache
				log.Error(err.Error(), slog.String("op", op))
				c.JSON(
					http.StatusOK,
					NewResponse(
						SetStatus(httpServer.StatusOK),
						SetNewAlias(httpServer.Path+req.NewAlias),
					),
				)
				return
			}
			if errors.Is(err, storage.ErrAliasNotFound) {
				log.Info("alias not found", slog.String("op", op))
				c.JSON(
					http.StatusBadRequest,
					NewResponse(
						SetStatus(httpServer.StatusError),
						SetError(httpServer.AliasNotFound),
					),
				)
				return
			}
			if errors.Is(err, storage.ErrNewAliasAlreadyExists) {
				log.Info("new alias already exists", slog.String("op", op))
				c.JSON(
					http.StatusBadRequest,
					NewResponse(
						SetStatus(httpServer.StatusError),
						SetError(httpServer.NewAliasAlreadyExists),
					),
				)
				return
			}
			log.Error(
				fmt.Sprintf("%s: %s", "failed to update alias by url", err.Error()),
				slog.String("op", op),
			)
			c.JSON(
				http.StatusBadRequest,
				NewResponse(
					SetStatus(httpServer.StatusError),
					SetError(httpServer.InternalError),
				),
			)
			return
		}

		log.Info(
			"success to update alias by url",
			slog.String("alias", req.Alias),
			slog.String("op", op),
		)
		c.JSON(
			http.StatusOK,
			NewResponse(
				SetStatus(httpServer.StatusOK),
				SetNewAlias(httpServer.Path+username+"/"+req.NewAlias),
			),
		)
	}
}

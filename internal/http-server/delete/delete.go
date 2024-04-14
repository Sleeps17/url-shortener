package delete

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	httpServer "url-shortener/internal/http-server"
	"url-shortener/internal/storage"

	"github.com/gin-gonic/gin"
)

type Request struct {
	Alias string `json:"alias" validate:"required"`
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type Decorator func(response *Response)

func SetStatus(status string) Decorator {
	return func(response *Response) {
		response.Status = status
	}
}

func SetError(err string) Decorator {
	return func(response *Response) {
		response.Error = err
	}
}

func NewResponse(decorators ...Decorator) Response {
	var resp Response

	for _, d := range decorators {
		d(&resp)
	}

	return resp
}

func Delete(log *slog.Logger, s storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "http-server.Delete"

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

		username := c.GetString("username")

		log.Debug(
			"try to handle delete request",
			slog.String("username", username),
			slog.String("alias", req.Alias),
			slog.String("op", op),
		)

		if err := s.DeleteURL(c, username, req.Alias); err != nil {
			if errors.Is(err, storage.ErrCacheDelete) {
				// failed to delete alias from cache
				log.Error(err.Error(), slog.String("op", op))
				c.JSON(
					http.StatusBadRequest,
					NewResponse(SetStatus(httpServer.StatusOK)),
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
			log.Error(err.Error(), slog.String("op", op))
			c.JSON(
				http.StatusInternalServerError,
				NewResponse(
					SetStatus(httpServer.StatusError),
					SetError(httpServer.InternalError),
				),
			)
			return
		}

		log.Info(
			"success handle delete url",
			slog.String("username", username),
			slog.String("alias", req.Alias),
			slog.String("op", op),
		)
		c.JSON(http.StatusOK, NewResponse(SetStatus(httpServer.StatusOK)))
	}
}

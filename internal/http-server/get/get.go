package get

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	httpServer "url-shortener/internal/http-server"
	"url-shortener/internal/storage"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Url    string `json:"url,omitempty"`
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

func Get(log *slog.Logger, s storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "http-server.Get"

		alias := c.Param("alias")
		username := c.Param("username")
		if alias == "" {
			log.Error("alias is empty", slog.String("op", op))
			c.JSON(
				http.StatusBadRequest,
				NewResponse(
					SetStatus(httpServer.StatusError),
					SetError(httpServer.BadRequest),
				),
			)
			return
		}

		if username == "" {
			log.Error("The username is missing", slog.String("op", op))
			c.JSON(
				http.StatusBadRequest,
				NewResponse(
					SetStatus(httpServer.StatusError),
					SetError(httpServer.BadRequest),
				),
			)
			return
		}

		log.Debug(
			"try to handle get request",
			slog.String("username", username),
			slog.String("alias", alias),
			slog.String("op", op),
		)

		url, err := s.GetURL(c, username, alias)
		if err != nil {
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
			} else if errors.Is(err, storage.ErrCacheGet) {
				log.Error(
					fmt.Sprintf("%s: %s", "failed to get url from cache", err.Error()),
					slog.String("op", op),
				)
			} else {
				log.Error(
					fmt.Sprintf("%s: %s", "failed to get url from storage", err.Error()),
					slog.String("op", op),
				)
				c.JSON(
					http.StatusInternalServerError,
					NewResponse(
						SetStatus(httpServer.StatusError),
						SetError(httpServer.InternalError),
					),
				)
				return
			}
		}

		log.Info(
			"success handle get url",
			slog.String("username", username),
			slog.String("alias", alias),
			slog.String("url", url),
			slog.String("op", op),
		)
		c.Redirect(http.StatusFound, url)
	}
}

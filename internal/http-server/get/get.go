package get

import (
	"errors"
	"github.com/gin-gonic/gin"
	"log/slog"
	httpServer "url-shortener/internal/http-server"
	"url-shortener/internal/storage"
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
		if alias == "" {
			log.Error("alias is empty", slog.String("op", op))
			c.JSON(400, NewResponse(SetStatus(httpServer.StatusError), SetError(httpServer.BadRequest)))
			return
		}

		url, err := s.GetURL(c, alias)
		if err != nil {
			if errors.Is(err, storage.ErrAliasNotFound) {
				c.JSON(400, NewResponse(SetStatus(httpServer.StatusError), SetError(httpServer.AliasNotFound)))
				return
			}
			log.Error(err.Error(), slog.String("op", op))
			c.JSON(500, NewResponse(SetStatus(httpServer.StatusError), SetError(httpServer.InternalError)))
			return
		}

		log.Info("Success handle get url", slog.String("op", op), slog.String("alias", alias))
		c.Redirect(302, url)
		return
	}
}

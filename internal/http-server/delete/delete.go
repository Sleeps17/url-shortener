package delete

import (
	"errors"
	"github.com/gin-gonic/gin"
	"log/slog"
	httpServer "url-shortener/internal/http-server"
	"url-shortener/internal/storage"
)

type Request struct {
	Alias string `json:"alias" validate:"required"`
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error"`
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
			log.Error(err.Error(), slog.String("op", op))
			c.JSON(400, NewResponse(SetStatus(httpServer.StatusError), SetError(httpServer.BadRequest)))
			return
		}

		if err := s.DeleteURL(c, req.Alias); err != nil {
			if errors.Is(err, storage.ErrAliasNotFound) {
				c.JSON(400, NewResponse(SetStatus(httpServer.StatusError), SetError(httpServer.AliasNotFound)))
				return
			}
			log.Error(err.Error(), slog.String("op", op))
			c.JSON(500, NewResponse(SetStatus(httpServer.StatusError), SetError(httpServer.InternalError)))
			return
		}

		c.JSON(200, NewResponse(SetStatus(httpServer.StatusOK)))
		log.Info("success handle delete url", slog.String("op", op), slog.String("alias", req.Alias))
	}
}

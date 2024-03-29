package update

import (
	"errors"
	"github.com/gin-gonic/gin"
	"log/slog"
	httpServer "url-shortener/internal/http-server"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"
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
			log.Error(err.Error(), slog.String("op", op))
			c.JSON(400, NewResponse(SetStatus(httpServer.StatusError), SetError(httpServer.BadRequest)))
			return
		}

		if req.NewAlias == "" {
			newAlias := random.Alias()
			if newAlias == "" {
				log.Error("failed to generate newAlias", slog.String("op", op))
				c.JSON(500, NewResponse(SetStatus(httpServer.StatusError), SetError(httpServer.InternalError)))
				return
			}
			req.NewAlias = newAlias
		}

		err := s.UpdateAlias(c, req.Alias, req.NewAlias)
		if err != nil {
			if errors.Is(err, storage.ErrCacheUpdate) {
				log.Error(err.Error(), slog.String("op", op))
				c.JSON(200, NewResponse(SetStatus(httpServer.StatusOK), SetNewAlias(httpServer.Path+req.NewAlias)))
				return
			}
			if errors.Is(err, storage.ErrAliasNotFound) {
				c.JSON(400, NewResponse(SetStatus(httpServer.StatusError), SetError(httpServer.AliasNotFound)))
				return
			}
			if errors.Is(err, storage.ErrNewAliasAlreadyExists) {
				c.JSON(400, NewResponse(SetStatus(httpServer.StatusError), SetError(httpServer.NewAliasAlreadyExists)))
				return
			}
			log.Error("failed to update alias by url", slog.String("op", op))
			c.JSON(500, NewResponse(SetStatus(httpServer.StatusError), SetError(httpServer.InternalError)))
		}

		c.JSON(200, NewResponse(SetStatus(httpServer.StatusOK), SetNewAlias(httpServer.Path+req.NewAlias)))
		log.Info("success to update alias by url", slog.String("op", op), slog.String("alias", req.Alias))
	}
}

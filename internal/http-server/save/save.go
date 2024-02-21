package save

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"log/slog"
	httpServer "url-shortener/internal/http-server"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"
)

type Request struct {
	Url   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Alias  string `json:"alias,omitempty"`
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

func SetAlias(alias string) Decorator {
	return func(response *Response) {
		response.Alias = alias
	}
}

func NewResponse(decorators ...Decorator) Response {
	var resp Response

	for _, d := range decorators {
		d(&resp)
	}

	return resp
}

func Save(log *slog.Logger, s storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "http-server.Save"

		var req Request
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Error(err.Error(), slog.String("op", op))
			c.JSON(400, NewResponse(SetStatus(httpServer.StatusError), SetError(httpServer.BadRequest)))
			return
		}

		if err := validator.New().Struct(req); err != nil {
			log.Error(err.Error(), slog.String("op", op))
			c.JSON(400, NewResponse(SetStatus(httpServer.StatusError), SetError(httpServer.BadRequest)))
			return
		}

		if req.Alias == "" {
			alias := random.Alias()
			if alias == "" {
				log.Error("failed to generate alias", slog.String("op", op))
				c.JSON(500, NewResponse(SetStatus(httpServer.StatusError), SetError(httpServer.InternalError)))
				return
			}
			req.Alias = alias
		}

		if err := s.SaveURL(c, req.Url, req.Alias); err != nil {
			if errors.Is(err, storage.ErrAliasAlreadyExist) {
				c.JSON(400, NewResponse(SetStatus(httpServer.StatusError), SetError(httpServer.AliasAlreadyExist)))
				return
			}
			log.Error(err.Error(), slog.String("op", op))
			c.JSON(500, NewResponse(SetStatus(httpServer.StatusError), SetError(httpServer.InternalError)))
			return
		}

		c.JSON(200, NewResponse(SetStatus(httpServer.StatusOK), SetAlias(httpServer.Path+req.Alias)))
		log.Info("success handle save url", slog.String("op", op), slog.String("alias", req.Alias))
	}
}

package save

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	httpServer "url-shortener/internal/http-server"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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

		if err := validator.New().Struct(req); err != nil {
			log.Error(
				fmt.Sprintf("%s: %s", "validation of url failed", err.Error()),
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

		if req.Alias == "" {
			alias := random.Alias()
			if alias == "" {
				log.Error("failed to generate alias", slog.String("op", op))
				c.JSON(
					http.StatusInternalServerError,
					NewResponse(
						SetStatus(httpServer.StatusError),
						SetError(httpServer.InternalError),
					),
				)
				return
			}
			req.Alias = alias
		}

		username := c.GetString("username")
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
			"try to handle save request",
			slog.String("username", username),
			slog.String("url", req.Url),
			slog.String("alias", req.Alias),
			slog.String("op", op),
		)

		if err := s.SaveURL(c, req.Url, req.Alias, username); err != nil {
			if errors.Is(err, storage.ErrCacheSet) {
				// failed to save url in cache
				log.Error(err.Error(), slog.String("op", op))
				c.JSON(
					http.StatusOK,
					NewResponse(
						SetStatus(httpServer.StatusOK),
						SetAlias(httpServer.Path+username+"/"+req.Alias),
					),
				)
				return
			}
			if errors.Is(err, storage.ErrAliasAlreadyExist) {
				log.Info(
					fmt.Sprintf("%s", "alias already exist"),
					slog.String("op", op),
				)
				c.JSON(
					http.StatusBadRequest,
					NewResponse(
						SetStatus(httpServer.StatusError),
						SetError(httpServer.AliasAlreadyExist),
					),
				)
				return
			}

			log.Error(
				fmt.Sprintf("%s: %s", "failed to handle save request", err.Error()),
				slog.String("op", op),
			)
			c.JSON(
				http.StatusBadRequest,
				NewResponse(SetStatus(httpServer.StatusError),
					SetError(httpServer.InternalError),
				),
			)
			return
		}

		log.Info(
			"success handle save url",
			slog.String("username", username),
			slog.String("alias", req.Alias),
			slog.String("op", op),
		)
		c.JSON(
			http.StatusOK,
			NewResponse(
				SetStatus(httpServer.StatusOK),
				SetAlias(httpServer.Path+username+"/"+req.Alias),
			),
		)
	}
}

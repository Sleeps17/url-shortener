package suite

import (
	"context"
	"net/http"
	"testing"
	"url-shortener/internal/config"
)

type Suite struct {
	*testing.T
	Cfg    *config.Config
	Client *http.Client
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()

	cfg := config.MustLoadByPath("../config/dev.yaml")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HttpServer.Timeout)

	t.Cleanup(func() {
		t.Helper()
		cancel()
	})

	return ctx, &Suite{Cfg: cfg, Client: http.DefaultClient}
}

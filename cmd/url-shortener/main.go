package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/exp/slog"
	"net/http"
	"os"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/url/delete"
	"url-shortener/internal/http-server/handlers/url/save"
	mwLogger "url-shortener/internal/http-server/middleware/logger"
	"url-shortener/internal/lib/logger/handlers/slogpretty"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage/sqlite"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	//init config - cleanenv
	cfg := config.MustLoad()
	fmt.Println(cfg)
	//init logger - slog
	log := setupLogger(cfg.Env)
	log.Info("starting url-shortener", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")
	//init storage - sqlite
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1) // or return
	}

	_ = storage

	//todo init router - chi, chi render
	//chi это маршрутизатор для создания сервисов Go HTTP
	router := chi.NewRouter()
	//todo middleware
	//запросу присваивается уникальный id(в grafana/kibana или через grep будет легко найти все соотв-ие записи по req-id)
	router.Use(middleware.RequestID)
	//ip юзера
	router.Use(middleware.RealIP)
	// внутренний логер мидлвэра(можно испльзовать опционально)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	// вспомогательные мидлвари
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	//auth
	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))
		r.Post("/", save.New(log, storage))
		//todo
		router.Delete("/{alias}", delete.New(log, storage))
	})

	router.Get("/{alias}", redirect.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}
	//run server
	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}
	log.Error("server stopped")

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = setupPrettySlog()
		//log = slog.New(
		//	slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		//)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}

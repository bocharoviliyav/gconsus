package handler

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-playground/validator/v10"

	"gconsus/lib/http/middleware"
	"gconsus/service"
)

type Config struct {
	Port           int   `validate:"required"`
	RequestSizeMax int64 `validate:"required,min=1"`
	ReadTimeout    int   `validate:"required,min=1,max=10"`
	WriteTimeout   int   `validate:"required,min=1,max=10"`
	IdleTimeout    int
}

func DefaultConfig() Config {
	return Config{
		Port:           8000,
		RequestSizeMax: 1 * 1024 * 1024,
		ReadTimeout:    1,
		WriteTimeout:   5,
		IdleTimeout:    15,
	}
}

func New(cfg Config, projectsService service.Service) (*http.Server, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("invalid config, %w", err)
	}

	// Enable HTTP/2
	var protocols http.Protocols

	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", rootPage)

	// NO trailing slash allowed with StripSlashes middleware!
	mux.HandleFunc("GET /api/v1/users/{login}/activity", v1UserActivity(projectsService))

	wrapped := middleware.Use(
		mux,
		middleware.StripSlashes,
		middleware.RequestSizeMax(cfg.RequestSizeMax),
		middleware.SetHeader("Content-Type", "application/json"),
		middleware.Recoverer,
	)

	srv := http.Server{
		Addr:         fmt.Sprintf(":%v", cfg.Port),
		Handler:      wrapped,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
		Protocols:    &protocols,
	}

	return &srv, nil
}

func rootPage(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)

	version := "<unknown version>"
	buildInfo, ok := debug.ReadBuildInfo()
	gitCommit := ""
	buildDate := ""

	if ok {
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				gitCommit = setting.Value
			case "vcs.time":
				buildDate = setting.Value
			}
		}

		if gitCommit != "" && buildDate != "" {
			version = gitCommit + " / " + buildDate
		}
	}

	_, _ = w.Write([]byte("Cloud.ru Folders service " + version + "\nfor api go to /api/v1"))
}

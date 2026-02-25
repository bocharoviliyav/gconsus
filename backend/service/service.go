package service

import (
	"gconsus/adapter/github"

	"github.com/go-playground/validator/v10"
)

type Service struct {
	ghClient github.Client
	validate *validator.Validate
}

func New(ghClient github.Client) (Service, error) {
	srv := Service{
		ghClient: ghClient,
		validate: validator.New(validator.WithRequiredStructEnabled()),
	}

	return srv, nil
}

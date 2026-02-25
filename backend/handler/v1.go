package handler

import (
	"errors"
	"net/http"
	"time"

	"gconsus/lib/http/rest"
	"gconsus/service"
)

func v1UserActivity(pService service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		login := r.PathValue("login")
		fromStr := r.URL.Query().Get("from")

		if fromStr == "" {
			rest.ReturnRequestError(w, "must be from timestamp param")

			return
		}

		toStr := r.URL.Query().Get("to")
		if toStr == "" {
			rest.ReturnRequestError(w, "must be to timestamp param")

			return
		}

		from, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			rest.ReturnRequestError(w, "invalid from timestamp param")

			return
		}

		to, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			rest.ReturnRequestError(w, "invalid to timestamp param")

			return
		}

		response, err := pService.UserActivity(r.Context(), login, from, to)
		if err != nil {
			var pErr service.ParamsError

			switch {
			case errors.As(err, &pErr):
				rest.ReturnRequestError(w, err.Error())
			default:
				rest.ReturnServerError(w)
			}

			return
		}

		rest.ReturnResponse(w, response)
	}
}

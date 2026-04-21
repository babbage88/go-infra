package cert_renew

import (
	"net/http"

	corecert "github.com/babbage88/infra-core/cert_renew"
)

func Renewcert_renew() http.HandlerFunc {
	return corecert.Renewcert_renew()
}

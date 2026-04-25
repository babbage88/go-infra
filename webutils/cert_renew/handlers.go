package cert_renew

import (
	"net/http"

	corecert "github.com/babbage88/infra-core/cert_renew"
)

// swagger:route POST /renew Certificates Renew
// Request/Renew ssl certificate via cloudflare letsencrypt. Uses DNS Challenge
// responses:
//  200: CertificateDataRenewResponse
//  400: description:Bad Request
//  401: description:Unauthorized
//  500: description:Insernal Server Error
// produces:
// - application/json
// - application/zip
func Renewcert_renew() http.HandlerFunc {
	return corecert.Renewcert_renew()
}

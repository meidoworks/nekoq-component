package secretserver

import (
	"net/http"
	"strings"

	"github.com/meidoworks/nekoq-component/configure/secretaddon"
	"github.com/meidoworks/nekoq-component/configure/secretapi"
	"github.com/meidoworks/nekoq-component/http/chi2"
)

func ValidateJwtToken(verifier secretapi.JwtVerifier, allowed secretaddon.PermissionResourceList) func(w http.ResponseWriter, r *http.Request) chi2.Render {
	return func(w http.ResponseWriter, r *http.Request) chi2.Render {
		auth := strings.TrimSpace(r.Header.Get("Authorization"))
		auths := strings.Split(auth, " ")
		if len(auths) != 2 {
			return chi2.NewStatusRender(http.StatusUnauthorized)
		}
		if strings.TrimSpace(auths[0]) != "Bearer" {
			return chi2.NewStatusRender(http.StatusUnauthorized)
		}
		token := strings.TrimSpace(auths[1])

		tool := secretaddon.NewJwtTool(verifier)
		pass, err := tool.VerifyPermissionsOnJwtToken(token, allowed, secretaddon.AnyPermissionOperator)
		if err != nil {
			//FIXME should log error?
			return chi2.NewStatusRender(http.StatusUnauthorized)
		} else if !pass {
			return chi2.NewStatusRender(http.StatusUnauthorized)
		} else {
			return nil
		}
	}
}

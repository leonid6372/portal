package oauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	resp "portal/internal/lib/api/response"
	"reflect"
	"time"

	"github.com/go-chi/render"
)

type contextKey string

const (
	CredentialContext  contextKey = "oauth.credential"
	ClaimsContext      contextKey = "oauth.claims"
	ScopeContext       contextKey = "oauth.scope"
	AccessTokenContext contextKey = "oauth.accesstoken"
)

// BearerAuthentication middleware for go-chi
type BearerAuthentication struct {
	secretKey    string
	provider     *TokenProvider
	BearerServer *BearerServer
}

// NewBearerAuthentication create a BearerAuthentication middleware
func NewBearerAuthentication(secretKey string, formatter TokenSecureFormatter, bs *BearerServer) *BearerAuthentication {
	ba := &BearerAuthentication{secretKey: secretKey, BearerServer: bs}
	if formatter == nil {
		formatter = NewSHA256RC4TokenSecurityProvider([]byte(secretKey))
	}
	ba.provider = NewTokenProvider(formatter)
	return ba
}

// Authorize is the OAuth 2.0 middleware for go-chi resource server.
// Authorize creates a BearerAuthentication middleware and return the Authorize method.
func Authorize(secretKey string, formatter TokenSecureFormatter, bs *BearerServer) func(next http.Handler) http.Handler {
	return NewBearerAuthentication(secretKey, formatter, bs).Authorize
}

// Authorize verifies the bearer token authorizing or not the request.
// Token is retrieved from the Authorization HTTP header that respects the format
// Authorization: Bearer {access_token}
func (ba *BearerAuthentication) Authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				http.Error(w, "cookie not found", http.StatusBadRequest)
			default:
				http.Error(w, "server error", http.StatusInternalServerError)
			}
			return
		}
		auth := string([]byte(cookie.Value))

		token, err := ba.checkAuthorization(auth, w, r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, "Not authorized: "+err.Error())
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, CredentialContext, token.Credential)
		ctx = context.WithValue(ctx, ClaimsContext, token.Claims)
		ctx = context.WithValue(ctx, ScopeContext, token.Scope)
		ctx = context.WithValue(ctx, AccessTokenContext, auth)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Check header and token.
func (ba *BearerAuthentication) checkAuthorization(auth string, w http.ResponseWriter, r *http.Request) (t *Token, err error) {
	token, err := ba.provider.DecryptToken(auth)
	if err != nil {
		return nil, errors.New("Invalid token")
	}
	if time.Now().UTC().After(token.CreationDate.Add(token.ExpiresIn)) {
		scope := r.FormValue("scope")
		cookie, cookieErr := r.Cookie("refresh_token")
		if cookieErr != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				http.Error(w, "cookie not found", http.StatusBadRequest)
			default:
				http.Error(w, "server error", http.StatusInternalServerError)
			}
			return nil, fmt.Errorf("Not authorized")
		}
		refreshToken := string([]byte(cookie.Value))
		response, statusCode := ba.BearerServer.generateTokenResponse(GrantType("refresh_token"), "", "", refreshToken, scope, "", "", r)

		if statusCode != 200 {
			w.WriteHeader(statusCode)
			render.JSON(w, r, resp.Error(response))
			return
		}

		http.SetCookie(w,
			&http.Cookie{
				Name:  "access_token",
				Value: reflect.Indirect(reflect.ValueOf(response)).FieldByName("Token").String(),
			})

		http.SetCookie(w,
			&http.Cookie{
				Name:  "refresh_token",
				Value: reflect.Indirect(reflect.ValueOf(response)).FieldByName("RefreshToken").String(),
			})
	}
	return token, nil
}

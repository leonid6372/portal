package oauth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"time"

	resp "portal/internal/lib/api/response"
	"portal/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
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
	Log          *slog.Logger
}

// NewBearerAuthentication create a BearerAuthentication middleware
func NewBearerAuthentication(secretKey string, formatter TokenSecureFormatter, bs *BearerServer, log *slog.Logger) *BearerAuthentication {
	ba := &BearerAuthentication{secretKey: secretKey, BearerServer: bs, Log: log}
	if formatter == nil {
		formatter = NewSHA256RC4TokenSecurityProvider([]byte(secretKey))
	}
	ba.provider = NewTokenProvider(formatter)
	return ba
}

// Authorize is the OAuth 2.0 middleware for go-chi resource server.
// Authorize creates a BearerAuthentication middleware and return the Authorize method.
func Authorize(secretKey string, formatter TokenSecureFormatter, bs *BearerServer, log *slog.Logger) func(next http.Handler) http.Handler {
	return NewBearerAuthentication(secretKey, formatter, bs, log).Authorize
}

// Authorize verifies the bearer token authorizing or not the request.
// Token is retrieved from the Authorization HTTP header that respects the format
// Authorization: Bearer {access_token}
func (ba *BearerAuthentication) Authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const op = "lib.oauth.Authorize"

		log := ba.Log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		cookie, err := r.Cookie("access_token")
		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				log.Error("cookie not found")
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, resp.Error("cookie not found"))
			default:
				log.Error("server error", sl.Err(err))
				w.WriteHeader(http.StatusInternalServerError)
				render.JSON(w, r, resp.Error("server error"))
			}
			return
		}
		auth := cookie.Value

		token, err := ba.checkAuthorization(auth, w, r)
		if err != nil {
			log.Error("Not authorized", sl.Err(err))
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, resp.Error("Not authorized"))
			return
		}
		fmt.Println(token.Claims)
		ctx := r.Context()
		ctx = context.WithValue(ctx, CredentialContext, token.Credential)
		ctx = context.WithValue(ctx, ClaimsContext, token.Claims)
		ctx = context.WithValue(ctx, ScopeContext, token.Scope)
		ctx = context.WithValue(ctx, AccessTokenContext, auth)

		log.Info("authorization success")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Check header and token.
func (ba *BearerAuthentication) checkAuthorization(auth string, w http.ResponseWriter, r *http.Request) (t *Token, err error) {
	token, err := ba.provider.DecryptToken(auth)
	if err != nil {
		return nil, errors.New("Invalid token: " + err.Error())
	}

	if time.Now().UTC().After(token.CreationDate.Add(token.ExpiresIn)) {
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				http.Error(w, "cookie not found", http.StatusBadRequest)
			default:
				http.Error(w, "server error", http.StatusInternalServerError)
			}
			return nil, fmt.Errorf("Not authorized: " + err.Error())
		}
		refreshToken := cookie.Value
		response, statusCode := ba.BearerServer.generateTokenResponse(GrantType("refresh_token"), "", "", refreshToken, r)

		if statusCode != 200 {
			return nil, errors.New("Error while token generating: " + reflect.ValueOf(response).String())
		}

		http.SetCookie(w,
			&http.Cookie{
				Name:     "access_token",
				Value:    reflect.Indirect(reflect.ValueOf(response)).FieldByName("Token").String(),
				Expires:  time.Now().Add(2160 * time.Hour), // Время зачитски куки из браузера (ttl находитися в token.ExpiresIn)
				HttpOnly: true,
			})

		http.SetCookie(w,
			&http.Cookie{
				Name:     "refresh_token",
				Value:    reflect.Indirect(reflect.ValueOf(response)).FieldByName("RefreshToken").String(),
				Expires:  time.Now().Add(2160 * time.Hour),
				HttpOnly: true,
			})
	}
	return token, nil

}

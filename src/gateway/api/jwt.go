package api

import (
	"context"
	"fmt"
	"gateway/errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type JWTClaimsContextType int

const JWTClaimsContextKey JWTClaimsContextType = 0

func (a *API) CreateToken(aud string) (*string, error) {
	if gw, err := a.Gateway(); err != nil {
		return nil, err
	} else if gw == nil || len(gw.Certificates) == 0 || gw.PrivateKey == nil {
		return nil, fmt.Errorf("cannot create jwt: no gateway certificates")
	} else {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"gateway": gw.ID,
			"aud":     aud,
			"nbf":     time.Now().Unix(),
		})

		if t, err := token.SignedString(gw.PrivateKey); err != nil {
			return nil, err
		} else {
			return &t, nil
		}
	}
}

func (a *API) CreateTokenBLE(in map[string]interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{})
	if t, err := a.CreateToken("app"); err != nil {
		return nil, err
	} else {
		out["token"] = t
	}
	return out, nil
}

func (a *API) VerifyTokenMiddleware(w http.ResponseWriter, r *http.Request) *http.Request {
	var ah []string

	var tokenCookie *http.Cookie
	for _, cookie := range r.Cookies() {
		if cookie.Name == "token" {
			tokenCookie = cookie
			break
		}
	}

	var formToken string
	if err := r.ParseForm(); err == nil {
		formToken = r.PostFormValue("__authToken")
	}

	if header, ok := r.Header["Authorization"]; ok {
		ah = header
	} else if header, ok := r.Header["X-Authorization"]; ok {
		ah = header
	} else if tokenCookie != nil {
		ah = []string{fmt.Sprintf("Bearer %s", tokenCookie.Value)}
	} else if formToken != "" {
		ah = []string{fmt.Sprintf("Bearer %s", formToken)}
	} else if _, ok := r.Header["Upgrade"]; ok {
		if r.Header["Upgrade"][0] == "websocket" {
			ah = []string{fmt.Sprintf("Bearer %s", strings.TrimPrefix(r.URL.Path, "/"))}
		}
	}

	if ah == nil {
		log.Println(r.Header)
		errors.NewUnauthorized("must provide bearer token").Write(w)
		return nil
	}

	if len(ah) != 1 {
		errors.NewUnauthorized("must provide bearer token").Write(w)
		return nil
	} else if !strings.HasPrefix(strings.ToLower(ah[0]), "bearer ") {
		errors.NewUnauthorized("must provide bearer token").Write(w)
		return nil
	}

	tokenString := strings.Trim(strings.SplitN(ah[0], " ", 2)[1], " ")

	if gw, err := a.Gateway(); err != nil {
		errors.NewInternalServerError("could not load gateway: %+v", err).Println().Write(w)
		return nil
	} else if gw == nil {
		errors.NewInternalServerError("could not load gateway: nil").Println().Write(w)
		return nil
	} else if publicKey := gw.PublicKey(); publicKey == nil {
		errors.NewInternalServerError("no public key for gateway").Println().Write(w)
		return nil
	} else if token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return publicKey, nil
	}); err != nil {
		errors.NewUnauthorized("invalid bearer token: %+v", err).Write(w)
		return nil
	} else if claims, ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
		errors.NewUnauthorized("invalid bearer token").Write(w)
		return nil
	} else {
		ctx := context.WithValue(r.Context(), JWTClaimsContextKey, claims)
		return r.WithContext(ctx)
	}
}

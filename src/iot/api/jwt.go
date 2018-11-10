package api

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"strings"
	"time"
)

type JWTClaimsContextType int

const JWTClaimsContextKey JWTClaimsContextType = 0

func (a *API) CreateToken() (*string, error) {
	if gw, err := a.Gateway(); err != nil {
		return nil, err
	} else if gw == nil || len(gw.Certificates) == 0 || gw.PrivateKey == nil {
		return nil, fmt.Errorf("cannot create jwt: no gateway certificates")
	} else {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"gateway": gw.ID,
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
	if t, err := a.CreateToken(); err != nil {
		return nil, err
	} else {
		out["token"] = t
	}
	return out, nil
}

func (a *API) VerifyTokenMiddleware(w http.ResponseWriter, r *http.Request) *http.Request {
	if !strings.HasPrefix(strings.ToLower(r.Header["Authorization"][0]), "bearer ") {
		a.Error(w, fmt.Errorf("must provide bearer token"), http.StatusUnauthorized)
		return nil
	}

	tokenString := strings.Trim(strings.SplitN(r.Header["Authorization"][0], " ", 2)[1], " ")

	if gw, err := a.Gateway(); err != nil {
		a.Error(w, fmt.Errorf("could not load gateway: %+v", err), http.StatusUnauthorized)
		return nil
	} else if gw == nil {
		a.Error(w, fmt.Errorf("could not load gateway: nil"), http.StatusUnauthorized)
		return nil
	} else if publicKey := gw.PublicKey(); publicKey == nil {
		a.Error(w, fmt.Errorf("no public key for gateway"), http.StatusUnauthorized)
		return nil
	} else if token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return publicKey, nil
	}); err != nil {
		a.Error(w, fmt.Errorf("invalid bearer token: %+v", err), http.StatusUnauthorized)
		return nil
	} else if claims, ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
		a.Error(w, fmt.Errorf("invalid bearer token"), http.StatusUnauthorized)
		return nil
	} else {
		ctx := context.WithValue(r.Context(), JWTClaimsContextKey, claims)
		return r.WithContext(ctx)
	}
}

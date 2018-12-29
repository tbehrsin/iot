package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"gateway/errors"
	"math/big"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type CreateEmailTokenRequest struct {
	Email string `json:"email"`
}

type CreateEmailTokenServerRequest struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

func (a *API) CreateEmailTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateEmailTokenRequest

	defer r.Body.Close()

	if b, err := ioutil.ReadAll(r.Body); err != nil {
		errors.NewBadRequest(err).Println().Write(w)
	} else if err := json.Unmarshal(b, &req); err != nil {
		errors.NewBadRequest(err).Println().Write(w)
	} else if token, err := a.CreateToken("email"); err != nil {
		errors.NewInternalServerError(err).Println().Write(w)
	} else if b, err := json.Marshal(CreateEmailTokenServerRequest{*token, req.Email}); err != nil {
		errors.NewInternalServerError(err).Println().Write(w)
	} else if req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/auth/", Server), bytes.NewReader(b)); err != nil {
		errors.NewInternalServerError(err).Println().Write(w)
	} else if gw, err := a.Gateway(); err != nil {
		errors.NewBadRequest(err).Println().Write(w)
	} else if gw == nil {
		errors.NewBadRequest("no gateway created").Println().Write(w)
	} else if httpClient := gw.HTTPClient(); httpClient == nil {
		errors.NewInternalServerError("failed to create https client").Println().Write(w)
	} else {
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Cache-Control", "no-cache")

		if res, err := httpClient.Do(req); err != nil {
			errors.NewInternalServerError(err).Println().Write(w)
		} else {
			defer res.Body.Close()

			if b, err := ioutil.ReadAll(res.Body); err != nil {
				errors.NewInternalServerError(err).Println().Write(w)
			} else {
				w.WriteHeader(res.StatusCode)
				w.Write(b)
			}
		}
	}
}

type CreateTokenResponse struct {
	Token string `json:"token"`
}

func (a *API) CreateCLITokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), JWTClaimsContextKey, jwt.MapClaims{"aud": "cli"})
	a.ConvertToken(w, r.WithContext(ctx))
}

type ConvertTokenRequest struct {
	Code string `json:"code"`
}

type ConvertTokenResponse struct {
	Token string `json:"token"`
}

var tokenConversions = map[string]string{
	"email": "app",
	"cli":   "developer",
}

func (a *API) ConvertToken(w http.ResponseWriter, r *http.Request) {
	var req ConvertTokenRequest

	claims := r.Context().Value(JWTClaimsContextKey).(jwt.MapClaims)
	if newClaim, ok := tokenConversions[claims["aud"].(string)]; !ok {
		errors.NewForbidden().Write(w)
		return
	} else if b, err := ioutil.ReadAll(r.Body); err != nil {
		errors.NewBadRequest(err).Println().Write(w)
	} else if err := json.Unmarshal(b, &req); err != nil {
		errors.NewBadRequest(err).Println().Write(w)
	} else if gw, err := a.Gateway(); err != nil {
		errors.NewBadRequest(err).Println().Write(w)
	} else if gw == nil {
		errors.NewBadRequest("no gateway created").Println().Write(w)
	} else if !gw.ValidateAuthCode(req.Code) {
		errors.NewForbidden("invalid auth code").Write(w)
	} else if token, err := a.CreateToken(newClaim); err != nil {
		errors.NewInternalServerError(err).Println().Write(w)
	} else {
		errors.APIJSON(w, ConvertTokenResponse{*token})
	}
}

type AuthCode struct {
	Code   string    `json:"code"`
	Issued time.Time `json:"issued"`
}

func generateAuthCode() (*AuthCode, error) {
	alphabet := "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	out := ""
	for i := 0; i < 6; i++ {
		if n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet)))); err != nil {
			return nil, err
		} else {
			out += string(alphabet[int(n.Int64())])
		}
	}
	return &AuthCode{
		out,
		time.Now(),
	}, nil
}

func (gw *Gateway) ValidateAuthCode(code string) bool {
	for _, authCode := range gw.AuthCodes {
		if time.Now().Sub(authCode.Issued) > time.Duration(60)*time.Second {
			continue
		}

		if authCode.Code == code {
			return true
		}
	}
	return false
}

func (gw *Gateway) PruneAuthCodes() {
	deleted := 0
	for i, authCode := range gw.AuthCodes {
		j := i - deleted
		if time.Now().Sub(authCode.Issued) > time.Duration(60)*time.Second {
			deleted++
			if len(gw.AuthCodes) == j+1 {
				gw.AuthCodes = gw.AuthCodes[:j]
			} else {
				gw.AuthCodes = append(gw.AuthCodes[:j], gw.AuthCodes[j+1:]...)
			}
		}
	}
}

func (a *API) CreateAuthCode(w http.ResponseWriter, r *http.Request) {
	if gw, err := a.Gateway(); err != nil {
		errors.NewBadRequest(err).Println().Write(w)
	} else if gw == nil {
		errors.NewBadRequest("no gateway created").Println().Write(w)
	} else if authCode, err := generateAuthCode(); err != nil {
		errors.NewInternalServerError(err).Println().Write(w)
	} else {
		gw.PruneAuthCodes()
		gw.AuthCodes = append(gw.AuthCodes, *authCode)
		if err := gw.Update(); err != nil {
			errors.NewInternalServerError(err).Println().Write(w)
		} else {
			errors.APIJSON(w, *authCode)
		}
	}
}

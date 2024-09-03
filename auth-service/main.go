package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cast"
)

type JWT struct {
	AccessTokenRegisterClaims  *jwt.RegisteredClaims `json:"access_token_cliams"`
	RefreshTokenRegisterClaims *jwt.RegisteredClaims `json:"refresh_token_cliams"`
	AccessToken                string                `json:"access_token"`
	RefreshToken               string                `json:"refresh_token"`
}

var (
	keycloakURL  = "http://keycloak.localhost"
	client       = resty.New().SetBaseURL("http://keycloak:8080").SetDebug(true)
	realm        = "demo-realm"
	clientID     = "demo-client"
	clientSecret = "gjST5sguY1SqOAgO5srxw9XMy7gx7kyH"
	redirectURI  = "http://localhost/auth/callback"
	publicKeyStr = "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAsJFpPFMJJEcWS6gMMFFoLIPT444+SqlvTTnGefAzPU4fp201TSMrwI+25SgdqsvwTEepaqHRoCQzlRzru726s6vBWE6+6zsoOJVxNoDHcBbWVRosbsEb5g3jAoLhFjPRylrfNIR2u1qN60IOKcTG8vXzlepVSnUP9YWbiV/3SXm76QrIkFp66vSDLNFgDG5l0HEkLmQzRKk15lz5o+P2LF3M5D9CvPDoNXzj/UXTbGB6AqhEOF8a0OLz+WRCTTd9KIHLNMGjER017a95zutXZ1YHunUugycx05EX4T738Di7HIH3NK7lSbpWC57IqJYGSEov7bb46uTZY13ohqEa5wIDAQAB"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Route to start the OAuth 2.0 flow
	e.GET("/login", func(c echo.Context) error {
		authURL, err := url.Parse(keycloakURL)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		authURL.Path = fmt.Sprintf("realms/%s/protocol/openid-connect/auth", realm)
		authURL.ForceQuery = true
		queryparms := authURL.Query()
		queryparms.Add("client_id", clientID)
		queryparms.Add("redirect_uri", redirectURI)
		queryparms.Add("response_type", "code")
		queryparms.Add("scope", "openid profile email")
		rawQuery, err := url.QueryUnescape(queryparms.Encode())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		authURL.RawQuery = rawQuery

		return c.Redirect(http.StatusTemporaryRedirect, authURL.String())
	})

	// Route to handle callback from Keycloak
	e.GET("/callback", func(c echo.Context) error {
		code := c.QueryParam("code")
		if code == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Authorization code not provided"})
		}

		ctx := c.Request().Context()

		// Exchange the authorization code for an access token
		path := fmt.Sprintf("realms/%s/protocol/openid-connect/token", realm)
		body := map[string]string{
			"grant_type":    "authorization_code",
			"client_id":     clientID,
			"client_secret": clientSecret,
			"code":          code,
			"redirect_uri":  redirectURI,
		}

		resp, err := client.R().
			SetContext(ctx).
			SetFormData(body).
			Post(path)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		respBody := map[string]interface{}{}

		if err := json.Unmarshal(resp.Body(), &respBody); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		if resp.StatusCode() >= 400 {
			return echo.NewHTTPError(http.StatusInternalServerError, cast.ToString(respBody["error"]))
		}

		token := JWT{
			AccessToken:  cast.ToString(respBody["access_token"]),
			RefreshToken: cast.ToString(respBody["refresh_token"]),
			AccessTokenRegisterClaims: &jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Unix(cast.ToInt64(respBody["expires_in"]), 0)),
			},
			RefreshTokenRegisterClaims: &jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Unix(cast.ToInt64(respBody["refresh_expires_in"]), 0)),
			},
		}

		return c.JSON(http.StatusOK, token)
	})

	// A protected route that requires a valid JWT
	e.GET("/auth", func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing Authorization header"})
		}

		token := authHeader[len("Bearer "):]

		// ctx := c.Request().Context()

		// // Introspect the token
		// path := fmt.Sprintf("realms/%s/protocol/openid-connect/token/introspect", realm)
		// body := map[string]string{
		// 	"client_id":       clientID,
		// 	"client_secret":   clientSecret,
		// 	"token":           token,
		// 	"token_type_hint": "requesting_party_token",
		// }
		//
		// resp, err := client.R().
		// 	SetContext(ctx).
		// 	SetFormData(body).
		// 	Post(path)
		// if err != nil {
		// 	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		// }
		//
		// respBody := map[string]interface{}{}
		//
		// json.Unmarshal(resp.Body(), &respBody)
		//
		// if resp.StatusCode() >= 400 {
		// 	return echo.NewHTTPError(http.StatusInternalServerError, cast.ToString(respBody["error"]))
		// }
		//
		// if !cast.ToBool(respBody["active"]) {
		// 	return echo.NewHTTPError(http.StatusUnauthorized, "Access Token is not active")
		// }

		// Verify and parse the JWT
		claims := jwt.MapClaims{}
		if _, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, errors.New("Unexpected signing method: " + token.Header["alg"].(string))
			}

			// verify exp
			registerCliams := token.Claims.(jwt.MapClaims)
			if registerCliams["exp"].(float64) <= float64(time.Now().Unix()) {
				return nil, errors.New("Access Token is Expired")
			}

			buf, err := base64.StdEncoding.DecodeString(publicKeyStr)
			if err != nil {
				return nil, err
			}
			parsedKey, err := x509.ParsePKIXPublicKey(buf)
			if err != nil {
				return nil, err
			}
			publicKey, ok := parsedKey.(*rsa.PublicKey)
			if !ok {
				return nil, errors.New("Invalid public key")
			}
			return publicKey, nil
		}); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		username := cast.ToString(claims["preferred_username"])
		userID := cast.ToString(claims["sub"])

		c.Response().Header().Set("X-Auth-User", username)
		c.Response().Header().Set("X-Auth-User-ID", userID)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Authenticated",
		})
	})

	e.Logger.Fatal(e.Start(":8001"))
}

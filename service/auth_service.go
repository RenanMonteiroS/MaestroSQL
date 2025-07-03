package service

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/RenanMonteiroS/MaestroSQLWeb/model"
)

type AuthService struct{}

func NewAuthService() AuthService {
	return AuthService{}
}

func (as *AuthService) GenerateStateOAuthCookie() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (ds *AuthService) GetOAuth2UserInfo(accessToken string, method string) (model.OAuth2User, error) {

	var (
		userContent model.OAuth2User
		res         *http.Response
		req         *http.Request
		err         error
	)

	client := http.Client{}
	slog.Info("Searching for user information", "OAuth2 Method", method)
	if method == "google" {
		req, err = http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo?access_token="+accessToken, nil)
		if err != nil {
			slog.Error("Cannot wrap request", "URL", "https://www.googleapis.com/oauth2/v2/userinfo", "HTTP Verb", "GET", "Error", err)
			return model.OAuth2User{}, err
		}

		res, err = client.Do(req)
		if err != nil {
			slog.Error("Cannot get user information while fetching URL", "URL", "https://www.googleapis.com/oauth2/v2/userinfo", "HTTP Verb", "GET", "Error", err)
			return model.OAuth2User{}, err
		}

		err = json.NewDecoder(res.Body).Decode(&userContent.GoogleOAuth2User)
		if err != nil {
			slog.Error("Cannot decode JSON information", "Error", err)
			return model.OAuth2User{}, err
		}

	} else if method == "microsoft" {
		req, err = http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me", nil)
		if err != nil {
			slog.Error("Cannot wrap request", "URL", "https://graph.microsoft.com/v1.0/me", "HTTP Verb", "GET", "Error", err)
			return model.OAuth2User{}, err
		}
		req.Header = http.Header{
			"Authorization": {"Bearer " + accessToken},
		}

		res, err = client.Do(req)
		if err != nil {
			slog.Error("Cannot get user information while fetching URL", "URL", "https://graph.microsoft.com/v1.0/me", "HTTP Verb", "GET", "Error", err)
			return model.OAuth2User{}, err
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			bodyBytes, _ := io.ReadAll(res.Body)
			slog.Error("Cannot get user information while fetching URL", "URL", "https://graph.microsoft.com/v1.0/me", "HTTP Verb", "GET", "Error", string(bodyBytes))
			return model.OAuth2User{}, errors.New(string(bodyBytes))
		}

		err = json.NewDecoder(res.Body).Decode(&userContent.MicrosoftOAuth2User)
		if err != nil {
			slog.Error("Cannot decode JSON information", "Error", err)
			return model.OAuth2User{}, err
		}

	} else {
		return model.OAuth2User{}, errors.New("Authentication method not allowed")
	}

	defer res.Body.Close()

	slog.Info("User information fetched successfully")
	return userContent, err
}

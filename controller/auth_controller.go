package controller

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/RenanMonteiroS/MaestroSQLWeb/config"
	"github.com/RenanMonteiroS/MaestroSQLWeb/service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
)

type AuthController struct {
	service service.AuthService
}

func NewAuthController(sv service.AuthService) AuthController {
	return AuthController{service: sv}
}

var googleOAuthConfig = &oauth2.Config{
	RedirectURL:  config.GoogleOAuth2RedirectURL,
	ClientID:     config.GoogleOAuth2ClientID,
	ClientSecret: config.GoogleOAuth2ClientSecret,
	Scopes: []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
	},
	Endpoint: google.Endpoint,
}

var microsoftOAuthConfig = &oauth2.Config{
	RedirectURL:  config.MicrosoftOAuth2RedirectURL,
	ClientID:     config.MicrosoftOAuth2ClientID,
	ClientSecret: config.MicrosoftOAuth2ClientSecret,
	Scopes: []string{
		"openid",
		"profile",
		"User.Read",
		"email",
	},
	Endpoint: microsoft.AzureADEndpoint(config.MicrosoftOAuth2AzureADEndpoint),
}

func (ac *AuthController) LoginHandler(c *gin.Context) {
	var url string
	state := ac.service.GenerateStateOAuthCookie()

	session := sessions.Default(c)
	session.Set("oauth_state", state)
	session.Save()

	oAuthMethod := c.Query("method")
	if oAuthMethod == "google" {
		url = googleOAuthConfig.AuthCodeURL(state)
	} else if oAuthMethod == "microsoft" {
		url = microsoftOAuthConfig.AuthCodeURL(state)
	} else {
		c.JSON(http.StatusNotFound, map[string]any{"msg": "Login method not allowed"})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (ac *AuthController) GoogleCallBackHandler(c *gin.Context) {
	// Gets the session
	session := sessions.Default(c)

	// If the session passed through the callback route is not valid, returns an error message. Prevents from CSRF attacks
	if c.Query("state") != session.Get("oauth_state") {
		slog.Error("Invalid OAuth2 state")
		c.JSON(http.StatusBadRequest, map[string]any{"msg": "Invalid OAuth2 state"})
		return
	}

	// Gets the authorization code passed through the callback route, and convert it into an authorization token. It needs to be validated to prevent from CSRF attacks.
	code := c.Query("code")
	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		slog.Error("Cannot convert the authorization code into a token", "Error", err)
		c.JSON(http.StatusBadRequest, map[string]any{"msg": fmt.Sprintf("Cannot convert the authorization code into a token. Error: %v", err)})
		return
	}

	userInfo, err := ac.service.GetOAuth2UserInfo(token.AccessToken, "google")
	if err != nil {
		slog.Error("Cannot get user information", "Error", err)
		c.JSON(http.StatusBadRequest, map[string]any{"msg": fmt.Sprintf("Cannot get user information. Error: %v", err)})
		return
	}

	session.Set("userEmail", userInfo.GoogleOAuth2User.Email)
	err = session.Save()
	if err != nil {
		slog.Error("Failed to save session", "User", userInfo.GoogleOAuth2User.Email)
		c.JSON(http.StatusInternalServerError, map[string]any{"msg": "Failed to save session"})
	}
	c.SetCookie("userEmail", userInfo.GoogleOAuth2User.Email, 3600, "/", "", false, false)

	slog.Info("Authentication done successfully", "User", userInfo.GoogleOAuth2User.Email)

	c.Redirect(http.StatusPermanentRedirect, "/")
}

func (ac *AuthController) MicrosoftCallBackHandler(c *gin.Context) {
	// Gets the session
	session := sessions.Default(c)

	// If the session passed through the callback route is not valid, returns an error message. Prevents from CSRF attacks
	if c.Query("state") != session.Get("oauth_state") {
		slog.Error("Invalid OAuth2 state")
		c.JSON(http.StatusBadRequest, map[string]any{"msg": "Invalid OAuth2 state"})
		return
	}

	// Gets the authorization code passed through the callback route, and convert it into an authorization token. It needs to be validated to prevent from CSRF attacks.
	code := c.Query("code")
	token, err := microsoftOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		slog.Error("Cannot convert the authorization code into a token")
		c.JSON(http.StatusBadRequest, map[string]any{"msg": "Cannot convert the authorization code into a token"})
		return
	}

	userInfo, err := ac.service.GetOAuth2UserInfo(token.AccessToken, "microsoft")
	if err != nil {
		slog.Error("Cannot get user information", "Error", err)
		c.JSON(http.StatusBadRequest, map[string]any{"msg": fmt.Sprintf("Cannot get user information %v:", err)})
		return
	}

	if userInfo.MicrosoftOAuth2User.Mail == nil {
		email := strings.Split(userInfo.MicrosoftOAuth2User.UserPrincipalName, "#")[0]
		email = strings.Replace(email, "_", "@", 1)
		userInfo.MicrosoftOAuth2User.Mail = &email
	}

	session.Set("userEmail", *userInfo.MicrosoftOAuth2User.Mail)
	err = session.Save()
	if err != nil {
		slog.Error("Failed to save session", "User", &userInfo.MicrosoftOAuth2User.Mail)
		c.JSON(http.StatusInternalServerError, map[string]any{"msg": "Failed to save session"})
	}
	c.SetCookie("userEmail", *userInfo.MicrosoftOAuth2User.Mail, 3600, "/", "", false, false)

	slog.Info("Authentication done successfully", "User", *userInfo.MicrosoftOAuth2User.Mail)

	c.Redirect(http.StatusPermanentRedirect, "/")
}

func (ac *AuthController) SessionHandler(c *gin.Context) {

}

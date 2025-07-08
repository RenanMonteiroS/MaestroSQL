package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

// Handles the user login. Allowed methods: OSI, Microsoft OAuth2 and Google OAuth2
func (ac *AuthController) LoginHandler(c *gin.Context) {
	var url string

	authMethod := c.Query("method")
	if authMethod == "google" {
		state := ac.service.GenerateStateOAuthCookie()
		session := sessions.Default(c)
		session.Set("oauth_state", state)
		session.Save()

		url = googleOAuthConfig.AuthCodeURL(state)
	} else if authMethod == "microsoft" {
		state := ac.service.GenerateStateOAuthCookie()

		session := sessions.Default(c)
		session.Set("oauth_state", state)
		session.Save()

		url = microsoftOAuthConfig.AuthCodeURL(state)
	} else if authMethod == "osi" {
		if c.Request.Method != "POST" {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"msg": "OSI login requires POST"})
			return
		}
		url = config.AuthenticatorURL + "/login"

		type osiLoginResponse struct {
			JWT      *string `json:"JWT"`
			Msg      string  `json:"msg"`
			Status   string  `json:"status"`
			UserInfo *struct {
				Email           string `json:"email"`
				ID              string `json:"id"`
				TokenExpiration string `json:"tokenExpiration"`
			} `json:"userInfo,omitempty"`
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			slog.Error("Error reading request body", "Error", err)
			c.JSON(http.StatusInternalServerError, map[string]any{"msg": fmt.Sprintf("Error binding JSON %v", err)})
			return
		}

		client := http.Client{}
		req, err := http.NewRequest("POST", url, bytes.NewReader(body))
		if err != nil {
			slog.Error("Error creating request", "URL", url, "Error", err)
			c.JSON(http.StatusInternalServerError, map[string]any{"msg": fmt.Sprintf("Error binding JSON %v", err)})
			return
		}
		header := http.Header{"Content-Type": {"application/json"}}
		req.Header = header

		res, err := client.Do(req)
		if err != nil {
			slog.Error("Error making request", "URL", url, "Error", err)
			c.JSON(http.StatusInternalServerError, map[string]any{"msg": fmt.Sprintf("Error binding JSON %v", err)})
			return
		}
		defer res.Body.Close()

		var osiRes osiLoginResponse

		err = json.NewDecoder(res.Body).Decode(&osiRes)
		if err != nil {
			slog.Error("Cannot decode request body", "URL", url, "Error", err)
			c.JSON(http.StatusInternalServerError, map[string]any{"msg": fmt.Sprintf("Cannot decode request body %v", err)})
			return
		}

		if res.StatusCode != 200 {
			slog.Error("Request status is not OK", "URL", url, "Error", osiRes.Msg)
			c.JSON(http.StatusBadRequest, map[string]any{"msg": fmt.Sprintf("Request status is not OK %v", osiRes.Msg)})
			return
		}

		session := sessions.Default(c)
		session.Set("userEmail", osiRes.UserInfo.Email)
		session.Save()

		slog.Info("Login done successfully", "User", osiRes.UserInfo.Email)

		c.JSON(http.StatusOK, osiRes)
		return

	} else {
		c.JSON(http.StatusNotFound, map[string]any{"msg": "Login method not allowed"})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, url)
}

// Deletes the user session
func (ac *AuthController) LogoutHandler(c *gin.Context) {
	session := sessions.Default(c)

	session.Delete("userEmail")

	session.Save()
	c.JSON(http.StatusOK, map[string]any{"msg": "Logout done successfully"})
}

// Receives the callback from Google OAuth2
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

// Receives the callback from Microsoft OAuth2
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
		slog.Error("Cannot convert the authorization code into a token", "Error", err)
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

// Checks if a session exists
func (ac *AuthController) SessionHandler(c *gin.Context) {
	session := sessions.Default(c)
	sessionUserEmail := session.Get("userEmail")

	slog.Info("Checking if a session exists", "User", sessionUserEmail)

	if sessionUserEmail == nil {
		slog.Info("Session not found")
		c.JSON(http.StatusUnauthorized, map[string]any{"msg": "Session not found"})
		return
	}

	slog.Info("Session found", "User", sessionUserEmail)
	c.JSON(http.StatusOK, map[string]any{"msg": fmt.Sprintf("Session found: %v", sessionUserEmail), "userInfo": sessionUserEmail})
}

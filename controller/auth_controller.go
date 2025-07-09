package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/RenanMonteiroS/MaestroSQLWeb/config"
	"github.com/RenanMonteiroS/MaestroSQLWeb/model"
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
func (ac *AuthController) LoginHandler(ctx *gin.Context) {
	var url string

	authMethod := ctx.Query("method")
	if authMethod == "google" {
		state := ac.service.GenerateStateOAuthCookie()
		session := sessions.Default(ctx)
		session.Set("oauth_state", state)
		session.Save()
		slog.Info("oauth_state set into the session", "Origin", ctx.ClientIP())

		url = googleOAuthConfig.AuthCodeURL(state)
	} else if authMethod == "microsoft" {
		state := ac.service.GenerateStateOAuthCookie()

		session := sessions.Default(ctx)
		session.Set("oauth_state", state)
		session.Save()
		slog.Info("oauth_state set into the session", "Origin", ctx.ClientIP())

		url = microsoftOAuthConfig.AuthCodeURL(state)
	} else if authMethod == "osi" {
		if ctx.Request.Method != "POST" {
			slog.Error("OSI login requires a POST request", "Origin", ctx.ClientIP(), "Error", "OSI login requires a POST request")
			ctx.JSON(http.StatusMethodNotAllowed, model.APIResponse{Status: "error", Code: http.StatusMethodNotAllowed, Message: "OSI login requires a POST request", Errors: map[string]any{"methodNotAllowed": ctx.Request.Method}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
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

		body, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			slog.Error("Cannot bind JSON from request body", "Origin", ctx.ClientIP(), "Error", err.Error())
			ctx.JSON(http.StatusInternalServerError, model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Cannot bind JSON from request body", Errors: map[string]any{"bindJson": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
			return
		}

		client := http.Client{}
		req, err := http.NewRequest("POST", url, bytes.NewReader(body))
		if err != nil {
			slog.Error("Error creating request", "Origin", ctx.ClientIP(), "URL", url, "Error", err.Error())
			ctx.JSON(http.StatusInternalServerError, model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Error creating request", Errors: map[string]any{"request": err.Error(), "url": url}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
			return
		}
		header := http.Header{"Content-Type": {"application/json"}}
		req.Header = header

		res, err := client.Do(req)
		if err != nil {
			slog.Error("Error making request", "Origin", ctx.ClientIP(), "URL", url, "Error", err.Error())
			ctx.JSON(http.StatusInternalServerError, model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Error making request", Errors: map[string]any{"request": err.Error(), "url": url}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
			return
		}
		defer res.Body.Close()

		var osiRes osiLoginResponse

		err = json.NewDecoder(res.Body).Decode(&osiRes)
		if err != nil {
			slog.Error("Cannot bind JSON from request body", "Origin", ctx.ClientIP(), "Error", err.Error())
			ctx.JSON(http.StatusInternalServerError, model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Cannot bind JSON from request body", Errors: map[string]any{"bindJson": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
			return
		}

		if res.StatusCode != 200 {
			slog.Error("OSI Response is not OK", "Origin", ctx.ClientIP(), "Error", osiRes.Msg)
			ctx.JSON(http.StatusBadRequest, model.APIResponse{Status: "error", Code: http.StatusBadRequest, Message: "OSI Response is not OK", Errors: map[string]any{"osiMsg": osiRes.Msg}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
			return
		}

		session := sessions.Default(ctx)
		session.Set("userEmail", osiRes.UserInfo.Email)
		session.Save()

		slog.Info("Login done successfully", "Origin", ctx.ClientIP(), "User", osiRes.UserInfo.Email)

		ctx.JSON(http.StatusOK, model.APIResponse{Status: "success", Code: http.StatusOK, Message: "Login done successfully", Data: map[string]any{"user": osiRes.UserInfo.Email}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
		return

	} else {
		slog.Error("Login method not allowed", "Origin", ctx.ClientIP(), "URL", url)
		if strings.Contains(ctx.GetHeader("Accept"), "application/json") {
			ctx.JSON(http.StatusNotFound, model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Login method not allowed", Errors: map[string]any{"loginMethod": "Login method not allowed"}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
		}
		ctx.Redirect(http.StatusNotFound, "/404")
		return
	}

	slog.Info("Redirecting to OAuth2 callback URL", "Origin", ctx.ClientIP(), "URL", url)
	ctx.Redirect(http.StatusTemporaryRedirect, url)
}

// Deletes the user session
func (ac *AuthController) LogoutHandler(ctx *gin.Context) {
	session := sessions.Default(ctx)

	user := session.Get("userEmail")
	oauthState := session.Get("oauth_state")

	if user != nil {
		slog.Info("Logout done successfully", "User", user)
		session.Delete("userEmail")
	}

	if oauthState != nil {
		session.Delete("oauth_state")
	}

	session.Save()

	ctx.JSON(http.StatusOK, model.APIResponse{Status: "success", Code: http.StatusOK, Message: "Logout done successfully", Data: map[string]any{"user": user}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
	return
}

// Receives the callback from Google OAuth2
func (ac *AuthController) GoogleCallBackHandler(ctx *gin.Context) {
	// Gets the session
	session := sessions.Default(ctx)

	// If the session passed through the callback route is not valid, returns an error message. Prevents from CSRF attacks
	if ctx.Query("state") != session.Get("oauth_state") {
		slog.Error("Invalid OAuth2 state", "Origin", ctx.ClientIP(), "URL Query State", ctx.Query("state"), "Session OAuth2 state", session.Get("oauth_state"))
		ctx.JSON(http.StatusBadRequest, model.APIResponse{Status: "error", Code: http.StatusBadRequest, Message: "Invalid OAuth2 state", Errors: map[string]any{"oauth2": "Invalid OAuth2 state"}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
		return
	}

	// Gets the authorization code passed through the callback route, and convert it into an authorization token. It needs to be validated to prevent from CSRF attacks.
	code := ctx.Query("code")
	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		slog.Error("Cannot convert the authorization code into a token", "Origin", ctx.ClientIP(), "Error", err.Error())
		ctx.JSON(http.StatusBadRequest, model.APIResponse{Status: "error", Code: http.StatusBadRequest, Message: "Cannot convert the authorization code into a token.", Errors: map[string]any{"exchange": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
		return
	}

	slog.Info("Getting OAuth2 user information", "Origin", ctx.ClientIP())
	userInfo, err := ac.service.GetOAuth2UserInfo(token.AccessToken, "google")
	if err != nil {
		slog.Error("Cannot get user information", "Origin", ctx.ClientIP(), "Error", err.Error())
		ctx.JSON(http.StatusBadRequest, model.APIResponse{Status: "error", Code: http.StatusBadRequest, Message: "Cannot get user information", Errors: map[string]any{"userInfo": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
		return
	}

	session.Set("userEmail", userInfo.GoogleOAuth2User.Email)
	session.Save()

	slog.Info("Authentication done successfully", "Origin", ctx.ClientIP(), "User", userInfo.GoogleOAuth2User.Email)

	ctx.Redirect(http.StatusPermanentRedirect, "/")
}

// Receives the callback from Microsoft OAuth2
func (ac *AuthController) MicrosoftCallBackHandler(ctx *gin.Context) {
	// Gets the session
	session := sessions.Default(ctx)

	// If the session passed through the callback route is not valid, returns an error message. Prevents from CSRF attacks
	if ctx.Query("state") != session.Get("oauth_state") {
		slog.Error("Invalid OAuth2 state", "Origin", ctx.ClientIP(), "URL Query State", ctx.Query("state"), "Session OAuth2 state", session.Get("oauth_state"))
		ctx.JSON(http.StatusBadRequest, model.APIResponse{Status: "error", Code: http.StatusBadRequest, Message: "Invalid OAuth2 state", Errors: map[string]any{"oauth2": "Invalid OAuth2 state"}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
		return
	}

	// Gets the authorization code passed through the callback route, and convert it into an authorization token. It needs to be validated to prevent from CSRF attacks.
	code := ctx.Query("code")
	token, err := microsoftOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		slog.Error("Cannot convert the authorization code into a token", "Origin", ctx.ClientIP(), "Error", err.Error())
		ctx.JSON(http.StatusBadRequest, model.APIResponse{Status: "error", Code: http.StatusBadRequest, Message: "Cannot convert the authorization code into a token.", Errors: map[string]any{"exchange": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
		return
	}

	slog.Info("Getting OAuth2 user information", "Origin", ctx.ClientIP())
	userInfo, err := ac.service.GetOAuth2UserInfo(token.AccessToken, "microsoft")
	if err != nil {
		slog.Error("Cannot get user information", "Origin", ctx.ClientIP(), "Error", err.Error())
		ctx.JSON(http.StatusBadRequest, model.APIResponse{Status: "error", Code: http.StatusBadRequest, Message: "Cannot get user information", Errors: map[string]any{"userInfo": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
		return
	}

	if userInfo.MicrosoftOAuth2User.Mail == nil {
		email := strings.Split(userInfo.MicrosoftOAuth2User.UserPrincipalName, "#")[0]
		email = strings.Replace(email, "_", "@", 1)
		userInfo.MicrosoftOAuth2User.Mail = &email
	}

	session.Set("userEmail", *userInfo.MicrosoftOAuth2User.Mail)
	session.Save()

	slog.Info("Authentication done successfully", "Origin", ctx.ClientIP(), "User", userInfo.GoogleOAuth2User.Email)

	ctx.Redirect(http.StatusPermanentRedirect, "/")
}

// Checks if a session exists
func (ac *AuthController) SessionHandler(ctx *gin.Context) {
	session := sessions.Default(ctx)
	sessionUserEmail := session.Get("userEmail")

	slog.Info("Checking if a session exists", "Origin", ctx.ClientIP(), "User", sessionUserEmail)

	if sessionUserEmail == nil {
		slog.Info("None session was found", "Origin", ctx.ClientIP())
		ctx.JSON(http.StatusUnauthorized, model.APIResponse{Status: "error", Code: http.StatusUnauthorized, Message: "None session was found", Errors: map[string]any{"session": "None session was found"}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})
		return
	}

	slog.Info("Session found", "Origin", ctx.ClientIP(), "User", sessionUserEmail)
	ctx.JSON(http.StatusOK, model.APIResponse{Status: "success", Code: http.StatusOK, Message: "Session found", Data: map[string]any{"user": sessionUserEmail}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Request.URL.Path})

}

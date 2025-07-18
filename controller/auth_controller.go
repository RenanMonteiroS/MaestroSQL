package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/RenanMonteiroS/MaestroSQLWeb/config"
	"github.com/RenanMonteiroS/MaestroSQLWeb/model"
	"github.com/RenanMonteiroS/MaestroSQLWeb/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
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
func (ac *AuthController) LoginHandler(ctx *fiber.Ctx) error {
	var url string

	authMethod := ctx.Query("method")
	if authMethod == "google" {
		state := ac.service.GenerateStateOAuthCookie()
		sess, ok := ctx.Locals("session").(*session.Session)
		if !ok {
			return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Internal server error: session not found", Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
		}
		sess.Set("oauth_state", state)
		sess.Save()
		slog.Info("oauth_state set into the session", "Origin", ctx.IP())

		url = googleOAuthConfig.AuthCodeURL(state)
	} else if authMethod == "microsoft" {
		state := ac.service.GenerateStateOAuthCookie()

		sess, ok := ctx.Locals("session").(*session.Session)
		if !ok {
			return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Internal server error: session not found", Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
		}
		sess.Set("oauth_state", state)
		sess.Save()
		slog.Info("oauth_state set into the session", "Origin", ctx.IP())

		url = microsoftOAuthConfig.AuthCodeURL(state)
	} else if authMethod == "osi" {
		if ctx.Method() != "POST" {
			slog.Error("OSI login requires a POST request", "Origin", ctx.IP(), "Error", "OSI login requires a POST request")
			return ctx.Status(http.StatusMethodNotAllowed).JSON(model.APIResponse{Status: "error", Code: http.StatusMethodNotAllowed, Message: "OSI login requires a POST request", Errors: map[string]any{"methodNotAllowed": ctx.Method()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
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

		type osiLoginRequest struct {
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required"`
			MfaToken string `json:"mfaKey" binding:"required"`
		}

		var osiRequestBody osiLoginRequest

		err := ctx.BodyParser(&osiRequestBody)
		if err != nil {
			slog.Error("Cannot parse request body", "Origin", ctx.IP(), "Error", err.Error())
			return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Cannot parse request body", Errors: map[string]any{"bindJson": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})

		}

		jsonBody, err := json.Marshal(osiRequestBody)
		if err != nil {
			slog.Error("Cannot marshal request body", "Origin", ctx.IP(), "Error", err.Error())
			return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Cannot marshal request body", Errors: map[string]any{"bindJson": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
		}

		client := http.Client{}
		req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
		if err != nil {
			slog.Error("Error creating request", "Origin", ctx.IP(), "URL", url, "Error", err.Error())
			return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Error creating request", Errors: map[string]any{"request": err.Error(), "url": url}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
		}
		header := http.Header{"Content-Type": {"application/json"}}
		req.Header = header

		res, err := client.Do(req)
		if err != nil {
			slog.Error("Error making request", "Origin", ctx.IP(), "URL", url, "Error", err.Error())
			return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Error making request", Errors: map[string]any{"request": err.Error(), "url": url}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
		}
		defer res.Body.Close()

		var osiRes osiLoginResponse

		err = json.NewDecoder(res.Body).Decode(&osiRes)
		if err != nil {
			slog.Error("Cannot bind JSON from request body", "Origin", ctx.IP(), "Error", err.Error())
			return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Cannot bind JSON from request body", Errors: map[string]any{"bindJson": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
		}

		if res.StatusCode != 200 {
			slog.Error("OSI Response is not OK", "Origin", ctx.IP(), "Error", osiRes.Msg)
			return ctx.Status(http.StatusBadRequest).JSON(model.APIResponse{Status: "error", Code: http.StatusBadRequest, Message: "OSI Response is not OK", Errors: map[string]any{"osiMsg": osiRes.Msg}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
		}

		sess, ok := ctx.Locals("session").(*session.Session)
		if !ok {
			return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Internal server error: session not found", Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
		}
		sess.Set("userEmail", osiRes.UserInfo.Email)
		sess.Save()

		slog.Info("Login done successfully", "Origin", ctx.IP(), "User", osiRes.UserInfo.Email)

		return ctx.Status(http.StatusOK).JSON(model.APIResponse{Status: "success", Code: http.StatusOK, Message: "Login done successfully", Data: map[string]any{"user": osiRes.UserInfo.Email}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	} else {
		slog.Error("Login method not allowed", "Origin", ctx.IP(), "URL", url)
		if strings.Contains(ctx.Get("Accept"), "application/json") {
			return ctx.Status(http.StatusNotFound).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Login method not allowed", Errors: map[string]any{"loginMethod": "Login method not allowed"}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
		}
		return ctx.Redirect("/404", http.StatusNotFound)

	}

	slog.Info("Redirecting to OAuth2 callback URL", "Origin", ctx.IP(), "URL", url)
	return ctx.Redirect(url, http.StatusTemporaryRedirect)
}

// Deletes the user session
func (ac *AuthController) LogoutHandler(ctx *fiber.Ctx) error {
	sess, ok := ctx.Locals("session").(*session.Session)
	if !ok {
		return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Internal server error: session not found", Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	}

	user := sess.Get("userEmail")
	oauthState := sess.Get("oauth_state")

	if user != nil {
		slog.Info("Logout done successfully", "User", user)
		sess.Delete("userEmail")
	}

	if oauthState != nil {
		sess.Delete("oauth_state")
	}

	sess.Save()

	return ctx.Status(http.StatusOK).JSON(model.APIResponse{Status: "success", Code: http.StatusOK, Message: "Logout done successfully", Data: map[string]any{"user": user}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
}

// Receives the callback from Google OAuth2
func (ac *AuthController) GoogleCallBackHandler(ctx *fiber.Ctx) error {
	// Gets the session
	sess, ok := ctx.Locals("session").(*session.Session)
		if !ok {
			return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Internal server error: session not found", Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
		}

	// If the session passed through the callback route is not valid, returns an error message. Prevents from CSRF attacks
	if ctx.Query("state") != sess.Get("oauth_state") {
		slog.Error("Invalid OAuth2 state", "Origin", ctx.IP(), "URL Query State", ctx.Query("state"), "Session OAuth2 state", sess.Get("oauth_state"))
		return ctx.Status(http.StatusBadRequest).JSON(model.APIResponse{Status: "error", Code: http.StatusBadRequest, Message: "Invalid OAuth2 state", Errors: map[string]any{"oauth2": "Invalid OAuth2 state"}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	}

	// Gets the authorization code passed through the callback route, and convert it into an authorization token. It needs to be validated to prevent from CSRF attacks.
	code := ctx.Query("code")
	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		slog.Error("Cannot convert the authorization code into a token", "Origin", ctx.IP(), "Error", err.Error())
		return ctx.Status(http.StatusBadRequest).JSON(model.APIResponse{Status: "error", Code: http.StatusBadRequest, Message: "Cannot convert the authorization code into a token.", Errors: map[string]any{"exchange": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	}

	slog.Info("Getting OAuth2 user information", "Origin", ctx.IP())
	userInfo, err := ac.service.GetOAuth2UserInfo(token.AccessToken, "google")
	if err != nil {
		slog.Error("Cannot get user information", "Origin", ctx.IP(), "Error", err.Error())
		return ctx.Status(http.StatusBadRequest).JSON(model.APIResponse{Status: "error", Code: http.StatusBadRequest, Message: "Cannot get user information", Errors: map[string]any{"userInfo": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	}

	sess.Set("userEmail", userInfo.GoogleOAuth2User.Email)
	sess.Save()

	slog.Info("Authentication done successfully", "Origin", ctx.IP(), "User", userInfo.GoogleOAuth2User.Email)

	return ctx.Redirect("/", http.StatusPermanentRedirect)
}

// Receives the callback from Microsoft OAuth2
func (ac *AuthController) MicrosoftCallBackHandler(ctx *fiber.Ctx) error {
	// Gets the session
	sess, ok := ctx.Locals("session").(*session.Session)
		if !ok {
			return ctx.Status(http.StatusInternalServerError).JSON(model.APIResponse{Status: "error", Code: http.StatusInternalServerError, Message: "Internal server error: session not found", Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
		}

	// If the session passed through the callback route is not valid, returns an error message. Prevents from CSRF attacks
	if ctx.Query("state") != sess.Get("oauth_state") {
		slog.Error("Invalid OAuth2 state", "Origin", ctx.IP(), "URL Query State", ctx.Query("state"), "Session OAuth2 state", sess.Get("oauth_state"))
		return ctx.Status(http.StatusBadRequest).JSON(model.APIResponse{Status: "error", Code: http.StatusBadRequest, Message: "Invalid OAuth2 state", Errors: map[string]any{"oauth2": "Invalid OAuth2 state"}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	}

	// Gets the authorization code passed through the callback route, and convert it into an authorization token. It needs to be validated to prevent from CSRF attacks.
	code := ctx.Query("code")
	token, err := microsoftOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		slog.Error("Cannot convert the authorization code into a token", "Origin", ctx.IP(), "Error", err.Error())
		return ctx.Status(http.StatusBadRequest).JSON(model.APIResponse{Status: "error", Code: http.StatusBadRequest, Message: "Cannot convert the authorization code into a token.", Errors: map[string]any{"exchange": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	}

	slog.Info("Getting OAuth2 user information", "Origin", ctx.IP())
	userInfo, err := ac.service.GetOAuth2UserInfo(token.AccessToken, "microsoft")
	if err != nil {
		slog.Error("Cannot get user information", "Origin", ctx.IP(), "Error", err.Error())
		return ctx.Status(http.StatusBadRequest).JSON(model.APIResponse{Status: "error", Code: http.StatusBadRequest, Message: "Cannot get user information", Errors: map[string]any{"userInfo": err.Error()}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
	}

	if userInfo.MicrosoftOAuth2User.Mail == nil {
		email := strings.Split(userInfo.MicrosoftOAuth2User.UserPrincipalName, "#")[0]
		email = strings.Replace(email, "_", "@", 1)
		userInfo.MicrosoftOAuth2User.Mail = &email
	}

	sess.Set("userEmail", *userInfo.MicrosoftOAuth2User.Mail)
	sess.Save()

	slog.Info("Authentication done successfully", "Origin", ctx.IP(), "User", userInfo.GoogleOAuth2User.Email)

	return ctx.Redirect("/", http.StatusPermanentRedirect)
}

// Checks if a session exists
func (ac *AuthController) SessionHandler(ctx *fiber.Ctx) error {
	session := ctx.Locals("session").(*session.Session)
	sessionUserEmail := session.Get("userEmail")

	slog.Info("Checking if a session exists", "Origin", ctx.IP(), "User", sessionUserEmail)

	if sessionUserEmail == nil {
		slog.Info("None session was found", "Origin", ctx.IP())
		return ctx.Status(http.StatusUnauthorized).JSON(model.APIResponse{Status: "error", Code: http.StatusUnauthorized, Message: "None session was found", Errors: map[string]any{"session": "None session was found"}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})

	}

	slog.Info("Session found", "Origin", ctx.IP(), "User", sessionUserEmail)
	return ctx.Status(http.StatusOK).JSON(model.APIResponse{Status: "success", Code: http.StatusOK, Message: "Session found", Data: map[string]any{"user": sessionUserEmail}, Timestamp: time.Now().Format(time.RFC3339), Path: ctx.Path()})
}

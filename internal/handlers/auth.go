package handlers

import (
	"cmp"
	"encoding/json"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/webpoint-solutions-llc/go-starter/internal/config"
	"github.com/webpoint-solutions-llc/go-starter/internal/constants"
	"github.com/webpoint-solutions-llc/go-starter/internal/dto"
	"github.com/webpoint-solutions-llc/go-starter/internal/errorhandler"
	"github.com/webpoint-solutions-llc/go-starter/internal/types"
	"github.com/webpoint-solutions-llc/go-starter/internal/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	Clientid     = os.Getenv(`GOOGLE_CLIENT_ID`)
	Clientsecret = os.Getenv(`GOOGLE_CLIENT_SECRET`)
	Redirecturi  = os.Getenv(`GOOGLE_REDIRECT_URI`)

	googleUserInfoEndpoint = cmp.Or(os.Getenv("GOOGLE_USER_INFO_URL"), "https://www.googleapis.com/oauth2/v3/userinfo")

	googleOauthConfig = &oauth2.Config{
		ClientID:     Clientid,
		ClientSecret: Clientsecret,
		RedirectURL:  Redirecturi,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	}
)

// GoogleLogin godoc
// @Summary Login with google
// @Description Login with google oauth open in browser for login
// @Tags Authentication
// @Security ApiKeyAuth
// @Router /api/v1/auth/google/login [get]
func (h *Handler) GoogleLogin(c echo.Context) error {
	if config.Cfg.GoogleClientID == "" || config.Cfg.GoogleClientSecret == "" || config.Cfg.GoogleRedirectURI == "" {
		return errorhandler.ErrorInternal("Google Oauth2 is not configured")
	}

	url := googleOauthConfig.AuthCodeURL("state")

	return c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleLogin godoc
// @Summary Login with google
// @Description Login with google oauth
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.CodeExchangeRequest true "code gotten from google"
// @Security ApiKeyAuth
// @Success 200 {object} dto.CodeExchangeRequest "Successfully logged in"
// @Failure 400 {object} errorhandler.HttpError "Invalid request format"
// @Failure 500 {object} errorhandler.HttpError "Internal server error"
// @Router /api/v1/auth/google/callback [post]
func (h *Handler) GoogleCodeExchange(c echo.Context) error {
	ctx := c.Request().Context()

	var requestData dto.CodeExchangeRequest

	if err := c.Bind(&requestData); err != nil || requestData.Code == "" {
		return errorhandler.ErrorBadRequest("Code is required")
	}

	// Exchange the authorization code for an access token
	token, err := googleOauthConfig.Exchange(ctx, requestData.Code)
	if err != nil {
		c.Logger().Error(err)
		return errorhandler.ErrorBadRequest("Invalid Code")
	}

	client := googleOauthConfig.Client(ctx, token)
	googleUserInfo, err := client.Get(googleUserInfoEndpoint)
	if err != nil {
		return errorhandler.ErrorInternal(err)
	}
	defer googleUserInfo.Body.Close()

	// Parse the user info response
	var userInfo dto.GoogleCallbackUserInfo
	if err := json.NewDecoder(googleUserInfo.Body).Decode(&userInfo); err != nil {
		return errorhandler.ErrorInternal(err)
	}

	userAgent := c.Request().UserAgent()

	rawIP := c.RealIP()
	parsed, _ := netip.ParseAddr(rawIP)

	ipPtr := &parsed

	var name string

	if userInfo.Name != "" {
		name = userInfo.Name
	} else {
		name = userInfo.GivenName + " " + userInfo.FamilyName
	}

	loginParams := dto.LoginParams{
		Email:    userInfo.Email,
		GoogleID: &userInfo.Subject,
		Name:     &name,

		UserAgent: &userAgent,
		IPAddress: ipPtr,
	}

	res, err := h.svc.InserSocialLoginUser(ctx, loginParams)
	if err != nil {
		c.Logger().Error(err)
		return errorhandler.ErrorInternal("Failed to insert User")
	}

	return h.res.JSON(c, res, dto.ResponderOptions{Message: constants.MsgSignInSuccessful})
}

// AppleLogin godoc
// @Summary Login with apple
// @Description Login with apple oauth open in browser for login
// @Tags Authentication
// @Security ApiKeyAuth
// @Router /api/v1/auth/apple/login [get]
func (h *Handler) AppleLogin(c echo.Context) error {
	authURL, err := h.svc.GetAppleLoginURL()
	if err != nil {
		return errorhandler.ErrorInternal(err)
	}

	// Redirect the user to Apple's authorization page
	return c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// AppleLogin godoc
// @Summary Login with apple
// @Description Login with apple oauth
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.CodeExchangeRequest true "code gotten from apple"
// @Security ApiKeyAuth
// @Success 200 {object} dto.CodeExchangeRequest "Successfully logged in"
// @Failure 400 {object} errorhandler.HttpError "Invalid request format"
// @Failure 500 {object} errorhandler.HttpError "Internal server error"
// @Router /api/v1/auth/apple/callback [post]
func (h *Handler) AppleCodeExchange(c echo.Context) error {
	ctx := c.Request().Context()

	var requestData dto.CodeExchangeRequest

	if err := c.Bind(&requestData); err != nil || requestData.Code == "" {
		return errorhandler.ErrorBadRequest("Code is required")
	}

	// Exchange the authorization code for an access token
	claim, err := h.svc.ExchangeCodeWithApple(ctx, requestData.Code)
	if err != nil {
		return errorhandler.ErrorBadRequest(err)
	}

	userAgent := c.Request().UserAgent()

	rawIP := c.RealIP()
	parsed, _ := netip.ParseAddr(rawIP)

	ipPtr := &parsed

	loginParams := dto.LoginParams{
		Email:   claim.Email,
		AppleID: &claim.Subject,
		Name:    &claim.Name,

		UserAgent: &userAgent,
		IPAddress: ipPtr,
	}

	res, err := h.svc.InserSocialLoginUser(ctx, loginParams)
	if err != nil {
		return errorhandler.ErrorInternal("Failed to insert User")
	}

	return h.res.JSON(c, res, dto.ResponderOptions{Message: constants.MsgSignInSuccessful})
}

// EmailLogin godoc
// @Summary Login with email
// @Description Login with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.EmailLoginRequest true "Email Login Request"
// @Success 200 {object} dto.LoginResponse "Successfully logged in"
// @Failure 400 {object} errorhandler.HttpError "Invalid request format"
// @Failure 500 {object} errorhandler.HttpError "Internal server error"
// @Security ApiKeyAuth
// @Router /api/v1/auth/login [post]
func (h *Handler) EmailLogin(c echo.Context) error {
	ctx := c.Request().Context()

	var req dto.EmailLoginRequest
	if err := c.Bind(&req); err != nil {
		return errorhandler.ErrorBadRequest("Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return errorhandler.ErrorBadRequest(err)
	}

	userAgent := c.Request().UserAgent()

	rawIP := c.RealIP()
	parsed, _ := netip.ParseAddr(rawIP)

	ipPtr := &parsed
	params := dto.LoginParams{
		Email:     req.Email,
		Password:  &req.Password,
		UserAgent: &userAgent,
		IPAddress: ipPtr,
		IsAdmin:   req.IsAdmin,
	}

	res, err := h.svc.EmailLogin(ctx, params)
	if err != nil {
		c.Logger().Error("Failed to login user ", "error", err)

		return errorhandler.ErrorInternal(err)
	}

	return h.res.JSON(c, res, dto.ResponderOptions{Message: constants.MsgSignInSuccessful})
}

// EmailSignUp godoc
// @Summary Sign up with email
// @Description Sign up with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.EmailSignUpRequest true "Email Sign Up Request"
// @Success 200 {object} dto.SuccessMessageResponse "Successfully signed up"
// @Failure 400 {object} errorhandler.HttpError "Invalid request format"
// @Failure 500 {object} errorhandler.HttpError "Internal server error"
// @Security ApiKeyAuth
// @Router /api/v1/auth/signup [post]
func (h *Handler) EmailSignUp(c echo.Context) error {
	ctx := c.Request().Context()

	var req dto.EmailSignUpRequest
	if err := c.Bind(&req); err != nil {
		return errorhandler.ErrorBadRequest("Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return errorhandler.ErrorBadRequest(err)
	}

	_, err := h.svc.EmailSignUp(ctx, req)
	if err != nil {
		c.Logger().Error("Failed to sign up user ", "error", err)

		return err
	}

	return h.res.JSON(c, nil, dto.ResponderOptions{Message: constants.MsgSignInSuccessful})
}

func (h *Handler) EmailVerify(c echo.Context) error {
	ctx := c.Request().Context()

	redirectURL, err := url.Parse(config.Cfg.FrontendURL + "/login")
	if err != nil {
		return errorhandler.ErrorInternal(err)
	}

	q := redirectURL.Query()
	q.Set("is_verified", "false")
	redirectURL.RawQuery = q.Encode()

	token := c.QueryParam("token")
	if token == "" {
		c.Logger().Error("Token is required")
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL.String())
	}

	res, err := h.svc.VerifyUserEmail(ctx, token)
	if err != nil {

		q.Set("email", utils.EncryptAndEncodeURL(res.Email))
		q.Set("is_verified", "true")
		redirectURL.RawQuery = q.Encode()

		c.Logger().Error("Failed to verify user email ", "error", err)

		redirectFailURL, err := url.Parse(config.Cfg.FrontendURL + "/signup")
		if err != nil {
			return errorhandler.ErrorInternal(err)
		}
		q := redirectFailURL.Query()
		q.Set("is_verified", "false")
		q.Set("email", utils.EncryptAndEncodeURL(res.Email))
		redirectFailURL.RawQuery = q.Encode()

		return c.Redirect(http.StatusTemporaryRedirect, redirectFailURL.String())
	}
	q.Set("is_verified", "true")
	q.Set("email", utils.EncryptAndEncodeURL(res.Email))
	redirectURL.RawQuery = q.Encode()

	return c.Redirect(http.StatusTemporaryRedirect, redirectURL.String())
}

// PasswordResetEmail godoc
// @Summary Get email verification link
// @Description generates password reset link and sends it to the user email
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.PasswordResetRequest true "password reset request"
// @Success 200 {object} dto.PasswordResetRequest "Successfully retrieved email"
// @Failure 400 {object} errorhandler.HttpError "email is required"
// @Failure 404 {object} errorhandler.HttpError "email not found"
// @Failure 500 {object} errorhandler.HttpError "Internal server error"
// @Security ApiKeyAuth
// @Router /api/v1/auth/password-reset [post]
func (h *Handler) PasswordReset(c echo.Context) error {
	ctx := c.Request().Context()

	var req dto.PasswordResetRequest
	if err := c.Bind(&req); err != nil {
		return errorhandler.ErrorBadRequest("Invalid request format")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	_, err := h.svc.PasswordResetLink(ctx, req.Email)
	if err != nil {
		return errorhandler.ErrorNotFound(err)
	}
	return h.res.JSON(c, dto.SuccessMessageResponse{
		Message: constants.MsgEmailVerificationSent,
	})
}

// PasswordResetTokenCheck godoc
// @Summary Reset Token check api
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.PasswordResetTokenCheck true "password reset confirm request"
// @Success 200 {object} dto.SuccessMessageResponse "password reset successfully"
// @Failure 400 {object} errorhandler.HttpError "Token expired"
// @Failure 404 {object} errorhandler.HttpError "User not found"
// @Failure 500 {object} errorhandler.HttpError "Internal server error"
// @Security x-reset-token
// @Router /api/v1/auth/password-token-check [post]
func (h *Handler) PasswordResetTokenCheck(c echo.Context) error {
	ctx := c.Request().Context()
	var req dto.PasswordResetTokenCheck
	if err := c.Bind(&req); err != nil {
		return errorhandler.ErrorBadRequest("Invalid request format")
	}
	if err := c.Validate(&req); err != nil {
		return errorhandler.ErrorBadRequest(err)
	}
	_, err := h.svc.PasswordResetTokenCheck(ctx, req)
	if err != nil {
		return err
	}
	return h.res.JSON(c, dto.SuccessMessageResponse{Message: constants.MsgValidToken})
}

// PasswordResetConfirm godoc
// @Summary Reset password api
// @Description Resets the password for the user, token is sent in x-api-token, token and password are required
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.PasswordResetConfirmRequest true "password reset confirm request"
// @Success 200 {object} dto.SuccessMessageResponse "password reset successfully"
// @Failure 400 {object} errorhandler.HttpError "Token expired"
// @Failure 404 {object} errorhandler.HttpError "User not found"
// @Failure 500 {object} errorhandler.HttpError "Internal server error"
// @Security ApiKeyAuth
// @Security x-reset-token
// @Router /api/v1/auth/password-reset/confirm [post]
func (h *Handler) PasswordResetConfirm(c echo.Context) error {
	ctx := c.Request().Context()
	var req dto.PasswordResetConfirmRequest
	if err := c.Bind(&req); err != nil {
		return errorhandler.ErrorBadRequest("Invalid request format")
	}

	authHeader := c.Request().Header.Get("x-reset-token")
	if authHeader == "" {
		return errorhandler.ErrorUnauthorized("Authorization header is required")
	}

	if err := c.Validate(&req); err != nil {
		return errorhandler.ErrorBadRequest(err)
	}
	_, err := h.svc.PasswordResetConfirm(ctx, req)
	if err != nil {
		return errorhandler.ErrorInternal(err)
	}
	return h.res.JSON(c, dto.SuccessMessageResponse{Message: constants.MsgPasswordUpdatedSuccessfully})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Uses a valid refresh token to generate a new access token
// @Tags Authentication
// @Security ApiKeyAuth
// @Security BearerAuth
// @Success 200 {object} dto.RefreshTokenResponse "New access token"
// @Failure 400 {object} errorhandler.HttpError "Invalid refresh token"
// @Failure 500 {object} errorhandler.HttpError "Internal error"
// @Router /api/v1/auth/refresh-token [get]
func (h *Handler) RefreshToken(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return errorhandler.ErrorUnauthorized("Authorization header is required")
	}

	// extract the token from the header
	refreshToken := strings.TrimPrefix(authHeader, "Bearer ")

	hashedToken := utils.GenerateRefreshTokenHash(refreshToken)

	// Parse userID as uuid.UUID

	session, err := h.svc.GetSessionByRefreshTokenHash(c.Request().Context(), hashedToken)
	if err != nil {
		c.Logger().Error("Refresh", "error", err)
		return errorhandler.ErrorBadRequest("Invalid refresh_token")
	}

	newAccessToken, err := utils.GenerateJWT(types.JWTPlayload{
		UserID:    session.UserID.String(),
		Sid:       session.ID.String(),
		TokenType: types.TokenTypeAccess,
		Duration:  config.Cfg.AccessTokenDuration,
	})
	if err != nil {
		return errorhandler.ErrorInternal("Failed to generate access token")
	}

	return h.res.JSON(c, dto.RefreshTokenResponse{
		AccessToken: newAccessToken,
	})
}

// ResendEmailVerification godoc
// @Summary Resend email for verification
// @Description Resend email for verification
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.ResendEmailVerificationRequest true "Email Sign Up Request"
// @Success 200 {object} dto.SuccessMessageResponse "Successfully signed up"
// @Failure 400 {object} errorhandler.HttpError "Invalid request format"
// @Failure 500 {object} errorhandler.HttpError "Internal server error"
// @Security ApiKeyAuth
// @Router /api/v1/auth/resend-verification-email [post]
func (h *Handler) ResendEmailVerification(c echo.Context) error {
	ctx := c.Request().Context()

	var req dto.ResendEmailVerificationRequest
	if err := c.Bind(&req); err != nil {
		return errorhandler.ErrorBadRequest("Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return errorhandler.ErrorBadRequest(err)
	}

	_, err := h.svc.ResendEmailVerification(ctx, req.Email)
	if err != nil {
		return errorhandler.ErrorInternal("Failed to resend verification email : " + err.Error())
	}

	return h.res.JSON(c, dto.SuccessMessageResponse{Message: constants.MsgEmailVerificationSent})
}

// Logout godoc
// @Summary logout
// @Description invaliate current session
// @Tags Authentication
// @Success 200 {object} dto.SuccessMessageResponse "Sign out successful."
// @Failure 400 {object} errorhandler.HttpError "Invalid request format"
// @Failure 500 {object} errorhandler.HttpError "Internal server error"
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /api/v1/auth/logout [get]
func (h *Handler) Logout(c echo.Context) error {
	ctx := c.Request().Context()

	claims := c.Get("claims").(*types.CustomClaims)

	sessionID, err := uuid.Parse(claims.Sid)
	if err != nil {
		return errorhandler.ErrorBadRequest("Invalid Sesion ID")
	}

	_, err = h.svc.CleanupSession(ctx, sessionID)
	if err != nil {
		return errorhandler.ErrorInternal("Failed to logout")
	}

	return h.res.JSON(c, dto.SuccessMessageResponse{Message: constants.MsgSignOutSuccessful})
}

// PasswordUpdate godoc
// @Summary Update Password
// @Description update password from user settings
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.PasswordUpdateRequest true "password update request"
// @Success 200 {object} dto.PasswordUpdateRequest "Password updated successfully"
// @Failure 400 {object} errorhandler.HttpError "Bad requst"
// @Failure 500 {object} errorhandler.HttpError "Internal server error"
// @Security ApiKeyAuth
// @Security BearerAuth
// @Router /api/v1/auth/password-update [post]
func (h *Handler) UpdatePassword(c echo.Context) error {
	ctx := c.Request().Context()
	var req dto.PasswordUpdateRequest
	if err := c.Bind(&req); err != nil {
		return errorhandler.ErrorBadRequest("Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return errorhandler.ErrorBadRequest(err)
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return errorhandler.ErrorBadRequest(constants.MsgReLogin)
	}

	err = h.svc.UpdatePassword(ctx, req, userID)
	if err != nil {
		c.Logger().Error(err)
		return errorhandler.ErrorInternal("Failed to Update password")
	}
	return h.res.JSON(c, dto.SuccessMessageResponse{Message: constants.MsgPasswordUpdatedSuccessfully})
}

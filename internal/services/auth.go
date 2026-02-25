package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/matthewhartstonge/argon2"
	"github.com/webpoint-solutions-llc/go-starter/internal/config"
	"github.com/webpoint-solutions-llc/go-starter/internal/constants"
	"github.com/webpoint-solutions-llc/go-starter/internal/db/sqlc"
	"github.com/webpoint-solutions-llc/go-starter/internal/dto"
	"github.com/webpoint-solutions-llc/go-starter/internal/errorhandler"
	"github.com/webpoint-solutions-llc/go-starter/internal/templates"
	"github.com/webpoint-solutions-llc/go-starter/internal/types"
	"github.com/webpoint-solutions-llc/go-starter/internal/utils"
)

func (s *Service) EmailLogin(ctx context.Context, params dto.LoginParams) (dto.LoginResponse, error) {
	user, err := s.q.GetUserByEmail(ctx, params.Email)

	invalidLoginResponse := dto.LoginResponse{}

	if params.IsAdmin && user.Role != sqlc.UserRoleADMIN {
		return dto.LoginResponse{}, errorhandler.ErrorForbidden(constants.MsgInvalidAdmin)
	}
	if !params.IsAdmin && user.Role == sqlc.UserRoleADMIN {
		return dto.LoginResponse{}, errorhandler.ErrorForbidden(constants.MsgInvalidUserLogin)
	}

	if err != nil {
		if err == pgx.ErrNoRows {
			return invalidLoginResponse, errorhandler.ErrorBadRequest(constants.MsgInvalidCredentials)
		}
		return invalidLoginResponse, err
	}

	if !user.IsEmailVerified {
		return invalidLoginResponse, errorhandler.ErrorBadRequest(constants.MsgEmailNotVerified)
	}

	if user.Password == nil {
		return invalidLoginResponse, errorhandler.ErrorBadRequest(constants.MsgPasswordNotSet)
	}

	wrongPasswordAttemptKey := params.Email + "wrong_password_attempt"

	ok, err := argon2.VerifyEncoded([]byte(*params.Password), []byte(*user.Password))

	if !ok || err != nil {
		if err := s.checkForWrongPasswordAttempt(ctx, wrongPasswordAttemptKey); err != nil {
			return invalidLoginResponse, errorhandler.ErrorBadRequest(constants.MsgMaxWrongPasswordAttempt)
		}
		return invalidLoginResponse, errorhandler.ErrorBadRequest(constants.MsgInvalidCredentials)
	}
	if user.Status == "DEACTIVATED" && user.Status != "SUSPENDED" {
		_, err := s.q.ChangeAccountStatus(ctx, sqlc.ChangeAccountStatusParams{
			ID:     user.ID,
			Status: sqlc.UserStatus("ACTIVE"),
		})
		if err != nil {
			return invalidLoginResponse, errorhandler.ErrorBadRequest(constants.MsgFailedActivate)
		}
	}

	plainToken, tokenHash, err := utils.GenerateRefreshToken()
	if err != nil {
		return invalidLoginResponse, err
	}

	session, err := s.q.CreateUserSession(ctx, sqlc.CreateUserSessionParams{
		UserID:           user.ID,
		RefreshTokenHash: tokenHash,
		UserAgent:        params.UserAgent,
		IpAddress:        params.IPAddress,
		ExpiresAt:        time.Now().Add(time.Duration(config.Cfg.RefreshTokenDuration) * time.Second),
	})
	if err != nil {
		return invalidLoginResponse, err
	}

	accessToken, err := utils.GenerateJWT(types.JWTPlayload{
		UserID:    user.ID.String(),
		TokenType: types.TokenTypeAccess,
		Duration:  config.Cfg.AccessTokenDuration,
		Sid:       session.ID.String(),
		Role:      types.Role(user.Role),
	})
	if err != nil {
		s.logger.Error("Failed to generate access token", "err", err)
		return invalidLoginResponse, errorhandler.ErrorInternal("Failed to generate access token")
	}
	resp := dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: plainToken,
		IsFirstLogin: session.FirstLogin.Bool,
	}

	return resp, nil
}

func (s *Service) EmailSignUp(ctx context.Context, params dto.EmailSignUpRequest) (message string, err error) {
	argon := argon2.DefaultConfig()
	passwordHash, err := argon.HashEncoded([]byte(params.Password))
	if err != nil {
		return "", errorhandler.ErrorInternal("Failed to hash password")
	}

	stringPassword := string(passwordHash)

	args := sqlc.CreateUserParams{
		Name:        params.Name,
		PhoneNumber: &params.PhoneNumber,
		Email:       params.Email,
		Password:    &stringPassword,
	}

	res, err := s.q.CreateUser(ctx, args)
	if err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return "", errorhandler.ErrorConflict(constants.MsgUserAlreadyExists)
			}
		} else {
			return "", err
		}
		return "", err
	}

	// generate both email_verify token
	token, err := utils.GenerateJWT(types.JWTPlayload{
		UserID:    res.ID.String(),
		TokenType: types.TokenTypeVerify,
		Duration:  config.Cfg.EmailVerifyDuration,
	})
	if err != nil {
		return "", err
	}
	// TODO: email send part queue
	go func() {
		// Send verification email
		body, err := templates.LoadAndExecuteTemplate("email_verify.html", dto.EmailVerificationData{
			Name:             res.Name,
			VerificationLink: config.Cfg.BackendURL + "/api/v1/auth/email-verify?token=" + token,
		})
		// Ingore if email is not sent
		if err != nil {
			s.logger.Error("Failed to send email", "err", err)
		}

		err = utils.SendEmail(res.Email, "Verify your Email Address", body)
		// Ingore if email is not sent
		if err != nil {
			s.logger.Error("Failed to send email", "err", err)
		}
	}()

	go func() {
		usermeta := map[string]string{
			"id": res.ID.String(),
		}
		id, err := s.CreateOrGetStripeCustomer(res.Email, res.Name, usermeta)
		if err != nil {
			s.logger.Error("Failed to Create Stripe User", "err", err)
		}

		if id != "" {

			user := dto.BaiscUserInfo{
				StripeCustomerID: &id,
			}

			params := dto.UpdateUserInfoParams{
				User: user,
			}

			updateErr := s.UpdateUser(context.Background(), res.ID, params)
			if updateErr != nil {
				s.logger.Debug("Error on updating user: ", "err", updateErr.Error())
			} else {
				s.logger.Debug("Update stripe user id")
			}
		}
	}()

	return "User created successfully", nil
}

func (s *Service) ResendEmailVerification(ctx context.Context, email string) (message string, err error) {
	user, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errorhandler.ErrorBadRequest(constants.MsgUserNotFound)
		}

		return "", err
	}

	if user.IsEmailVerified {
		return "", errorhandler.ErrorBadRequest(constants.MsgAlreadyVerified)
	}

	// generate both email_verify token
	token, err := utils.GenerateJWT(types.JWTPlayload{
		UserID:    user.ID.String(),
		TokenType: types.TokenTypeVerify,
		Duration:  config.Cfg.EmailVerifyDuration,
	})
	if err != nil {
		return "", err
	}

	// TODO: email send part queue
	// Send verification email
	body, err := templates.LoadAndExecuteTemplate("email_verify.html", dto.EmailVerificationData{
		Name:             user.Name,
		VerificationLink: config.Cfg.BackendURL + "/api/v1/auth/email-verify?token=" + token,
	})
	// Ingore if email is not sent
	if err != nil {
		s.logger.Error("Failed to send email", "err", err)
	}

	err = utils.SendEmail(user.Email, "Verify your Email Address!", body)
	// Ingore if email is not sent
	if err != nil {
		s.logger.Error("Failed to send email", "err", err)
	}

	return "Email Sent successfully", nil
}

func (s *Service) InserSocialLoginUser(ctx context.Context, params dto.LoginParams) (dto.LoginResponse, error) {
	var user sqlc.GetUserByEmailRow
	var err error

	if params.IsAdmin && user.Role != sqlc.UserRoleADMIN {
		return dto.LoginResponse{}, errorhandler.ErrorForbidden(constants.MsgInvalidAdmin)
	}
	if !params.IsAdmin && user.Role == sqlc.UserRoleADMIN {
		return dto.LoginResponse{}, errorhandler.ErrorForbidden(constants.MsgInvalidUserLogin)
	}

	// Try to get the user by email if email is provided
	if params.Email != "" {
		user, err = s.q.GetUserByEmail(ctx, params.Email)
	}

	// If not found by email, or email is empty, try by AppleID (for Apple login)
	if (err != nil && errors.Is(err, sql.ErrNoRows)) || params.Email == "" {
		if params.AppleID != nil {
			userByApple, errApple := s.q.GetUserByAppleID(ctx, params.AppleID)
			if errApple == nil {
				user = sqlc.GetUserByEmailRow{
					ID:    userByApple.ID,
					Name:  userByApple.Name,
					Email: userByApple.Email,
				}
				err = nil
			}
		}
	}

	// If still not found, create the user
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		createParams := sqlc.CreateUserParams{
			Name:            "",
			Email:           params.Email,
			IsEmailVerified: true,
		}

		if params.Name != nil {
			createParams.Name = *params.Name
		}
		if params.AppleID != nil {
			createParams.AppleID = params.AppleID
		}
		if params.GoogleID != nil {
			createParams.GoogleID = params.GoogleID
		}

		newUserResp, err := s.q.CreateUser(ctx, createParams)
		if err != nil {
			return dto.LoginResponse{}, err
		}

		user = sqlc.GetUserByEmailRow{
			ID:    newUserResp.ID,
			Name:  newUserResp.Name,
			Email: newUserResp.Email,
			Role:  newUserResp.Role,
		}
	} else if err != nil {
		// Some other DB error
		return dto.LoginResponse{}, err
	}

	if params.Name != nil && user.Name == "" {
		_ = s.q.UpdateUser(ctx, sqlc.UpdateUserParams{
			ID:   user.ID,
			Name: params.Name,
		})
		user.Name = *params.Name
	}
	if user.Status == "DEACTIVATED" && user.Status != "SUSPENDED" {
		_, err := s.q.ChangeAccountStatus(ctx, sqlc.ChangeAccountStatusParams{
			ID:     user.ID,
			Status: sqlc.UserStatus("ACTIVE"),
		})
		if err != nil {
			return dto.LoginResponse{}, errorhandler.ErrorBadRequest(constants.MsgFailedActivate)
		}
	}

	// Generate refresh token
	plainToken, tokenHash, err := utils.GenerateRefreshToken()
	if err != nil {
		return dto.LoginResponse{}, err
	}

	// create stripe user
	go func() {
		usermeta := map[string]string{
			"id": user.ID.String(),
		}
		stripeID, err := s.CreateOrGetStripeCustomer(user.Email, user.Name, usermeta)
		if err != nil {
			s.logger.Error("Failed to Create Stripe User", "err", err)
		}

		if stripeID != "" {
			userPayload := dto.BaiscUserInfo{
				StripeCustomerID: &stripeID,
			}
			params := dto.UpdateUserInfoParams{
				User: userPayload,
			}
			updateErr := s.UpdateUser(context.Background(), user.ID, params)
			if updateErr != nil {
				s.logger.Debug("Error on updating user: ", "err", updateErr.Error())
			} else {
				s.logger.Debug("Update user")
			}
		}
	}()

	// Create session
	sessionParams := sqlc.CreateUserSessionParams{
		UserID:           user.ID,
		RefreshTokenHash: tokenHash,
		UserAgent:        params.UserAgent,
		IpAddress:        params.IPAddress,
		ExpiresAt:        time.Now().Add(time.Duration(config.Cfg.RefreshTokenDuration) * time.Second),
	}
	session, err := s.q.CreateUserSession(ctx, sessionParams)
	if err != nil {
		return dto.LoginResponse{}, err
	}

	// Generate access token
	accessToken, err := utils.GenerateJWT(types.JWTPlayload{
		UserID:    user.ID.String(),
		TokenType: types.TokenTypeAccess,
		Duration:  config.Cfg.AccessTokenDuration,
		Sid:       session.ID.String(),
		Role:      types.Role(user.Role),
	})
	if err != nil {
		return dto.LoginResponse{}, errorhandler.ErrorInternal(constants.MsgUnableToCreateToken)
	}
	// Return login response
	return dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: plainToken,
		IsFirstLogin: session.FirstLogin.Bool,
	}, nil
}

func (s *Service) VerifyUserEmail(ctx context.Context, token string) (res sqlc.UpdateEmailStatusRow, err error) {
	claims, err := utils.VerifyJWT(token)
	if err != nil {
		return res, err
	}
	if claims.TokenType != string(types.TokenTypeVerify) {
		return res, errorhandler.ErrorBadRequest(constants.MsgInvalidTokenType)
	}

	userID, err := uuid.Parse(claims.ID)
	if err != nil {
		return res, err
	}

	user, err := s.q.GetUserByID(ctx, userID)
	if err != nil {
		return res, err
	}

	if user.IsEmailVerified {
		res.Email = user.Email
		return res, errorhandler.ErrorBadRequest(constants.MsgAlreadyVerified)
	}

	args := sqlc.UpdateEmailStatusParams{
		IsEmailVerified: true,
		ID:              userID,
	}

	res, err = s.q.UpdateEmailStatus(ctx, args)
	if err != nil {

		email, err2 := s.q.GetUserEmailByID(ctx, userID)
		if err2 == nil {
			res.Email = email
		}
		return res, err

	}

	return res, nil
}

func (s *Service) PasswordResetLink(ctx context.Context, email string) (res sqlc.CreateResetTokenRow, err error) {
	user, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return res, errorhandler.ErrorBadRequest(constants.MsgUserNotFound)
		}
		return res, errorhandler.ErrorInternal(err.Error())
	}

	if !user.IsEmailVerified {
		return res, errorhandler.ErrorBadRequest(constants.MsgEmailNotVerified)
	}
	key, err := utils.GenerateRandomSecret()
	if err != nil {
		return res, errorhandler.ErrorInternal(err.Error())
	}

	tokenExpires := time.Now().Add(time.Duration(config.Cfg.PasswordResetDuration) * time.Second)

	args := sqlc.CreateResetTokenParams{
		ResetToken:   &key,
		ResetExpires: &tokenExpires,
		Email:        email,
	}

	res, err = s.q.CreateResetToken(ctx, args)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return res, errorhandler.ErrorNotFound("email doesn't exists")
		}
		return res, errorhandler.ErrorInternal(err.Error())
	}
	// TODO : email send part queue
	go func() {
		// Send verification email
		body, err := templates.LoadAndExecuteTemplate("password_reset.html", dto.PasswordResetData{
			Name:              res.Name,
			PasswordResetLink: config.Cfg.FrontendURL + "/reset-password?token=" + *res.ResetToken,
		})
		// Ignore if email is not sent
		if err != nil {
			s.logger.Error("Failed to send email", "err", err)
		}
		err = utils.SendEmail(email, "Reset Your Password", body)
		// Ignore if email is not sent
		if err != nil {
			s.logger.Error("Failed to send email", "err", err)
		}
	}()
	return res, nil
}

func (s *Service) PasswordResetConfirm(ctx context.Context, params dto.PasswordResetConfirmRequest) (res sqlc.UpdatePasswordResetRow, err error) {
	user, err := s.q.GetUserByResetToken(ctx, &params.Token)
	if err != nil {
		if err == pgx.ErrNoRows {
			return res, errorhandler.ErrorBadRequest(constants.MsgPasswordResetLinkExpired)
		}
	}
	currentTime := time.Now()
	expiresAt := *user.ResetExpires

	if currentTime.After(expiresAt) {
		return res, errorhandler.ErrorBadRequest(constants.MsgTokenExpired)
	}

	argon := argon2.DefaultConfig()
	password_hash, err := argon.HashEncoded([]byte(params.Password))
	if err != nil {
		return res, errorhandler.ErrorInternal("Failed to hash password")
	}

	hashedPassword := string(password_hash)

	args := sqlc.UpdatePasswordResetParams{
		Password:   &hashedPassword,
		ID:         user.ID,
		ResetToken: &params.Token,
	}

	res, err = s.q.UpdatePasswordReset(ctx, args)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return res, errorhandler.ErrorNotFound(constants.MsgUserNotFound)
		}
		return res, errorhandler.ErrorInternal(err)
	}
	return res, nil
}

func (s *Service) GetAppleLoginURL() (string, error) {
	state, err := utils.GenerateRandomSecret()
	if err != nil {
		return "", err
	}

	return s.appleClient.CreateCallbackURL(state), nil
}

func (s *Service) ExchangeCodeWithApple(ctx context.Context, code string) (*types.AppleIDTokenClaims, error) {
	appleToken, err := s.appleClient.GetAppleToken(code, 10*time.Minute)
	if err != nil {
		return nil, err
	}

	var claims types.AppleIDTokenClaims
	_, _, err = jwt.NewParser().ParseUnverified(appleToken, &claims)
	if err != nil {
		return nil, err
	}

	name := extractNameFromClaims(claims)

	claims.Name = name

	if claims.Email == "" {
		claims.Email = claims.Subject + "@apple.com"
	}

	if claims.Name == "" {
		claims.Name = "Apple User"
	}

	return &claims, nil
}

func (s *Service) CleanupSession(ctx context.Context, sessionID uuid.UUID) (*uuid.UUID, error) {
	userID, err := s.q.RevokeSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	go func() {
		err = s.q.DeleteOldSessions(context.Background())
		// Ignore
		if err != nil {
			s.logger.Error("Failed to cleanup sessions", "err", err)
		}
	}()

	return &userID, nil
}

func (s *Service) UpdatePassword(ctx context.Context, params dto.PasswordUpdateRequest, userID uuid.UUID) error {
	argon := argon2.DefaultConfig()
	storedPassword, err := s.q.GetPasswordByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user password")
		return err
	}
	if storedPassword != nil && *storedPassword != "" {
		ok, err := argon2.VerifyEncoded([]byte(params.CurrentPassword), []byte(*storedPassword))
		if !ok || err != nil {
			return errorhandler.ErrorBadRequest(constants.MsgInvalidCredentials)
		}
	}
	passwordHash, err := argon.HashEncoded([]byte(params.NewPassword))
	if err != nil {
		s.logger.Error("Failed to generate a hash password")
		return err
	}
	stringPassword := string(passwordHash)
	arg := sqlc.UpdatePasswordByIDParams{
		Password: &stringPassword,
		ID:       userID,
	}

	err = s.q.UpdatePasswordByID(ctx, arg)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) PasswordResetTokenCheck(ctx context.Context, params dto.PasswordResetTokenCheck) (res sqlc.UpdatePasswordResetRow, err error) {
	user, err := s.q.GetUserByResetToken(ctx, &params.Token)
	if err != nil {
		if err == pgx.ErrNoRows {
			return res, errorhandler.ErrorBadRequest(constants.MsgPasswordResetLinkExpired)
		}
	}
	currentTime := time.Now()
	expiresAt := *user.ResetExpires

	if currentTime.After(expiresAt) {
		return res, errorhandler.ErrorBadRequest(constants.MsgTokenExpired)
	}
	return res, nil
}

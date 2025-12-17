package constants

// User Authentication
const (
	MsgRegistrationSuccessful = "Registration Successful! Welcome to The Model's Resource."
	MsgSignInSuccessful       = "Sign in successful!"

	MsgSignOutSuccessful           = "Sign out successful."
	MsgPasswordResetEmailSent      = "Password reset email sent. Please check your inbox."
	MsgPasswordUpdatedSuccessfully = "Password updated successfully."
	MsgRegistrationFailed          = "Registration failed. Please check your details."
	MsgSignInFailed                = "Sign in failed. Incorrect email or password."
	MsgUserAlreadyExists           = "Account already exists with this email."
	MsgEmailVerificationSent       = "Email verification link sent. Please verify your email."
	MsgEmailVerifiedSuccessfully   = "Email verified successfully!"
	MsgEmailVerificationFailed     = "Email verification failed. Link may be expired or invalid."
	MsgPasswordNotSet              = "Password not set, please use Login with Google/Apple"
	MsgPasswordResetLinkExpired    = "This link has expired. Please request a new password reset link."
	// need for review
	MsgInvalidCredentials = "Invalid credentials. Please try again."
	MsgInvalidUserLogin   = "Invalid credentials. Cannot use admin credentials."

	MsgInvalidAdmin        = "Invaild Admin credential."
	MsgEmailIsAlreadyInUse = "Account already exists with this email."
	MsgFailedActivate      = "Failed to activate account. Please try again later."
	MsgEmailNotVerified    = "Email not verified. Please verify your email to continue."
	MsgUserAlreadyExist    = "User already exists with this email."

	MsgInvalidTokenType = "Invalid token type provided."
	MsgValidToken       = "Token is valid."

	MsgUserNotFound = "User not found. Please check the details and try again."

	MsgTokenExpired = "Token expired. Please request a new one."

	MsgUnableToCreateToken     = "Failed to generate access token. Please try again."
	MsgMaxWrongPasswordAttempt = "Maximum password retry limit reached. Please try again after 60 seconds."
	MsgAlreadyVerified         = "Email is already verified."
	MsgReLogin                 = "Invalid user ID, Need to login again"
)

package apperror

const (
	MsgInvalidCredentials      = "Invalid credentials" //nolint:gosec
	MsgLoginRequestRequired    = "Login request is required"
	MsgRegisterRequestRequired = "Register request is required"
	MsgUsernameAlreadyInUse    = "Username already in use"
	MsgEmailAlreadyInUse       = "Email already in use"
	MsgRefreshRequestRequired  = "Refresh request is required"
	MsgRefreshTokenRequired    = "Refresh token is required"
	MsgSessionNotFound         = "Session not found"
	MsgSessionNotActive        = "Session is not active"
	MsgUserNotFound            = "User not found"
	MsgSessionExpiredOrRevoked = "Session expired or revoked"
	MsgAccountAccessRevoked    = "Account access has been revoked"
	MsgOperationFailed         = "Operation failed"
)

const (
	MsgGetUserByUsername    = "get user by username"
	MsgGetUserByEmail       = "get user by email"
	MsgGetSession           = "get session"
	MsgGetUser              = "get user"
	MsgGenerateRefreshToken = "generate refresh token"
	MsgHashRefreshToken     = "hash refresh token"
	MsgRotateSession        = "rotate session"
	MsgUpdateSession        = "update session"
	MsgSaveNewSession       = "save new session"
	MsgRevokeSession        = "revoke session"
	MsgGenerateAccessToken  = "generate access token"
)

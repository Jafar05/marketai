package dto

type ValidateTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

type ValidateTokenResponse struct {
	Valid  bool   `json:"valid"`
	UserID string `json:"user_id,omitempty"`
	Role   string `json:"role,omitempty"`
}

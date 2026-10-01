package dto

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=100"`
	Password string `json:"password" binding:"required,min=8,max=100"`
}

type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int64        `json:"expires_in"`
	User        UserResponse `json:"user"`
}

type UserResponse struct {
	ID             int64   `json:"id"`
	RoleID         int     `json:"role_id"`
	WorkScheduleID *int    `json:"work_schedule_id"`
	NIP            *string `json:"nip"`
	FullName       string  `json:"full_name"`
	Email          string  `json:"email"`
	PhoneNumber    string  `json:"phone_number"`
	IsActive       bool    `json:"is_active"`
}

type SetupAdminRequest struct {
	NIP         *string `json:"nip"`
	FullName    string  `json:"full_name" binding:"required,max=150"`
	Email       string  `json:"email" binding:"required,email,max=100"`
	Password    string  `json:"password" binding:"required,min=8,max=100"`
	PhoneNumber string  `json:"phone_number" binding:"required,max=20"`
}

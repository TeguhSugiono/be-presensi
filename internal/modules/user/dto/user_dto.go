package dto

type CreateUserRequest struct {
	RoleID         int     `json:"role_id" binding:"required"`
	WorkScheduleID *int    `json:"work_schedule_id"`
	NIP            *string `json:"nip"`
	FullName       string  `json:"full_name" binding:"required,max=150"`
	Email          string  `json:"email" binding:"required,email,max=100"`
	Password       string  `json:"password" binding:"required,min=8,max=100"`
	PhoneNumber    string  `json:"phone_number" binding:"required,max=20"`
	IsActive       bool    `json:"is_active"`
}

type UpdateUserRequest struct {
	RoleID         int     `json:"role_id" binding:"required"`
	WorkScheduleID *int    `json:"work_schedule_id"`
	NIP            *string `json:"nip"`
	FullName       string  `json:"full_name" binding:"required,max=150"`
	Email          string  `json:"email" binding:"required,email,max=100"`
	PhoneNumber    string  `json:"phone_number" binding:"required,max=20"`
	IsActive       bool    `json:"is_active"`
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

type ChangePasswordRequest struct {
	NewPassword     string `json:"new_password" binding:"required,min=8,max=100"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

package dto

type CreateRoleRequest struct {
	Name      string `json:"name" binding:"required,max=50"`
	IsTeacher bool   `json:"is_teacher"`
}

type UpdateRoleRequest struct {
	Name      string `json:"name" binding:"required,max=50"`
	IsTeacher bool   `json:"is_teacher"`
}

type RoleResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	IsTeacher bool   `json:"is_teacher"`
}

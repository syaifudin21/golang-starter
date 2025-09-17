package dtos

type UpdateAccountRequest struct {
	Name  string  `json:"name" validate:"required"`
	Phone *string `json:"phone"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

type UpdateUserRoleRequest struct {
	Role string `json:"role" validate:"required"`
}
package models

type Role string

const (
	RoleEmployee  Role = "employee"
	RoleModerator Role = "moderator"
)

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  Role   `json:"role"`
}

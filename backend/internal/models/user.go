package models

import "time"

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleSubadmin Role = "subadmin"
	RoleUser     Role = "user"
)

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Phone        string    `json:"phone"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

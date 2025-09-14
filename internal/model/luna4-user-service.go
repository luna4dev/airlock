package model

type Luna4Service string

const (
	Luna4ServicePrunk Luna4Service = "PRUNK"
)

type UserServicePermission string

const (
	UserServiceSuperUser UserServicePermission = "SUPER_USER"
	UserServiceUser      UserServicePermission = "USER"
)

type Luna4UserService struct {
	ID         string                `json:"id"`
	UserID     string                `json:"userId"`
	Service    Luna4Service          `json:"service"`
	Permission UserServicePermission `json:"permission"`
	ExpiresAt  *int64                `json:"expiresAt,omitempty"`
}

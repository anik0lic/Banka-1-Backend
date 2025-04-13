package dto

type ActuaryDTO struct {
	UserID       uint    `json:"userID" validate:"required"`
	Department   string  `json:"department" validate:"required,oneof=supervisor agent"`
	LimitAmount  float64 `json:"limitAmount,omitempty"`
	UsedLimit    float64 `json:"usedLimit,omitempty"`
	NeedApproval bool    `json:"needApproval"`
}

type FilteredActuaryDTO struct {
	ID           uint    `json:"id"`
	FirstName    string  `json:"firstName"`
	LastName     string  `json:"lastName"`
	Email        string  `json:"email"`
	LimitAmount  float64 `json:"limitAmount"`
	UsedLimit    float64 `json:"usedLimit"`
	NeedApproval bool    `json:"needApproval"`
	Position     string  `json:"position"`
	Department   string  `json:"department"`
}

type FilteredActuaryResponse struct {
	Data    []FilteredActuaryDTO `json:"data"`
	Success bool                 `json:"success"`
}

type EmployeeResponse struct {
	ID          uint     `json:"id"`
	FirstName   string   `json:"firstName"`
	LastName    string   `json:"lastName"`
	Username    string   `json:"username"`
	BirthDate   string   `json:"birthDate"`
	Gender      string   `json:"gender"`
	Email       string   `json:"email"`
	PhoneNumber string   `json:"phoneNumber"`
	Address     string   `json:"address"`
	Position    string   `json:"position"`
	Department  string   `json:"department"`
	Active      bool     `json:"active"`
	IsAdmin     bool     `json:"isAdmin"`
	Permissions []string `json:"permissions"`
}

type ActuaryProfitDTO struct {
	FullName string  `json:"fullName"`
	Profit   float64 `json:"profit"`
	Role     string  `json:"role"`
}

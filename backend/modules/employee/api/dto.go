package api

import (
	"time"

	"employee-portal/modules/shared"
)

// Request and response bodies (JSON).

type summaryDTO struct {
	ID           int64       `json:"id"`
	EmployeeCode string      `json:"employeeCode"`
	FirstName    string      `json:"firstName"`
	LastName     string      `json:"lastName"`
	PhotoURL     *string     `json:"photoUrl"`
	Role         shared.Role `json:"role"`
	IsActive     bool        `json:"isActive"`
}

type listResponse struct {
	Items []summaryDTO `json:"items"`
	Page  struct {
		Page     int  `json:"page"`
		PageSize int  `json:"pageSize"`
		HasNext  bool `json:"hasNext"`
	} `json:"page"`
	Total int64 `json:"total"`
}

type skillDTO struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"isActive"`
}

type detailResponse struct {
	ID           int64       `json:"id"`
	EmployeeCode string      `json:"employeeCode"`
	FirstName    string      `json:"firstName"`
	LastName     string      `json:"lastName"`
	PhotoURL     *string     `json:"photoUrl"`
	Contact      contactDTO  `json:"contact"`
	Skills       []skillDTO  `json:"skills"`
	Version      int32       `json:"version"`
	Role         shared.Role `json:"role"`
	IsActive     bool        `json:"isActive"`
}

type contactDTO struct {
	MobilePhone *string `json:"mobilePhone"`
	Email       *string `json:"email"`
	Address     *string `json:"address"`
}

type topSkillDTO struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	UsedBy int64  `json:"usedBy"`
}

type statsResponse struct {
	Employees        int64         `json:"employees"`
	CompleteProfiles int64         `json:"completeProfiles"`
	TopSkills        []topSkillDTO `json:"topSkills"`
	RecentChanges    []recentDTO   `json:"recentChanges"`
}

type recentDTO struct {
	ID            int64       `json:"id"`
	EmployeeID    int64       `json:"employeeId"`
	EmployeeCode  string      `json:"employeeCode"`
	Name          string      `json:"name"`
	Field         string      `json:"field"`
	ChangedAt     time.Time   `json:"changedAt"`
	ChangedByRole shared.Role `json:"changedByRole"`
}

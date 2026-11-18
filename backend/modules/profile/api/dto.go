package api

// Request and response bodies (JSON).

type contactDTO struct {
	MobilePhone *string `json:"mobilePhone"`
	Email       *string `json:"email"`
	Address     *string `json:"address"`
}

type skillDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type profileResponse struct {
	ID           int64      `json:"id"`
	EmployeeCode string     `json:"employeeCode"`
	FirstName    string     `json:"firstName"`
	LastName     string     `json:"lastName"`
	PhotoURL     *string    `json:"photoUrl"`
	Contact      contactDTO `json:"contact"`
	Skills       []skillDTO `json:"skills"`
	Version      int32      `json:"version"`
}

type updateContactRequest struct {
	Value *string `json:"value" binding:"required"`
}

type setSkillsRequest struct {
	SkillIDs []int64 `json:"skillIds" binding:"required"` // [] clears all skills; missing is an error
}

type versionResponse struct {
	Version int32 `json:"version"`
}

type photoResponse struct {
	Version  int32  `json:"version"`
	PhotoURL string `json:"photoUrl"`
}

type updateNameRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

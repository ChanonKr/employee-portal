package api

// Request and response bodies (JSON).

type skillResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type adminSkillResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	IsActive  bool   `json:"isActive"`
	SortOrder int32  `json:"sortOrder"`
	UsedBy    int64  `json:"usedBy"`
}

type createRequest struct {
	Name string `json:"name" binding:"required"`
}

type updateRequest struct {
	Name     *string `json:"name"`
	IsActive *bool   `json:"isActive"`
}

type orderItem struct {
	ID        int64  `json:"id" binding:"required"`
	SortOrder *int32 `json:"sortOrder" binding:"required"`
}

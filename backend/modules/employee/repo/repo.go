package repo

import (
	"context"

	"gorm.io/gorm"

	"employee-portal/modules/employee/domain"
	"employee-portal/modules/employee/model"
	"employee-portal/modules/employee/service"
	historydomain "employee-portal/modules/history/domain"
	profiledomain "employee-portal/modules/profile/domain"
	"employee-portal/modules/shared"
	skilldomain "employee-portal/modules/skill/domain"
	"employee-portal/pkg/database"
)

// Repository implements service.Repository with GORM.
type Repository struct {
	db *gorm.DB
}

var _ service.Repository = (*Repository)(nil)

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// withUser joins each employee to their account (code, role, active).
func withUser(db *gorm.DB) *gorm.DB {
	return db.Table("employees").Joins("JOIN users ON users.id = employees.user_id")
}

// matching filters by code or full name. The pattern is escaped in EmployeeQuery.LikePattern.
func matching(pattern *string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if pattern == nil {
			return db
		}
		return db.Where("users.employee_code ILIKE @p OR (employees.first_name || ' ' || employees.last_name) ILIKE @p",
			map[string]any{"p": *pattern})
	}
}

func (r *Repository) List(ctx context.Context, q domain.EmployeeQuery) ([]domain.EmployeeSummary, int64, error) {
	db := r.db.WithContext(ctx)
	filter := matching(q.LikePattern())

	var total int64
	if err := db.Scopes(withUser, filter).Count(&total).Error; err != nil {
		return nil, 0, database.MapError(err)
	}

	var rows []model.SummaryRow
	err := db.Scopes(withUser, filter).
		Select("employees.id, users.employee_code, employees.first_name, employees.last_name, employees.photo_path, users.role, users.is_active").
		Order("employees.first_name").Order("employees.last_name").Order("employees.id").
		Limit(q.PageSize).Offset(q.Offset()).
		Find(&rows).Error
	if err != nil {
		return nil, 0, database.MapError(err)
	}

	items := make([]domain.EmployeeSummary, len(rows))
	for i, row := range rows {
		items[i] = domain.EmployeeSummary{
			ID: row.ID, EmployeeCode: row.EmployeeCode, FirstName: row.FirstName, LastName: row.LastName,
			PhotoPath: row.PhotoPath, Role: shared.Role(row.Role), IsActive: row.IsActive,
		}
	}
	return items, total, nil
}

func (r *Repository) Get(ctx context.Context, id int64) (domain.EmployeeDetail, error) {
	db := r.db.WithContext(ctx)
	var row model.DetailRow
	err := db.Scopes(withUser).
		Select("employees.*, users.employee_code, users.role, users.is_active").
		Where("employees.id = ?", id).
		Take(&row).Error
	if err != nil {
		return domain.EmployeeDetail{}, database.MapError(err)
	}
	var skills []model.SkillRow
	err = db.Table("skills").Select("skills.id, skills.name, skills.is_active").
		Joins("JOIN employee_skills es ON es.skill_id = skills.id").
		Where("es.employee_id = ?", id).
		Order("skills.sort_order").Order("skills.name").
		Find(&skills).Error
	if err != nil {
		return domain.EmployeeDetail{}, database.MapError(err)
	}
	d := domain.EmployeeDetail{
		Profile: profiledomain.Profile{
			Employee: profiledomain.Employee{
				ID: row.ID, UserID: row.UserID, FirstName: row.FirstName, LastName: row.LastName,
				PhotoPath: row.PhotoPath, MobilePhone: row.MobilePhone, ContactEmail: row.ContactEmail,
				Address: row.Address, Version: row.Version,
			},
			EmployeeCode: row.EmployeeCode,
			Skills:       make([]skilldomain.Skill, len(skills)),
		},
		Role:     shared.Role(row.Role),
		IsActive: row.IsActive,
	}
	for i, s := range skills {
		d.Skills[i] = skilldomain.Skill{ID: s.ID, Name: s.Name, IsActive: s.IsActive}
	}
	return d, nil
}

func (r *Repository) Stats(ctx context.Context, topSkills, recentChanges int) (domain.DashboardStats, error) {
	db := r.db.WithContext(ctx)

	// "Complete" matches the profile checklist: photo, all contact fields, at least one skill.
	var counts struct{ Total, Complete int64 }
	err := db.Table("employees").Select(`count(*) AS total,
		count(*) FILTER (
		  WHERE photo_path IS NOT NULL AND mobile_phone IS NOT NULL
		    AND contact_email IS NOT NULL AND address IS NOT NULL
		    AND EXISTS (SELECT 1 FROM employee_skills es WHERE es.employee_id = employees.id)
		) AS complete`).Take(&counts).Error
	if err != nil {
		return domain.DashboardStats{}, database.MapError(err)
	}

	var top []model.TopSkillRow
	err = db.Table("employee_skills").
		Select("skills.id, skills.name, count(*) AS used_by").
		Joins("JOIN skills ON skills.id = employee_skills.skill_id").
		Group("skills.id").
		Order("used_by DESC").Order("skills.name").
		Limit(topSkills).
		Find(&top).Error
	if err != nil {
		return domain.DashboardStats{}, database.MapError(err)
	}

	var recent []model.RecentChangeRow
	err = db.Table("profile_change_logs").
		Select("profile_change_logs.id, profile_change_logs.employee_id, u.employee_code, e.first_name, e.last_name, " +
			"profile_change_logs.field, profile_change_logs.changed_at, cb.role AS changed_by_role").
		Joins("JOIN employees e ON e.id = profile_change_logs.employee_id").
		Joins("JOIN users u ON u.id = e.user_id").
		Joins("JOIN users cb ON cb.id = profile_change_logs.changed_by").
		Order("profile_change_logs.changed_at DESC").Order("profile_change_logs.id DESC").
		Limit(recentChanges).
		Find(&recent).Error
	if err != nil {
		return domain.DashboardStats{}, database.MapError(err)
	}

	s := domain.DashboardStats{
		Employees:        counts.Total,
		CompleteProfiles: counts.Complete,
		TopSkills:        make([]domain.SkillCount, len(top)),
		RecentChanges:    make([]domain.RecentChange, len(recent)),
	}
	for i, t := range top {
		s.TopSkills[i] = domain.SkillCount{Skill: skilldomain.Skill{ID: t.ID, Name: t.Name}, UsedBy: t.UsedBy}
	}
	for i, c := range recent {
		s.RecentChanges[i] = domain.RecentChange{
			ID: c.ID, EmployeeID: c.EmployeeID, EmployeeCode: c.EmployeeCode,
			FirstName: c.FirstName, LastName: c.LastName, Field: historydomain.FieldName(c.Field),
			ChangedAt: c.ChangedAt, ChangedByRole: shared.Role(c.ChangedByRole),
		}
	}
	return s, nil
}

package employee

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"employee-portal/modules/entity"
	"employee-portal/pkg/database"
	"employee-portal/pkg/database/sqlc"
)

type repository struct {
	q *sqlc.Queries
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{q: sqlc.New(pool)}
}

func (r *repository) List(ctx context.Context, q entity.EmployeeQuery) ([]entity.EmployeeSummary, int64, error) {
	pattern := q.LikePattern()
	rows, err := r.q.AdminListEmployees(ctx, sqlc.AdminListEmployeesParams{
		Pattern:    pattern,
		PageOffset: int32(q.Offset()),
		PageLimit:  int32(q.PageSize),
	})
	if err != nil {
		return nil, 0, database.MapError(err)
	}
	total, err := r.q.AdminCountEmployees(ctx, pattern)
	if err != nil {
		return nil, 0, database.MapError(err)
	}
	items := make([]entity.EmployeeSummary, len(rows))
	for i, row := range rows {
		items[i] = entity.EmployeeSummary{
			ID: row.ID, EmployeeCode: row.EmployeeCode, FirstName: row.FirstName, LastName: row.LastName,
			PhotoPath: row.PhotoPath, Role: entity.Role(row.Role), IsActive: row.IsActive,
		}
	}
	return items, total, nil
}

func (r *repository) Get(ctx context.Context, id int64) (entity.EmployeeDetail, error) {
	row, err := r.q.AdminGetEmployee(ctx, id)
	if err != nil {
		return entity.EmployeeDetail{}, database.MapError(err)
	}
	skills, err := r.q.ListEmployeeSkills(ctx, id)
	if err != nil {
		return entity.EmployeeDetail{}, database.MapError(err)
	}
	e := row.Employee
	d := entity.EmployeeDetail{
		Profile: entity.Profile{
			Employee: entity.Employee{
				ID: e.ID, UserID: e.UserID, FirstName: e.FirstName, LastName: e.LastName,
				PhotoPath: e.PhotoPath, MobilePhone: e.MobilePhone, ContactEmail: e.ContactEmail,
				Address: e.Address, Version: e.Version,
			},
			EmployeeCode: row.EmployeeCode,
			Skills:       make([]entity.Skill, len(skills)),
		},
		Role:     entity.Role(row.Role),
		IsActive: row.IsActive,
	}
	for i, s := range skills {
		d.Skills[i] = entity.Skill{ID: s.ID, Name: s.Name, IsActive: s.IsActive}
	}
	return d, nil
}

func (r *repository) Stats(ctx context.Context, topSkills, recentChanges int) (entity.DashboardStats, error) {
	counts, err := r.q.AdminCountProfiles(ctx)
	if err != nil {
		return entity.DashboardStats{}, database.MapError(err)
	}
	top, err := r.q.AdminTopSkills(ctx, int32(topSkills))
	if err != nil {
		return entity.DashboardStats{}, database.MapError(err)
	}
	recent, err := r.q.AdminRecentChanges(ctx, int32(recentChanges))
	if err != nil {
		return entity.DashboardStats{}, database.MapError(err)
	}

	s := entity.DashboardStats{
		Employees:        counts.Total,
		CompleteProfiles: counts.Complete,
		TopSkills:        make([]entity.SkillCount, len(top)),
		RecentChanges:    make([]entity.RecentChange, len(recent)),
	}
	for i, t := range top {
		s.TopSkills[i] = entity.SkillCount{Skill: entity.Skill{ID: t.ID, Name: t.Name}, UsedBy: t.UsedBy}
	}
	for i, c := range recent {
		s.RecentChanges[i] = entity.RecentChange{
			ID: c.ID, EmployeeID: c.EmployeeID, EmployeeCode: c.EmployeeCode,
			FirstName: c.FirstName, LastName: c.LastName, Field: entity.FieldName(c.Field),
			ChangedAt: c.ChangedAt, ChangedByRole: entity.Role(c.ChangedByRole),
		}
	}
	return s, nil
}

package server

import (
	authapi "employee-portal/modules/auth/api"
	authrepository "employee-portal/modules/auth/repository"
	authservice "employee-portal/modules/auth/service"
	employeeapi "employee-portal/modules/employee/api"
	employeerepository "employee-portal/modules/employee/repository"
	employeeservice "employee-portal/modules/employee/service"
	fileapi "employee-portal/modules/file/api"
	historyapi "employee-portal/modules/history/api"
	historyrepository "employee-portal/modules/history/repository"
	historyservice "employee-portal/modules/history/service"
	profileapi "employee-portal/modules/profile/api"
	profilerepository "employee-portal/modules/profile/repository"
	profileservice "employee-portal/modules/profile/service"
	settingapi "employee-portal/modules/setting/api"
	settingrepository "employee-portal/modules/setting/repository"
	settingservice "employee-portal/modules/setting/service"
	skillapi "employee-portal/modules/skill/api"
	skillrepository "employee-portal/modules/skill/repository"
	skillservice "employee-portal/modules/skill/service"
	"employee-portal/pkg/imaging"
)

// apis holds the HTTP adapter of every feature.
type apis struct {
	auth     *authapi.API
	profile  *profileapi.API
	history  *historyapi.API
	skill    *skillapi.API
	setting  *settingapi.API
	employee *employeeapi.API
	file     *fileapi.API
}

// wire builds each feature: repository → service → api. This and routes.go are the
// only places that know about every feature.
func (s *Server) wire() apis {
	return apis{
		auth: authapi.New(
			authservice.New(authrepository.New(s.db), s.tokens),
			authapi.CookieOptions{Secure: s.cfg.IsProduction()},
		),
		profile: profileapi.New(profileservice.New(profileservice.Deps{
			Repo:   profilerepository.New(s.db),
			Photos: imaging.NewProcessor(s.cfg.UploadMaxBytes),
			Files:  s.files,
			Log:    s.log,
		})),
		history:  historyapi.New(historyservice.New(historyrepository.New(s.db), s.cfg.HistoryRetentionDays)),
		skill:    skillapi.New(skillservice.New(skillrepository.New(s.db))),
		setting:  settingapi.New(settingservice.New(settingrepository.New(s.db), s.cfg.HistoryRetentionDays)),
		employee: employeeapi.New(employeeservice.New(employeerepository.New(s.db))),
		file:     fileapi.New(s.files),
	}
}

package server

import (
	authapi "employee-portal/modules/auth/api"
	authrepo "employee-portal/modules/auth/repo"
	authservice "employee-portal/modules/auth/service"
	employeeapi "employee-portal/modules/employee/api"
	employeerepo "employee-portal/modules/employee/repo"
	employeeservice "employee-portal/modules/employee/service"
	fileapi "employee-portal/modules/file/api"
	historyapi "employee-portal/modules/history/api"
	historyrepo "employee-portal/modules/history/repo"
	historyservice "employee-portal/modules/history/service"
	profileapi "employee-portal/modules/profile/api"
	profilerepo "employee-portal/modules/profile/repo"
	profileservice "employee-portal/modules/profile/service"
	settingapi "employee-portal/modules/setting/api"
	settingrepo "employee-portal/modules/setting/repo"
	settingservice "employee-portal/modules/setting/service"
	skillapi "employee-portal/modules/skill/api"
	skillrepo "employee-portal/modules/skill/repo"
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

// wire builds each feature: repo → service → api. This and routes.go are the
// only places that know about every feature.
func (s *Server) wire() apis {
	return apis{
		auth: authapi.New(
			authservice.New(authrepo.New(s.db), s.tokens),
			authapi.CookieOptions{Secure: s.cfg.IsProduction()},
		),
		profile: profileapi.New(profileservice.New(profileservice.Deps{
			Repo:   profilerepo.New(s.db),
			Photos: imaging.NewProcessor(s.cfg.UploadMaxBytes),
			Files:  s.files,
			Log:    s.log,
		})),
		history:  historyapi.New(historyservice.New(historyrepo.New(s.db), s.cfg.HistoryDays)),
		skill:    skillapi.New(skillservice.New(skillrepo.New(s.db))),
		setting:  settingapi.New(settingservice.New(settingrepo.New(s.db), s.cfg.HistoryDays)),
		employee: employeeapi.New(employeeservice.New(employeerepo.New(s.db))),
		file:     fileapi.New(s.files),
	}
}

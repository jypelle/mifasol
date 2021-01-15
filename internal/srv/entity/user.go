package entity

import "github.com/jypelle/mifasol/restApiV1"

// User

type UserEntity struct {
	UserId         restApiV1.UserId `db:"user_id"`
	CreationTs     int64            `db:"creation_ts"`
	UpdateTs       int64            `db:"update_ts"`
	Name           string           `db:"name"`
	HideExplicitFg bool             `db:"hide_explicit_fg"`
	AdminFg        bool             `db:"admin_fg"`
	Password       string           `db:"password"`
}

func (e *UserEntity) Fill(s *restApiV1.User) {
	s.Id = e.UserId
	s.CreationTs = e.CreationTs
	s.UpdateTs = e.UpdateTs
	s.Name = e.Name
	s.HideExplicitFg = e.HideExplicitFg
	s.AdminFg = e.AdminFg
}

func (e *UserEntity) LoadMeta(s *restApiV1.UserMeta) {
	if s != nil {
		e.Name = s.Name
		e.HideExplicitFg = s.HideExplicitFg
		e.AdminFg = s.AdminFg
	}
}

type DeletedUserEntity struct {
	UserId   restApiV1.UserId `db:"user_id"`
	DeleteTs int64            `db:"delete_ts"`
}

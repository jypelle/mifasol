package oldentity

import "github.com/jypelle/mifasol/restApiV1"

// User

type UserEntity struct {
	Id             restApiV1.UserId `storm:"id"`
	CreationTs     int64
	UpdateTs       int64  `storm:"index"`
	Name           string `storm:"unique"`
	HideExplicitFg bool
	AdminFg        bool
	Password       string
}

func (e *UserEntity) Fill(s *restApiV1.User) {
	s.Id = e.Id
	s.CreationTs = e.CreationTs
	s.UpdateTs = e.UpdateTs
	s.Name = e.Name
	s.HideExplicitFg = e.HideExplicitFg
	s.AdminFg = e.AdminFg
	s.Password = e.Password
}

func (e *UserEntity) LoadMeta(s *restApiV1.UserMeta) {
	if s != nil {
		e.Name = s.Name
		e.HideExplicitFg = s.HideExplicitFg
		e.AdminFg = s.AdminFg
	}
}

type DeletedUserEntity struct {
	Id       restApiV1.UserId `storm:"id"`
	DeleteTs int64            `storm:"index"`
}

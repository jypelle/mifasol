package entity

import "mifasol/restApiV1"

// User

type UserEntity struct {
	Id         string `storm:"id"`
	CreationTs int64
	UpdateTs   int64  `storm:"index"`
	Name       string `storm:"unique"`
	AdminFg    bool
	Password   string
}

func (e *UserEntity) Fill(s *restApiV1.User) {
	s.Id = e.Id
	s.CreationTs = e.CreationTs
	s.UpdateTs = e.UpdateTs
	s.Name = e.Name
	s.AdminFg = e.AdminFg
}

func (e *UserEntity) LoadMeta(s *restApiV1.UserMeta) {
	if s != nil {
		e.Name = s.Name
		e.AdminFg = s.AdminFg
	}
}

type DeletedUserEntity struct {
	Id       string `storm:"id"`
	DeleteTs int64  `storm:"index"`
}

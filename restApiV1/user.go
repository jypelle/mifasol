package restApiV1

// User

type UserComplete struct {
	User     `storm:"inline"`
	Password string `json:"password"`
}

type User struct {
	Id         string `json:"id" storm:"id"`
	CreationTs int64  `json:"creationTs"`
	UpdateTs   int64  `json:"updateTs" storm:"index"`
	UserMeta   `storm:"inline"`
}

type UserMeta struct {
	Name    string `json:"name" storm:"unique"`
	AdminFg bool   `json:"adminFlag"`
}

type UserMetaComplete struct {
	UserMeta `storm:"inline"`
	Password string `json:"password"`
}

type DeletedUser struct {
	Id       string `json:"id" storm:"id"`
	DeleteTs int64  `json:"deleteTs" storm:"index"`
}

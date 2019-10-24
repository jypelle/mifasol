package restApiV1

// User
type UserId string

type User struct {
	Id         UserId `json:"id"`
	CreationTs int64  `json:"creationTs"`
	UpdateTs   int64  `json:"updateTs"`
	UserMeta
}

type UserMeta struct {
	Name    string `json:"name"`
	AdminFg bool   `json:"adminFlag"`
}

type UserMetaComplete struct {
	UserMeta
	Password string `json:"password"`
}

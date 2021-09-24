package restApiV1

// User

var UndefinedUserId UserId = "xxx"

type UserId string

type User struct {
	Id         UserId `json:"id"`
	CreationTs int64  `json:"creationTs"`
	UpdateTs   int64  `json:"updateTs"`
	Password   string
	UserMeta
}

type UserMeta struct {
	Name           string `json:"name"`
	AdminFg        bool   `json:"adminFg"`
	HideExplicitFg bool   `json:"hideExplicitFg"`
}

type UserMetaComplete struct {
	UserMeta
	Password string `json:"password"`
}

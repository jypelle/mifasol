package model

type MainLayout struct {
	Title            string
	MenuTitle        string
	ServerHostname   string
	ServerPort       int64
	ServerSsl        bool
	ServerSelfSigned bool
}

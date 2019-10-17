package restClientV1

type RestConfig interface {
	GetCompleteConfigCertFilename() string
	GetServerHostname() string
	GetServerPort() int64
	GetServerSsl() bool
	GetServerSelfSigned() bool
	GetTimeout() int64
	GetUsername() string
	GetPassword() string
}

package restClientV1

type RestConfig interface {
	GetCert() []byte
	SetCert(cert []byte) error
	GetServerHostname() string
	GetServerPort() int64
	GetServerSsl() bool
	GetServerSelfSigned() bool
	GetTimeout() int64
	GetUsername() string
	GetPassword() string
}

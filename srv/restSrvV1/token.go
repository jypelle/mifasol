package restSrvV1

import (
	"crypto/rand"
	"fmt"
	"github.com/sirupsen/logrus"
	"mifasol/restApiV1"
	"mifasol/srv/svc"
	"mifasol/tool"
	"net/http"
	"time"
)

type internalToken struct {
	userId     string
	creationTs int64
}

func (s *RestServer) generateToken(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Generate token")

	values := r.URL.Query()

	grantType := values.Get("grant_type")
	name := values.Get("username")
	password := values.Get("password")

	if grantType == "" || name == "" || password == "" {
		s.apiErrorCodeResponse(w, restApiV1.InvalideRequestErrorCode)
		return
	}

	if grantType != "password" {
		s.apiErrorCodeResponse(w, restApiV1.UnsupportedGrantTypeErrorCode)
		return
	}

	user, err := s.service.ReadUserEntityByUserName(nil, name)
	if err != nil {
		if err == svc.ErrNotFound {
			s.apiErrorCodeResponse(w, restApiV1.InvalideGrantErrorCode)
			return
		}
		logrus.Panicf("Unable to read user: %v", err)
	}

	if user.Password != password {
		s.apiErrorCodeResponse(w, restApiV1.InvalideGrantErrorCode)
		return
	}

	b := make([]byte, 16)
	rand.Read(b)
	accessToken := fmt.Sprintf("%x", b)
	s.internalTokens.LoadOrStore(accessToken, &internalToken{userId: user.Id, creationTs: time.Now().UnixNano()})

	token := restApiV1.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		UserId:      user.Id,
	}

	tool.WriteJsonResponse(w, token)

}

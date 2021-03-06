package restSrvV1

import (
	"crypto/rand"
	"fmt"
	"github.com/jypelle/mifasol/internal/srv/storeerror"
	"github.com/jypelle/mifasol/internal/tool"
	"github.com/jypelle/mifasol/restApiV1"
	"net/http"
	"time"
)

type session struct {
	userId     restApiV1.UserId
	creationTs int64
}

func (s *RestServer) generateToken(w http.ResponseWriter, r *http.Request) {
	s.log.Debugf("Generate token")

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

	user, err := s.store.ReadUserByUserName(nil, name)
	if err != nil {
		if err == storeerror.ErrNotFound {
			s.apiErrorCodeResponse(w, restApiV1.InvalideGrantErrorCode)
			return
		}
		s.log.Panicf("Unable to read user: %v", err)
	}

	if user.Password != password {
		s.apiErrorCodeResponse(w, restApiV1.InvalideGrantErrorCode)
		return
	}

	b := make([]byte, 16)
	rand.Read(b)
	accessToken := fmt.Sprintf("%x", b)
	s.sessionMap.LoadOrStore(accessToken, &session{userId: user.Id, creationTs: time.Now().UnixNano()})

	token := restApiV1.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		UserId:      user.Id,
	}

	tool.WriteJsonResponse(w, token)

}

package restClientV1

import (
	"mifasol/restApiV1"
)

type ClientError interface {
	error
	Code() restApiV1.ErrorCode
	Description() string
}

func NewClientError(err error) ClientError {
	if err != nil {
		return &restApiV1.ApiError{ErrorCode: restApiV1.ClientErrorCode, ErrorDescription: err.Error()}
	} else {
		return nil
	}
}

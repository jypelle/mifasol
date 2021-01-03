package restApiV1

import "net/http"

type ErrorCode string

const (
	NotFoundErrorCode         ErrorCode = "not_found"
	NotImplementedErrorCode   ErrorCode = "not_implemented"
	InternalErrorCode         ErrorCode = "internal_error"
	MethodNotAllowedErrorCode ErrorCode = "method_not_allowed"
	InvalidTokenErrorCode     ErrorCode = "invalid_token"

	InvalideRequestErrorCode      ErrorCode = "invalid_request"
	UnsupportedGrantTypeErrorCode ErrorCode = "unsupported_grant_type"
	InvalideGrantErrorCode        ErrorCode = "invalid_grant"

	DeleteArtistWithSongsErrorCode  ErrorCode = "delete_artist_with_songs"
	DeleteAlbumWithSongsErrorCode   ErrorCode = "delete_album_with_songs"
	DeleteUserYourselfErrorCode     ErrorCode = "delete_user_yourself"
	CreateNotOwnedPlaylistErrorCode ErrorCode = "create_not_owned_playlist"

	ForbiddenErrorCode ErrorCode = "forbidden"

	// Client Error
	UnknownErrorCode ErrorCode = "unknown_error"
	ClientErrorCode  ErrorCode = "client_error"
)

func (e ErrorCode) StatusCode() int {
	switch e {
	case NotFoundErrorCode:
		return http.StatusNotFound
	case NotImplementedErrorCode:
		return http.StatusNotImplemented
	case InternalErrorCode:
		return http.StatusInternalServerError
	case MethodNotAllowedErrorCode:
		return http.StatusMethodNotAllowed
	case InvalidTokenErrorCode:
		return http.StatusUnauthorized
	case InvalideRequestErrorCode:
		return http.StatusBadRequest
	case UnsupportedGrantTypeErrorCode:
		return http.StatusBadRequest
	case InvalideGrantErrorCode:
		return http.StatusBadRequest
	case DeleteArtistWithSongsErrorCode:
		return http.StatusInternalServerError
	case DeleteAlbumWithSongsErrorCode:
		return http.StatusInternalServerError
	case DeleteUserYourselfErrorCode:
		return http.StatusInternalServerError
	case CreateNotOwnedPlaylistErrorCode:
		return http.StatusBadRequest
	case ForbiddenErrorCode:
		return http.StatusForbidden
	}

	return http.StatusInternalServerError
}

func (e ErrorCode) String() string {
	return string(e)
}

type ApiError struct {
	ErrorCode        ErrorCode `json:"error"`
	ErrorDescription string    `json:"error_description,omitempty"`
}

func (a *ApiError) Code() ErrorCode {
	return a.ErrorCode
}

func (a *ApiError) Description() string {
	return a.ErrorDescription
}

func (a *ApiError) Error() string {
	if a.ErrorDescription != "" {
		return string(a.ErrorCode) + ":" + a.ErrorDescription
	} else {
		return string(a.ErrorCode)
	}
}

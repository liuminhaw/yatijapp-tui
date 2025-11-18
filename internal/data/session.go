package data

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/liuminhaw/yatijapp-tui/internal/authclient"
)

type ListSessionsResponse struct {
	Metadata Metadata  `json:"metadata"`
	Sessions []Session `json:"sessions"`
	Error    string    `json:"error,omitempty"`
}

func ListSessions(
	info ListRequestInfo,
	client *authclient.AuthClient,
) (ListSessionsResponse, error) {
	var reqUrl string
	var err error
	if info.SrcUUID == "" {
		reqUrl, err = info.requestUrl("/v1/sessions")
	} else {
		reqUrl, err = info.requestUrl("/v1/actions/" + info.SrcUUID + "/sessions")
	}
	if err != nil {
		return ListSessionsResponse{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create request URL for GET Sessions",
		}
	}

	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return ListSessionsResponse{}, UnauthorizedApiDataErr{
			Err: err,
			Msg: "Failed to create GET request for Sessions",
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return ListSessionsResponse{}, respErrorCheck(err, "API request error: GET Sessions")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return ListSessionsResponse{}, UnexpectedApiDataErr{
				Err: err,
				Msg: "API error response decode failure",
			}
		}

		return ListSessionsResponse{}, UnauthorizedApiDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    responseErr.Error(),
		}
	}

	var responseData ListSessionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return ListSessionsResponse{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "API response decode failure",
		}
	}

	return responseData, nil
}

type GetSessionResponse struct {
	Session Session `json:"session"`
	Error   string  `json:"error,omitempty"`
}

func GetSession(serverURL, uuid string, client *authclient.AuthClient) (Session, error) {
	path, err := url.JoinPath(serverURL, "v1", "sessions", uuid)
	if err != nil {
		return Session{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create URL path for GET Session",
		}
	}

	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return Session{}, UnauthorizedApiDataErr{
			Err: err,
			Msg: "Failed to create GET request for Session",
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return Session{}, respErrorCheck(err, "API request error: GET Session")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return Session{}, UnexpectedApiDataErr{
				Err: err,
				Msg: "API error response decode failure",
			}
		}

		return Session{}, UnauthorizedApiDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    responseErr.Error(),
		}
	}

	var responseData GetSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return Session{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "API response decode error",
		}
	}

	responseData.Session.StartsAt = responseData.Session.StartsAt.Local()
	responseData.Session.CreatedAt = responseData.Session.CreatedAt.Local()
	responseData.Session.UpdatedAt = responseData.Session.UpdatedAt.Local()
	if responseData.Session.EndsAt.Valid {
		responseData.Session.EndsAt.Time = responseData.Session.EndsAt.Time.Local()
	}

	return responseData.Session, nil
}

func DeleteSession(serverUrl, uuid string, client *authclient.AuthClient) error {
	path, err := url.JoinPath(serverUrl, "v1", "sessions", uuid)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create URL path for DELETE Session",
		}
	}

	req, err := http.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return UnauthorizedApiDataErr{
			Err: err,
			Msg: "Failed to create DELETE request for Session",
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return respErrorCheck(err, "API request error: DELETE Session")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return UnexpectedApiDataErr{
				Err: err,
				Msg: "API error response decode failure",
			}
		}

		return UnauthorizedApiDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    responseErr.Error(),
		}
	}

	return nil
}

type SessionRequestBody struct {
	ActionUUID string       `json:"action_uuid"`
	StartsAt   time.Time    `json:"starts_at"` // in RFC3339 format
	EndsAt     sql.NullTime `json:"ends_at"`
	Notes      string       `json:"notes"`
}

func (b SessionRequestBody) Create(serverURL string, client *authclient.AuthClient) error {
	path, err := url.JoinPath(serverURL, "v1", "sessions")
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create URL path for POST Session",
		}
	}
	data, err := json.Marshal(b)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to marshal request body for POST Session",
		}
	}

	req, err := http.NewRequest(http.MethodPost, path, bytes.NewBuffer(data))
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create POST request for Session",
		}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return respErrorCheck(err, "API request error: POST Session")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return UnexpectedApiDataErr{
				Err: err,
				Msg: "API error response decode failure",
			}
		}

		return UnauthorizedApiDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    responseErr.Error(),
		}
	}

	return nil
}

func (b SessionRequestBody) Update(
	serverURL, uuid string, client *authclient.AuthClient,
) error {
	path, err := url.JoinPath(serverURL, "v1", "sessions", uuid)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create URL path for PATCH Session",
		}
	}

	data, err := json.Marshal(b)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create request body JSON",
		}
	}

	req, err := http.NewRequest(http.MethodPatch, path, bytes.NewBuffer(data))
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create PATCH request for Session",
		}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return respErrorCheck(err, "API request error: PATCH Session")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return UnexpectedApiDataErr{
				Err: err,
				Msg: "API error response decode failure",
			}
		}

		return UnauthorizedApiDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    responseErr.Error(),
		}
	}

	return nil
}

package data

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/liuminhaw/yatijapp-tui/internal/authclient"
)

type ListActivitiesResponse struct {
	Metadata   Metadata   `json:"metadata"`
	Activities []Activity `json:"activities"`
	Error      string     `json:"error,omitempty"`
}

func ListActivities(
	serverURL string,
	client *authclient.AuthClient,
	srcUUID string,
) ([]Activity, error) {
	var path string
	var err error
	if srcUUID == "" {
		path, err = url.JoinPath(serverURL, "v1", "activities")
	} else {
		path, err = url.JoinPath(serverURL, "v1", "targets", srcUUID, "activities")
	}
	if err != nil {
		return []Activity{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create URL path for GET Target Activities",
		}
	}

	req, err := http.NewRequest(
		http.MethodGet,
		path,
		nil,
	)
	if err != nil {
		return []Activity{}, UnauthorizedApiDataErr{
			Err: err,
			Msg: "Failed to create GET request for Activities",
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return []Activity{}, respErrorCheck(err, "API request error: GET Activities")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return []Activity{}, UnexpectedApiDataErr{
				Err: err,
				Msg: "API error response decode failure",
			}
		}

		return []Activity{}, UnauthorizedApiDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    responseErr.Error(),
		}
	}

	var responseData ListActivitiesResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return []Activity{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "API response decode error",
		}
	}

	return responseData.Activities, nil
}

type GetActivityResponse struct {
	Activity Activity `json:"activity"`
	Error    string   `json:"error,omitempty"`
}

func GetActivity(serverURL, uuid string, client *authclient.AuthClient) (Activity, error) {
	path, err := url.JoinPath(serverURL, "v1", "activities", uuid)
	if err != nil {
		return Activity{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create URL path for GET Activity",
		}
	}

	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return Activity{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create GET request for Activity",
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return Activity{}, respErrorCheck(err, "API request error: GET Activity")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return Activity{}, UnexpectedApiDataErr{
				Err: err,
				Msg: "API error response decode failure",
			}
		}

		return Activity{}, UnauthorizedApiDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    responseErr.Error(),
		}
	}

	var responseData GetActivityResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return Activity{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "API response decode error",
		}
	}

	responseData.Activity.CreatedAt = responseData.Activity.CreatedAt.Local()
	responseData.Activity.UpdatedAt = responseData.Activity.UpdatedAt.Local()
	responseData.Activity.LastActive = responseData.Activity.LastActive.Local()

	return responseData.Activity, nil
}

func DeleteActivity(serverURL, uuid string, client *authclient.AuthClient) error {
	path, err := url.JoinPath(serverURL, "v1", "activities", uuid)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create URL path for DELETE Activity",
		}
	}

	req, err := http.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create DELETE request for Activity",
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return respErrorCheck(err, "API request error: DELETE Activity")
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

type ActivityRequestBody struct {
	TargetUUID  string `json:"target_uuid"`
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     *Date  `json:"due_date,omitempty"`
	Notes       string `json:"notes"`
	Status      string `json:"status"`
}

func (b ActivityRequestBody) Create(serverURL string, client *authclient.AuthClient) error {
	path, err := url.JoinPath(serverURL, "v1", "activities")
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create URL path for POST Activity",
		}
	}
	data, err := json.Marshal(b)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create request body JSON",
		}
	}

	req, err := http.NewRequest(http.MethodPost, path, bytes.NewBuffer(data))
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create POST request for Activity",
		}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return respErrorCheck(err, "API request error: POST Activity")
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

func (b ActivityRequestBody) Update(
	serverURL, uuid string, client *authclient.AuthClient,
) error {
	path, err := url.JoinPath(serverURL, "v1", "activities", uuid)
	if err != nil {
		return UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create URL path for PATCH Activity",
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
			Msg: "Failed to create PATCH request for Activity",
		}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return respErrorCheck(err, "API request error: PATCH Activity")
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

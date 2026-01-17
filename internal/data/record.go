package data

import (
	"encoding/json"
	"net/http"

	"github.com/liuminhaw/yatijapp-tui/internal/authclient"
)

type ListRecordsResponse struct {
	Metadata Metadata `json:"metadata"`
	Records  []Record `json:"records"`
	Error    string   `json:"error,omitempty"`
}

func ListRecords(info ListRequestInfo, client *authclient.AuthClient) (ListRecordsResponse, error) {
	// time.Sleep(2 * time.Second) // Simulate a delay for loading records

	reqUrl, err := info.requestUrl("/v1/records")
	if err != nil {
		return ListRecordsResponse{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create request URL for GET Records",
		}
	}

	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return ListRecordsResponse{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create GET request for Records",
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return ListRecordsResponse{}, respErrorCheck(err, "API request error: GET Records")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return ListRecordsResponse{}, UnexpectedApiDataErr{
				Err: err,
				Msg: "API error response decode failure",
			}
		}

		return ListRecordsResponse{}, UnauthorizedApiDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    responseErr.Error(),
		}
	}

	var responseData ListRecordsResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return ListRecordsResponse{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "API response decode error",
		}
	}

	return responseData, nil
}

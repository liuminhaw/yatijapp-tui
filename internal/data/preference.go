package data

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/liuminhaw/yatijapp-tui/internal/authclient"
)

type Filter struct {
	SortBy    string   `json:"sortBy"`
	SortOrder string   `json:"sortOrder"`
	Status    []string `json:"status"`
}

// SortKey maps user-friendly sortBy values to API field names.
func (f Filter) SortKey() string {
	switch f.SortBy {
	case "id":
		return "serial_id"
	case "due date":
		return "due_date"
	case "created at":
		return "created_at"
	case "starts at":
		return "starts_at"
	case "updated at":
		return "updated_at"
	case "last active":
		return "last_active"
	default:
		return f.SortBy
	}
}

// SortOption maps API field names to user-friendly sortBy values.
func (f Filter) SortOption() string {
	switch f.SortBy {
	case "serial_id":
		return "id"
	case "due_date":
		return "due date"
	case "created_at":
		return "created at"
	case "starts_at":
		return "starts at"
	case "updated_at":
		return "updated at"
	case "last_active":
		return "last active"
	default:
		return f.SortBy
	}
}

type Filters struct {
	Target  Filter `json:"target"`
	Action  Filter `json:"action"`
	Session Filter `json:"session"`
}

type Preferences struct {
	Filters Filters `json:"filters"`
	Version string  `json:"version"`
}

var DefaultPreferences Preferences = Preferences{
	Filters: Filters{
		Target: Filter{
			SortBy:    "last_active",
			SortOrder: "descending",
			Status:    []string{"queued", "in progress"},
		},
		Action: Filter{
			SortBy:    "last_active",
			SortOrder: "descending",
			Status:    []string{"queued", "in progress"},
		},
		Session: Filter{
			SortBy:    "starts_at",
			SortOrder: "descending",
			Status:    []string{"in progress", "completed"},
		},
	},
	Version: "2025-12-09",
}

type RecordFilter struct {
	RecordType RecordType
	Filter     Filter
}

func (p Preferences) GetFilter(recordType RecordType) RecordFilter {
	switch recordType {
	case RecordTypeTarget:
		return RecordFilter{
			RecordType: RecordTypeTarget,
			Filter:     p.Filters.Target,
		}
	case RecordTypeAction:
		return RecordFilter{
			RecordType: RecordTypeAction,
			Filter:     p.Filters.Action,
		}
	case RecordTypeSession:
		return RecordFilter{
			RecordType: RecordTypeSession,
			Filter:     p.Filters.Session,
		}
	default:
		panic("unsupported record type in GetFilter")
	}
}

func (p *Preferences) formatApiResponse() {
	asc := "ascending"
	desc := "descending"
	if strings.HasPrefix(p.Filters.Target.SortBy, "-") {
		p.Filters.Target.SortBy = p.Filters.Target.SortBy[1:]
		p.Filters.Target.SortOrder = desc
	} else {
		p.Filters.Target.SortOrder = asc
	}
	if strings.HasPrefix(p.Filters.Action.SortBy, "-") {
		p.Filters.Action.SortBy = p.Filters.Action.SortBy[1:]
		p.Filters.Action.SortOrder = desc
	} else {
		p.Filters.Action.SortOrder = asc
	}
	if strings.HasPrefix(p.Filters.Session.SortBy, "-") {
		p.Filters.Session.SortBy = p.Filters.Session.SortBy[1:]
		p.Filters.Session.SortOrder = desc
	} else {
		p.Filters.Session.SortOrder = asc
	}
}

type PreferencesResponse struct {
	Preferences Preferences `json:"preferences"`
}

func GetPreferences(serverURL string, client *authclient.AuthClient) (Preferences, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/v1/users/preferences", serverURL),
		nil,
	)
	if err != nil {
		return Preferences{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "Failed to create GET request for Preferences",
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return Preferences{}, respErrorCheck(err, "API request error: GET Preferences")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return Preferences{}, UnexpectedApiDataErr{
				Err: err,
				Msg: "API error response decode failure",
			}
		}

		if resp.StatusCode == http.StatusNotFound {
			return Preferences{}, NotFoundApiDataErr{
				Err: responseErr,
				Msg: responseErr.Error(),
			}
		}

		return Preferences{}, UnauthorizedApiDataErr{
			Status: resp.StatusCode,
			Err:    responseErr,
			Msg:    responseErr.Error(),
		}
	}

	var responseData PreferencesResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return Preferences{}, UnexpectedApiDataErr{
			Err: err,
			Msg: "API response decode error",
		}
	}

	// Clean data
	responseData.Preferences.formatApiResponse()

	return responseData.Preferences, nil
}

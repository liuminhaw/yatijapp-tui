package data

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var FailedToLoadDataErr = errors.New("Failed to load data")

type ListTargetsResponse struct {
	Metadata Metadata `json:"metadata"`
	Targets  []Target `json:"targets"`
	Error    string   `json:"error,omitempty"`
}

func ListTargets(serverURL string) ([]Target, error) {
	// time.Sleep(3 * time.Second) // Simulate a delay for loading targets
	// resp, err := http.Get("http://localhost:8080/v1/targets")
	resp, err := http.Get(fmt.Sprintf("%s/v1/targets", serverURL))
	if err != nil {
		return []Target{}, errors.New("Failed to load data: GET Targets")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []Target{}, errors.New("Failed to load data: " + resp.Status)
	}

	var responseData ListTargetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return []Target{}, errors.New("Failed to load data: decode response")
	}

	return responseData.Targets, nil
}

type GetTargetResponse struct {
	Target Target `json:"target"`
	Error  string `json:"error,omitempty"`
}

func GetTarget(serverURL, uuid string) (Target, error) {
	// time.Sleep(3 * time.Second) // Simulate a delay for loading targets
	resp, err := http.Get(fmt.Sprintf("%s/v1/targets/%s", serverURL, uuid))
	if err != nil {
		return Target{}, errors.New("Failed to load data: GET Target")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Target{}, fmt.Errorf("Failed to load data: %s", resp.Status)
	}

	var responseData GetTargetResponse
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return Target{}, errors.New("Failed to load data: decode response")
	}

	responseData.Target.CreatedAt = responseData.Target.CreatedAt.Local()
	responseData.Target.UpdatedAt = responseData.Target.UpdatedAt.Local()

	return responseData.Target, nil
}

func DeleteTarget(serverURL, uuid string) error {
	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/v1/targets/%s", serverURL, uuid),
		nil,
	)
	if err != nil {
		return errors.New("Failed to create DELETE request for target")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("Failed to delete: DELETE target")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return errors.New("Request failed: decoding error response")
		}

		return fmt.Errorf("Failed to delete target: %s", responseErr.Error)
	}

	return nil
}

type TargetRequestBody struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	// DueDate     *time.Time `json:"due_date,omitempty"`
	DueDate *Date  `json:"due_date,omitempty"`
	Notes   string `json:"notes"`
	Status  string `json:"status"`
}

func (b TargetRequestBody) Create(serverURL string) error {
	data, err := json.Marshal(b)
	if err != nil {
		return errors.New("Failed to marshal request body JSON")
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/v1/targets", serverURL),
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		// return fmt.Errorf("Failed to create target: %v", err)
		return errors.New("Failed to create: POST target")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			// return err
			return errors.New("Request failed: decoding error response")
		}

		// return fmt.Errorf("Failed to create target: %s", resp.Status)
		return fmt.Errorf("Failed to create target: %s", responseErr.Error)
	}

	return nil
}

func (b TargetRequestBody) Update(serverURL string, uuid string) error {
	data, err := json.Marshal(b)
	if err != nil {
		return errors.New("Failed to marshal request body JSON")
	}

	req, err := http.NewRequest(
		http.MethodPatch,
		fmt.Sprintf("%s/v1/targets/%s", serverURL, uuid),
		bytes.NewBuffer(data),
	)
	if err != nil {
		return errors.New("Failed to create PATCH request for target update")
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("Failed to update: PATCH target")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var responseErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
			return errors.New("Request failed: decoding error response")
		}

		return fmt.Errorf("Failed to update target: %s", responseErr.Error)
	}

	return nil
}

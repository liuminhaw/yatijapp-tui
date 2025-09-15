package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/liuminhaw/yatijapp-tui/internal/authclient"
	"github.com/liuminhaw/yatijapp-tui/internal/data"
)

type yatijappRecord interface {
	GetActualType() data.RecordType

	GetUUID() string
	GetTitle() string
	GetDescription() string
	GetStatus() string
	GetNote() string
	GetDueDate() (time.Time, bool)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetLastActive() time.Time
	GetParentsUUID() map[data.RecordType]string
	GetParentsTitle() map[data.RecordType]string

	ListItemView(hasSrc, chosen bool, width int) string
}

type (
	allRecordsLoadedMsg struct {
		records []yatijappRecord
		msg     string
	}
	getRecordLoadedMsg struct {
		record yatijappRecord
		msg    string
	}
	recordDeletedMsg string

	targetListLoadedMsg struct {
		targets []data.Target
		msg     string
	}
	getTargetLoadedMsg struct {
		target data.Target
		msg    string
	}
	targetDeletedMsg string

	activityListLoadedMsg struct {
		activities []data.Activity
		msg        string
	}
	getActivityLoadedMsg struct {
		activity data.Activity
		msg      string
	}
	activityDeletedMsg string

	apiSuccessResponseMsg struct {
		msg      string
		source   tea.Model
		redirect tea.Model
	}
)

func loadAllTargets(serverURL, nilUUID, msg string, client *authclient.AuthClient) tea.Cmd {
	return func() tea.Msg {
		targets, err := data.ListTargets(serverURL, client)
		if err != nil {
			return err
		}

		records := make([]yatijappRecord, len(targets))
		for i, target := range targets {
			records[i] = target
		}

		return allRecordsLoadedMsg{
			records: records,
			msg:     msg,
		}
	}
}

func loadAllActivities(serverURL, srcUUID, msg string, client *authclient.AuthClient) tea.Cmd {
	return func() tea.Msg {
		activities, err := data.ListActivities(serverURL, client, srcUUID)
		if err != nil {
			return err
		}

		records := make([]yatijappRecord, len(activities))
		for i, activity := range activities {
			records[i] = activity
		}

		return allRecordsLoadedMsg{
			records: records,
			msg:     msg,
		}
	}
}

func loadTarget(serverURL, uuid, msg string, client *authclient.AuthClient) tea.Cmd {
	return func() tea.Msg {
		target, err := data.GetTarget(serverURL, uuid, client)
		if err != nil {
			return err
		}

		return getRecordLoadedMsg{
			record: target,
			msg:    msg,
		}
	}
}

func loadActivity(serverURL, uuid, msg string, client *authclient.AuthClient) tea.Cmd {
	return func() tea.Msg {
		activity, err := data.GetActivity(serverURL, uuid, client)
		if err != nil {
			return err
		}

		return getRecordLoadedMsg{
			record: activity,
			msg:    msg,
		}
	}
}

func deleteTarget(serverURL, uuid string, client *authclient.AuthClient) tea.Cmd {
	return func() tea.Msg {
		if err := data.DeleteTarget(serverURL, uuid, client); err != nil {
			return err
		}

		return recordDeletedMsg("Target deleted successfully.")
	}
}

func deleteActivity(serverURL, uuid string, client *authclient.AuthClient) tea.Cmd {
	return func() tea.Msg {
		if err := data.DeleteActivity(serverURL, uuid, client); err != nil {
			return err
		}

		return recordDeletedMsg("Activity deleted successfully.")
	}
}

type recordRequestData struct {
	// Common fields
	uuid        string
	title       string
	description string
	status      string
	note        string
	dueDate     string
	// Activity specific
	targetUUID string
}

func (d recordRequestData) targetRequestBody() data.TargetRequestBody {
	body := data.TargetRequestBody{
		Title:       d.title,
		Description: d.description,
		Status:      d.status,
		Notes:       d.note,
	}
	if d.dueDate != "" {
		dueDate, err := time.ParseInLocation("2006-01-02", d.dueDate, time.Local)
		if err == nil {
			dueDateTyped := data.Date(dueDate)
			body.DueDate = &dueDateTyped
		}
	}
	return body
}

func (d recordRequestData) activityRequestBody() data.ActivityRequestBody {
	body := data.ActivityRequestBody{
		TargetUUID:  d.targetUUID,
		Title:       d.title,
		Description: d.description,
		Status:      d.status,
		Notes:       d.note,
	}
	if d.dueDate != "" {
		dueDate, err := time.ParseInLocation("2006-01-02", d.dueDate, time.Local)
		if err == nil {
			dueDateTyped := data.Date(dueDate)
			body.DueDate = &dueDateTyped
		}
	}
	return body
}

func createTarget(
	serverURL string,
	d recordRequestData,
	src, redirect tea.Model,
	client *authclient.AuthClient,
) tea.Cmd {
	return func() tea.Msg {
		request := d.targetRequestBody()
		if err := request.Create(serverURL, client); err != nil {
			return err
		}

		return apiSuccessResponseMsg{
			msg:      "Target created successfully",
			source:   src,
			redirect: redirect,
		}
	}
}

func updateTarget(
	serverURL string,
	d recordRequestData,
	src, redirect tea.Model,
	client *authclient.AuthClient,
) tea.Cmd {
	return func() tea.Msg {
		request := d.targetRequestBody()
		if err := request.Update(serverURL, d.uuid, client); err != nil {
			return err
		}

		return apiSuccessResponseMsg{
			msg:      "Target updated successfully",
			source:   src,
			redirect: redirect,
		}
	}
}

func createActivity(
	serverURL string,
	d recordRequestData,
	src, redirect tea.Model,
	client *authclient.AuthClient,
) tea.Cmd {
	return func() tea.Msg {
		request := d.activityRequestBody()
		if err := request.Create(serverURL, client); err != nil {
			return err
		}

		return apiSuccessResponseMsg{
			msg:      "Activity created successfully",
			source:   src,
			redirect: redirect,
		}
	}
}

func updateActivity(
	serverURL string,
	d recordRequestData,
	src, redirect tea.Model,
	client *authclient.AuthClient,
) tea.Cmd {
	return func() tea.Msg {
		request := d.activityRequestBody()
		if err := request.Update(serverURL, d.uuid, client); err != nil {
			return err
		}

		return apiSuccessResponseMsg{
			msg:      "Activity updated successfully",
			source:   src,
			redirect: redirect,
		}
	}
}

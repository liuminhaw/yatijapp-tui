package main

import (
	"database/sql"
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
	GetChildrenCount() int64

	HasNote() bool

	ListItemView(hasSrc, chosen bool, width int) string
	ListItemDetailView(hasSrc bool, width int) string
}

type (
	allRecordsLoadedMsg struct {
		metadata data.Metadata
		records  []yatijappRecord
		src      string
		msg      string
		events   []string
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

	actionListLoadedMsg struct {
		actions []data.Action
		msg     string
	}
	getActionLoadedMsg struct {
		action data.Action
		msg    string
	}
	actionDeletedMsg string

	apiSuccessResponseMsg struct {
		msg      string
		source   tea.Model
		redirect tea.Model
	}

	loadMoreRecordsMsg struct {
		direction string
	}
)

func loadMoreRecords(direction string) tea.Cmd {
	return func() tea.Msg {
		return loadMoreRecordsMsg{direction: direction}
	}
}

func loadAllTargets(
	info data.ListRequestInfo,
	msg, src string,
	client *authclient.AuthClient,
) tea.Cmd {
	return func() tea.Msg {
		resp, err := data.ListTargets(info, client)
		if err != nil {
			return err
		}

		records := make([]yatijappRecord, len(resp.Targets))
		for i, target := range resp.Targets {
			records[i] = target
		}

		return allRecordsLoadedMsg{
			metadata: resp.Metadata,
			records:  records,
			src:      src,
			msg:      msg,
			events:   info.Events,
		}
	}
}

func loadAllActions(
	info data.ListRequestInfo,
	msg, src string,
	client *authclient.AuthClient,
) tea.Cmd {
	return func() tea.Msg {
		resp, err := data.ListActions(info, client)
		if err != nil {
			return err
		}

		records := make([]yatijappRecord, len(resp.Actions))
		for i, action := range resp.Actions {
			records[i] = action
		}

		return allRecordsLoadedMsg{
			metadata: resp.Metadata,
			records:  records,
			src:      src,
			msg:      msg,
			events:   info.Events,
		}
	}
}

func loadAllSessions(
	info data.ListRequestInfo,
	msg, src string,
	client *authclient.AuthClient,
) tea.Cmd {
	return func() tea.Msg {
		resp, err := data.ListSessions(info, client)
		if err != nil {
			return err
		}

		records := make([]yatijappRecord, len(resp.Sessions))
		for i, session := range resp.Sessions {
			records[i] = session
		}

		return allRecordsLoadedMsg{
			metadata: resp.Metadata,
			records:  records,
			src:      src,
			msg:      msg,
			events:   info.Events,
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

func loadAction(serverURL, uuid, msg string, client *authclient.AuthClient) tea.Cmd {
	return func() tea.Msg {
		action, err := data.GetAction(serverURL, uuid, client)
		if err != nil {
			return err
		}

		return getRecordLoadedMsg{
			record: action,
			msg:    msg,
		}
	}
}

func loadSession(serverURL, uuid, msg string, client *authclient.AuthClient) tea.Cmd {
	return func() tea.Msg {
		session, err := data.GetSession(serverURL, uuid, client)
		if err != nil {
			return err
		}

		return getRecordLoadedMsg{
			record: session,
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

func deleteAction(serverURL, uuid string, client *authclient.AuthClient) tea.Cmd {
	return func() tea.Msg {
		if err := data.DeleteAction(serverURL, uuid, client); err != nil {
			return err
		}

		return recordDeletedMsg("Action deleted successfully.")
	}
}

func deleteSession(serverURL, uuid string, client *authclient.AuthClient) tea.Cmd {
	return func() tea.Msg {
		if err := data.DeleteSession(serverURL, uuid, client); err != nil {
			return err
		}

		return recordDeletedMsg("Session deleted successfully.")
	}
}

type nullNote struct {
	valid bool
	note  string
}

type recordRequestData struct {
	// Common fields
	uuid        string
	title       string
	description string
	status      string
	note        nullNote
	dueDate     string
	// Action specific
	targetUUID string
	// Session specific
	startsAt   time.Time
	endsAt     sql.NullTime
	actionUUID string
}

func (d recordRequestData) targetRequestBody() data.TargetRequestBody {
	body := data.TargetRequestBody{
		Title:       d.title,
		Description: d.description,
		Status:      d.status,
	}
	if d.note.valid {
		body.Notes = d.note.note
	}
	if d.dueDate != "" {
		dueDate, err := time.ParseInLocation("2006-01-02", d.dueDate, time.Local)
		if err == nil {
			dueDateTyped := data.Date(dueDate)
			body.DueDate = dueDateTyped.String()
		}
	}
	return body
}

func (d recordRequestData) actionRequestBody() data.ActionRequestBody {
	body := data.ActionRequestBody{
		TargetUUID:  d.targetUUID,
		Title:       d.title,
		Description: d.description,
		Status:      d.status,
	}
	if d.note.valid {
		body.Notes = d.note.note
	}
	if d.dueDate != "" {
		dueDate, err := time.ParseInLocation("2006-01-02", d.dueDate, time.Local)
		if err == nil {
			dueDateTyped := data.Date(dueDate)
			body.DueDate = dueDateTyped.String()
		}
	}
	return body
}

func (d recordRequestData) sessionRequestBody() data.SessionRequestBody {
	body := data.SessionRequestBody{
		ActionUUID: d.actionUUID,
		EndsAt:     d.endsAt,
	}
	if d.note.valid {
		body.Notes = &d.note.note
	}
	if !d.startsAt.IsZero() {
		body.StartsAt = &d.startsAt
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
	serverURL, msg string,
	d recordRequestData,
	src, redirect tea.Model,
	client *authclient.AuthClient,
) tea.Cmd {
	responseMsg := "Target updated successfully"
	if msg != "" {
		responseMsg = msg
	}

	return func() tea.Msg {
		request := d.targetRequestBody()
		if err := request.Update(serverURL, d.uuid, client); err != nil {
			return err
		}

		return apiSuccessResponseMsg{
			msg:      responseMsg,
			source:   src,
			redirect: redirect,
		}
	}
}

func createAction(
	serverURL string,
	d recordRequestData,
	src, redirect tea.Model,
	client *authclient.AuthClient,
) tea.Cmd {
	return func() tea.Msg {
		request := d.actionRequestBody()
		if err := request.Create(serverURL, client); err != nil {
			return err
		}

		return apiSuccessResponseMsg{
			msg:      "Action created successfully",
			source:   src,
			redirect: redirect,
		}
	}
}

func updateAction(
	serverURL, msg string,
	d recordRequestData,
	src, redirect tea.Model,
	client *authclient.AuthClient,
) tea.Cmd {
	responseMsg := "Action updated successfully"
	if msg != "" {
		responseMsg = msg
	}

	return func() tea.Msg {
		request := d.actionRequestBody()
		if err := request.Update(serverURL, d.uuid, client); err != nil {
			return err
		}

		return apiSuccessResponseMsg{
			msg:      responseMsg,
			source:   src,
			redirect: redirect,
		}
	}
}

func createSession(
	serverURL string,
	d recordRequestData,
	src, redirect tea.Model,
	client *authclient.AuthClient,
) tea.Cmd {
	return func() tea.Msg {
		request := d.sessionRequestBody()
		if err := request.Create(serverURL, client); err != nil {
			return err
		}

		return apiSuccessResponseMsg{
			msg:      "New session started",
			source:   src,
			redirect: redirect,
		}
	}
}

func updateSession(
	serverURL, msg string,
	d recordRequestData,
	src, redirect tea.Model,
	client *authclient.AuthClient,
) tea.Cmd {
	responseMsg := "Session updated successfully"
	if msg != "" {
		responseMsg = msg
	}

	return func() tea.Msg {
		request := d.sessionRequestBody()
		if err := request.Update(serverURL, d.uuid, client); err != nil {
			return err
		}

		return apiSuccessResponseMsg{
			msg:      responseMsg,
			source:   src,
			redirect: redirect,
		}
	}
}

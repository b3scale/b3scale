package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/http/auth"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/urfave/cli/v2"
)

func (c *Cli) showMeetings(ctx *cli.Context) error {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Meeting ID", "Meeting Name", "Attendees", "Frontend", "Backend"})

	client, err := apiClient(ctx)
	if err != nil {
		return err
	}
	//	beKey := ctx.String("backend")

	meetings, err := client.MeetingsList(ctx.Context, nil)
	if err != nil {
		return err
	}

	meetingCount := 0
	for i, m := range meetings {
		meetingCount = i + 1
		backendName := "<unknown>"
		if m.BackendID != nil {
			frontend, err := client.BackendRetrieve(ctx.Context, *m.BackendID)
			if err != nil {
				return err
			}
			backendName = frontend.Backend.Host
		}
		frontendName := "<unknown>"
		if m.FrontendID != nil {
			frontend, err := client.FrontendRetrieve(ctx.Context, *m.FrontendID)
			if err != nil {
				frontendName = frontend.Frontend.Key
			}
		}
		t.AppendRow(table.Row{m.Meeting.MeetingID, m.Meeting.MeetingName, len(m.Meeting.Attendees), frontendName, backendName})
	}
	t.AppendFooter(table.Row{"", "", "Total", meetingCount})
	t.Render()
	return nil
}

func (c *Cli) createMeeting(ctx *cli.Context) error {
	name := ctx.String("name")
	if name == "" {
		// We reuse the ref generator
		name = auth.GenerateRef(2)
	}
	feKey := ctx.String("frontend")
	if feKey == "" {
		return fmt.Errorf("a frontend is required")
	}

	client, err := apiClient(ctx)
	if err != nil {
		return err
	}

	// New MeetingID
	meetingID := auth.GenerateNonce(23)

	// Convert bbb params
	params := ctx.StringSlice("param")
	bbbParams := bbb.Params{
		bbb.ParamMeetingID: meetingID,
		bbb.ParamName:      name,
	}

	for _, param := range params {
		tokens := strings.SplitN(param, "=", 2)
		if len(tokens) != 2 {
			return fmt.Errorf("invalid param: %s", param)
		}
		bbbParams[tokens[0]] = tokens[1]
	}

	// Get frontend
	state, err := getFrontendByKey(ctx.Context, client, feKey)
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("frontend not found")
	}

	// Make BBB request
	apiHost := ctx.String("api")
	backend := &bbb.Backend{
		Host:   apiHost + "/bbb/" + feKey,
		Secret: state.Frontend.Secret,
	}
	request := bbb.CreateRequest(bbbParams, nil).WithBackend(backend)

	fmt.Println("Creating Meeting:", name, "(", meetingID, ")")

	bc := bbb.NewClient()
	res, err := bc.Do(ctx.Context, request)
	if err != nil {
		return err
	}

	createRes := res.(*bbb.CreateResponse)
	if createRes.Meeting == nil {
		return fmt.Errorf("meeting was not created on server")
	}

	if !res.IsSuccess() {
		return fmt.Errorf("meeting was not created on server")
	}

	fmt.Println("Meeting created:", createRes)

	attendeeName := auth.GenerateRef(2)
	fmt.Println("Joining Meeting as:", attendeeName)
	joinReq := bbb.JoinRequest(bbb.Params{
		bbb.ParamMeetingID: meetingID,
		"fullName":         attendeeName,
		"role":             "MODERATOR",
	}).WithBackend(backend)
	res, err = bc.Do(ctx.Context, joinReq)
	if err != nil {
		return err
	}
	fmt.Println("Joined Meeting:", meetingID)

	joinRes := res.(*bbb.JoinResponse)
	headers := joinRes.Header()

	fmt.Println("URL:", headers.Get("Location"))

	return nil
}

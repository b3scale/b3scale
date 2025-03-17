package main

import (
	"encoding/json"
	"fmt"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/store"
	"github.com/urfave/cli/v2"
)

// showRecordings returns the recording
func (c *Cli) showRecordings(ctx *cli.Context) error {
	var (
		recs []*store.RecordingState
		err  error
	)

	feID := ctx.String("frontend-id")
	feKey := ctx.String("frontend")

	if feID == "" && feKey == "" {
		return fmt.Errorf("--frontend-id or --frontend is required")
	}

	client, err := apiClient(ctx)
	if err != nil {
		return err
	}

	if feID != "" {
		recs, err = client.RecordingsListByFrontendID(ctx.Context, feID)
		if err != nil {
			return err
		}
	} else {
		recs, err = client.RecordingsListByFrontendKey(ctx.Context, feKey)
		if err != nil {
			return err
		}
	}

	if ctx.Bool("json") {
		buf, _ := json.MarshalIndent(recs, "", "   ")
		fmt.Println(string(buf))
		return nil
	}

	fmt.Println("Recording\tTimestamp\tMeeting")
	for _, rec := range recs {
		meta := rec.Recording.Metadata
		fmt.Println(rec.RecordID, rec.Recording.EndTime, meta["meetingName"])
	}

	return nil
}

// showRecording returns the recording
func (c *Cli) showRecording(ctx *cli.Context) error {
	id := ctx.Args().Get(0)
	if id == "" {
		return fmt.Errorf("a recording id is required")
	}

	client, err := apiClient(ctx)
	if err != nil {
		return err
	}

	rec, err := client.RecordingsRetrieve(ctx.Context, id)
	if err != nil {
		return err
	}

	buf, _ := json.MarshalIndent(rec, "", "   ")
	fmt.Println(string(buf))

	return nil
}

// setRecordingVisibility changes a recording visbility
func (c *Cli) setRecordingVisibility(ctx *cli.Context) error {
	recID := ctx.Args().Get(0)
	vVal := ctx.Args().Get(1)

	if recID == "" {
		return fmt.Errorf("a recording id is required")
	}
	if vVal == "" {
		return fmt.Errorf("a visibility is required")
	}

	v, err := bbb.ParseRecordingVisibility(vVal)
	if err != nil {
		return err
	}

	fmt.Println("Update recording:", recID, "Visiblity:", v)

	client, err := apiClient(ctx)
	if err != nil {
		return err
	}

	if _, err := client.RecordingsSetVisibility(ctx.Context, recID, v); err != nil {
		return err
	}

	fmt.Println("Ok.")

	return nil
}

package requests

import (
	"strings"
	"testing"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/store"
)

// Test Filters

func TestMaybeFilterRecordingMeetingIDs(t *testing.T) {
	params := bbb.Params{
		bbb.ParamMeetingID: "id1,id2,id3",
	}

	qry := store.QueryRecordingsByFrontendKey("fk").
		Columns("recordings.state").
		From("recordings")

	qry = maybeFilterRecordingMeetingIDs(qry, params)

	sql, args, err := qry.ToSql()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(sql, "meeting_id =") {
		t.Error("unexpected sql:", sql)
	}

	if args[1].(string) != "id1" {
		t.Error("unexpected arg:", args[1])
	}
}

func TestMaybeFilterRecordingIDs(t *testing.T) {
	params := bbb.Params{
		bbb.ParamRecordID: "id1,id2,id3",
	}

	qry := store.QueryRecordingsByFrontendKey("fk").
		Columns("recordings.state").
		From("recordings")

	qry = maybeFilterRecordingIDs(qry, params)

	sql, args, err := qry.ToSql()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(sql, "record_id LIKE") {
		t.Error("unexpected sql:", sql)
	}

	if args[1].(string) != "id1%" {
		t.Error("unexpected arg:", args[1])
	}
}

func TestMaybeFilterRecordingMeta(t *testing.T) {
	params := bbb.Params{
		"meta_gl-listed":     "true",
		"meta_meetingId":     "foo",
		"meta_'; DROP TABLE": "students",
	}

	qry := store.QueryRecordingsByFrontendKey("fk").
		Columns("recordings.state").
		From("recordings")

	qry = maybeFilterRecordingMeta(qry, params)

	sql, args, err := qry.ToSql()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(sql, "gl-listed") {
		t.Error("sql should contain gl-listed:", sql)
	}
	if strings.Contains(sql, "DROP TABLE") {
		t.Error("harmful SQL should be filtered:", sql)
	}

	if len(args) != 4 {
		t.Error("expected 4 args:", args)
	}
}

func TestMaybeFilterRecordingStates(t *testing.T) {
	params := bbb.Params{
		bbb.ParamState: "published",
	}

	qry := store.QueryRecordingsByFrontendKey("fk").
		Columns("recordings.state").
		From("recordings")

	qry = maybeFilterRecordingStates(qry, params)

	sql, args, err := qry.ToSql()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(sql, "recordings.state -> 'State'") {
		t.Error("unexpected sql:", sql)
	}

	if args[1].(string) != "\"published\"" {
		t.Error("unexpected arg:", args[1])
	}
}

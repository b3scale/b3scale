package requests

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/cluster"
)

const (
	// CheckKnownValue is a well known value
	CheckKnownValue = "b3scl"
)

// The FrontendKeyMeetingID is the combination of a frontend
// key and a meeting
type FrontendKeyMeetingID struct {
	FrontendKey string
	MeetingID   string
}

// EncodeToString encodes the combined ID as base64 string
func (id *FrontendKeyMeetingID) EncodeToString() string {
	rep := []string{id.FrontendKey, id.MeetingID, CheckKnownValue}
	data, _ := json.Marshal(rep)
	return base64.URLEncoding.EncodeToString(data)
}

// DecodeFrontendKeyMeetingID decodes a base64 encoded gob
// representation of the meeting id
func DecodeFrontendKeyMeetingID(id string) *FrontendKeyMeetingID {
	// Base64 and json decode
	data, err := base64.URLEncoding.DecodeString(id)
	if err != nil {
		return nil
	}
	rep := []string{}
	if err := json.Unmarshal(data, &rep); err != nil {
		return nil
	}

	// Check data
	if len(rep) != 3 {
		return nil
	}
	if rep[2] != CheckKnownValue {
		return nil
	}

	// Rebuild struct
	return &FrontendKeyMeetingID{
		FrontendKey: rep[0],
		MeetingID:   rep[1],
	}
}

// Decode the meetingID if it is encoded, otherwise
// just be transparent
func maybeDecodeMeetingID(id string) string {
	fkmid := DecodeFrontendKeyMeetingID(id)
	if fkmid == nil {
		return id
	}
	return fkmid.MeetingID
}

// Apply the meetingID rewrite to meetingID fields
// of a meeting and breakout.
func maybeRewriteMeeting(m *bbb.Meeting) *bbb.Meeting {
	if m == nil {
		return nil
	}
	m.MeetingID = maybeDecodeMeetingID(m.MeetingID)
	if m.Breakout != nil {
		m.Breakout.ParentMeetingID = maybeDecodeMeetingID(
			m.Breakout.ParentMeetingID)
	}
	return m
}

func maybeRewriteMeetingsCollection(c []*bbb.Meeting) []*bbb.Meeting {
	for _, m := range c {
		m = maybeRewriteMeeting(m)
	}
	return c
}

func maybeRewriteRecording(r *bbb.Recording) *bbb.Recording {
	r.MeetingID = maybeDecodeMeetingID(r.MeetingID)

	// Also update recording metadata
	if r.Metadata != nil {
		r.Metadata["meetingId"] = r.MeetingID
	}

	return r
}

func maybeRewriteRecordingsCollection(c []*bbb.Recording) []*bbb.Recording {
	for _, r := range c {
		r = maybeRewriteRecording(r)
	}
	return c
}

// RewriteUniqueMeetingID ensures that the meeting id is unique
// by combining FrontendKey and MeetingID.
//
// The resonse may contain MeetingIDs. If this is the case,
// the meeting ID will be decoded and the original meeting
// id will be restored.
func RewriteUniqueMeetingID() cluster.RequestMiddleware {
	return func(next cluster.RequestHandler) cluster.RequestHandler {
		return rewriteUniqueMeetingIDHandler(next)
	}
}

func rewriteUniqueMeetingIDHandler(next cluster.RequestHandler) cluster.RequestHandler {
	return func(ctx context.Context, req *bbb.Request) (bbb.Response, error) {
		req = rewriteUniqueMeetingIDRequest(req)
		res, err := next(ctx, req)
		if err != nil {
			return nil, err
		}
		return rewriteUniqueMeetingIDResponse(res)
	}
}

// Rewrite the request
// Warning: this mutates the request. We'll change this
// if it actually becomes a problem.
func rewriteUniqueMeetingIDRequest(req *bbb.Request) *bbb.Request {
	meetingID, _ := req.Params.MeetingID()
	meetingIDs, ok := req.Params.MeetingIDs()
	if !ok {
		return req // nothing to do here.
	}
	frontendKey := req.Frontend.Key

	fkmids := make([]string, 0, len(meetingIDs))
	for _, id := range meetingIDs {
		// Encode key and secret
		fkmid := (&FrontendKeyMeetingID{
			FrontendKey: frontendKey,
			MeetingID:   id}).EncodeToString()
		fkmids = append(fkmids, fkmid)
	}

	fkmid := strings.Join(fkmids, ",")

	log.Debug().
		Str("frontendKey", frontendKey).
		Str("orgMeetingID", meetingID).
		Str("newMeetingID", fkmid).
		Msg("rewrote meetingID")

	// Update request params
	req.Params[bbb.ParamMeetingID] = fkmid
	return req
}

// Rewrite the response
func rewriteUniqueMeetingIDResponse(res bbb.Response) (bbb.Response, error) {
	// We need to treat each reponse a bit differently
	switch r := res.(type) {
	case *bbb.JoinResponse:
		r.MeetingID = maybeDecodeMeetingID(r.MeetingID)
	case *bbb.CreateResponse:
		r.Meeting = maybeRewriteMeeting(r.Meeting)
	case *bbb.GetMeetingInfoResponse:
		r.Meeting = maybeRewriteMeeting(r.Meeting)
	case *bbb.GetMeetingsResponse:
		r.Meetings = maybeRewriteMeetingsCollection(r.Meetings)
	case *bbb.GetRecordingsResponse:
		r.Recordings = maybeRewriteRecordingsCollection(r.Recordings)
	}

	return res, nil
}

package requests

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
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
		req, err := rewriteUniqueMeetingIDRequest(req)
		if err != nil {
			return nil, err
		}
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
func rewriteUniqueMeetingIDRequest(req *bbb.Request) (*bbb.Request, error) {
	meetingID, ok := req.Params.MeetingID()
	if !ok {
		return req, nil // nothing to do here.
	}

	frontendKey := req.Frontend.Key

	// Encode key and secret
	fkmid := (&FrontendKeyMeetingID{
		FrontendKey: frontendKey,
		MeetingID:   meetingID}).EncodeToString()

	// Update request params
	req.Params[bbb.ParamMeetingID] = fkmid

	return req, nil
}

// Rewrite the response
func rewriteUniqueMeetingIDResponse(res bbb.Response) (bbb.Response, error) {
	return res, nil
}

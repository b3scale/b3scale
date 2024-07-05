package api

import "testing"

func TestParseRecordIDPath(t *testing.T) {
	path := "/playback/presentation/2.3/9b897750e3453b1daa4563788af47ef90e063aa3-1716030289891"
	rID, ok := parseRecordIDPath(path)
	if !ok {
		t.Error("expected record id match")
	}
	if rID != "9b897750e3453b1daa4563788af47ef90e063aa3-1716030289891" {
		t.Error("unexpected recordID:", rID)
	}

	path = "/presentation/9b897750e3453b1daa4563788af47ef90e063aa3-1716030289891/video/webcams.webm"
	rID, ok = parseRecordIDPath(path)
	if !ok {
		t.Error("expected record id match")
	}
	if rID != "9b897750e3453b1daa4563788af47ef90e063aa3-1716030289891" {
		t.Error("unexpected recordID:", rID)
	}

	path = "/playback/video/9b897750e3453b1daa4563788af47ef90e063aa3-1716030289891/video-js/video.min.js"
	rID, ok = parseRecordIDPath(path)
	if !ok {
		t.Error("expected record id match")
	}
	if rID != "9b897750e3453b1daa4563788af47ef90e063aa3-1716030289891" {
		t.Error("unexpected recordID:", rID)
	}

	path = "/playback/video/9b897750e3453b1daa4563788af47ef90e063aa3-1716030289891/video-0.m4v"
	rID, ok = parseRecordIDPath(path)
	if !ok {
		t.Error("expected record id match")
	}
	if rID != "9b897750e3453b1daa4563788af47ef90e063aa3-1716030289891" {
		t.Error("unexpected recordID:", rID)
	}

}

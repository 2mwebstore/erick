package handlers

import (
	"encoding/json"
	"io"
	"strings"
	"testing"
)

// decodeMessage exists because "Malformed request." was returned for every
// decoding failure, which made a skew between a loaded page and a deployed
// server impossible to diagnose from the browser. These assert that each
// failure names itself.
func TestDecodeMessageNamesTheProblem(t *testing.T) {
	type target struct {
		Title string `json:"title"`
		Order int    `json:"order"`
	}

	decodeInto := func(body string) error {
		decoder := json.NewDecoder(strings.NewReader(body))
		decoder.DisallowUnknownFields()
		return decoder.Decode(&target{})
	}

	cases := []struct {
		name     string
		body     string
		contains string
	}{
		{"empty body", ``, "empty"},
		{"truncated json", `{"title": `, "not valid JSON"},
		{"wrong type", `{"order": "not a number"}`, "wrong type"},
		{"unknown field", `{"nonsense": "x"}`, "unexpected field"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := decodeInto(tc.body)
			if err == nil {
				t.Fatal("expected a decoding error")
			}

			got := decodeMessage(err)
			if !strings.Contains(strings.ToLower(got), strings.ToLower(tc.contains)) {
				t.Errorf("decodeMessage(%v) = %q, want it to mention %q", err, got, tc.contains)
			}
			if got == "Malformed request." {
				t.Errorf("%s fell through to the generic message", tc.name)
			}
		})
	}
}

// An unknown field is the version-skew case, so the message has to say what to
// do about it rather than only what went wrong.
func TestDecodeMessageTellsTheEditorToReload(t *testing.T) {
	decoder := json.NewDecoder(strings.NewReader(`{"gone":"x"}`))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&struct{}{})

	got := decodeMessage(err)
	if !strings.Contains(got, "reload") {
		t.Errorf("message %q should tell the editor to reload", got)
	}
	if !strings.Contains(got, `"gone"`) {
		t.Errorf("message %q should name the offending field", got)
	}
}

// Anything unrecognised still gets an answer rather than an empty string.
func TestDecodeMessageFallsBack(t *testing.T) {
	if got := decodeMessage(io.ErrClosedPipe); got != "Malformed request." {
		t.Errorf("unexpected fallback: %q", got)
	}
}

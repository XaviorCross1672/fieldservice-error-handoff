package fieldservice

import (
	"context"
	"encoding/json"
	"testing"
)

type recordedCall struct {
	method, path, key string
	payload           any
}

type stubCaller struct {
	enabled bool
	calls   []recordedCall
}

func (s *stubCaller) Call(_ context.Context, method, path string, payload any, key string, out any) error {
	s.calls = append(s.calls, recordedCall{method: method, path: path, key: key, payload: payload})
	if out != nil {
		raw, _ := json.Marshal(flagValue{Value: s.enabled})
		return json.Unmarshal(raw, out)
	}
	return nil
}

func TestCaptureDispatchFailureFollowUpDecision(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		want     bool
		wantText string
	}{
		{name: "flag enabled queues technician", enabled: true, want: true, wantText: "technician follow-up queued"},
		{name: "flag disabled leaves triage only", enabled: false, want: false, wantText: "captured for dispatch triage"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stub := &stubCaller{enabled: tc.enabled}
			got, err := CaptureDispatchFailure(context.Background(), stub, WorkOrderFailure{
				WorkOrderID: "WO-1842", DispatchStatus: "photo_upload_failed", PhotoCount: 3,
				TechnicianID: "tech-17", Exception: "image checksum mismatch",
			})
			if err != nil {
				t.Fatal(err)
			}
			if !got.Captured || got.Required != tc.want || got.Reason != tc.wantText {
				t.Fatalf("got %+v", got)
			}
			if len(stub.calls) != 2 || stub.calls[0].path != "/v1/errors/capture" || stub.calls[1].path != "/v1/flags/get_value/fieldservice-photo-follow-up" {
				t.Fatalf("unexpected handoff: %+v", stub.calls)
			}
			if stub.calls[0].key == "" {
				t.Fatal("capture request has no idempotency key")
			}
		})
	}
}

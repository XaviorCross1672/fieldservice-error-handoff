package fieldservice

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
)

type Caller interface {
	Call(context.Context, string, string, any, string, any) error
}

type WorkOrderFailure struct {
	WorkOrderID    string
	DispatchStatus string
	PhotoCount     int
	TechnicianID   string
	Exception      string
}

type FollowUp struct {
	Captured     bool
	Required     bool
	TechnicianID string
	Reason       string
}

type capturePayload struct {
	Title       string         `json:"title"`
	Message     string         `json:"message"`
	Level       string         `json:"level"`
	Fingerprint []string       `json:"fingerprint"`
	Exception   string         `json:"exception"`
	Context     map[string]any `json:"context"`
}

type flagValue struct {
	Value bool `json:"value"`
}

func CaptureDispatchFailure(ctx context.Context, api Caller, failure WorkOrderFailure) (FollowUp, error) {
	payload := capturePayload{
		Title:       "work-order dispatch failed",
		Message:     fmt.Sprintf("dispatch %s with %d photos", failure.DispatchStatus, failure.PhotoCount),
		Level:       "error",
		Fingerprint: []string{"work-order-dispatch", failure.DispatchStatus},
		Exception:   failure.Exception,
		Context: map[string]any{
			"work_order_id":   failure.WorkOrderID,
			"dispatch_status": failure.DispatchStatus,
			"photo_count":     failure.PhotoCount,
			"technician_id":   failure.TechnicianID,
		},
	}
	idempotencyKey := fmt.Sprintf("work-order-error-%x", sha256.Sum256([]byte(failure.WorkOrderID+"\x00"+failure.DispatchStatus+"\x00"+failure.Exception)))
	if err := api.Call(ctx, http.MethodPost, "/v1/errors/capture", payload, idempotencyKey, nil); err != nil {
		return FollowUp{}, err
	}

	var flag flagValue
	if err := api.Call(ctx, http.MethodGet, "/v1/flags/get_value/fieldservice-photo-follow-up", nil, "", &flag); err != nil {
		return FollowUp{}, err
	}
	reason := "captured for dispatch triage"
	if flag.Value {
		reason = "technician follow-up queued"
	}
	return FollowUp{Captured: true, Required: flag.Value, TechnicianID: failure.TechnicianID, Reason: reason}, nil
}

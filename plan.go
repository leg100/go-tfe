package tfe

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"
)

// Compile-time proof of interface implementation.
var _ Plans = (*plans)(nil)

// Plans describes all the plan related methods that the Terraform Enterprise
// API supports.
//
// TFE API docs: https://www.terraform.io/docs/enterprise/api/plan.html
type Plans interface {
	// Read a plan by its ID.
	Read(ctx context.Context, planID string) (*Plan, error)

	// Logs retrieves the logs of a plan.
	Logs(ctx context.Context, planID string) (io.Reader, error)

	// Retrieve the JSON execution plan
	JSONOutput(ctx context.Context, planID string) ([]byte, error)
}

// plans implements Plans.
type plans struct {
	client *Client
}

// PlanStatus represents a plan state.
type PlanStatus string

//List all available plan statuses.
const (
	PlanCanceled    PlanStatus = "canceled"
	PlanCreated     PlanStatus = "created"
	PlanErrored     PlanStatus = "errored"
	PlanFinished    PlanStatus = "finished"
	PlanMFAWaiting  PlanStatus = "mfa_waiting"
	PlanPending     PlanStatus = "pending"
	PlanQueued      PlanStatus = "queued"
	PlanRunning     PlanStatus = "running"
	PlanUnreachable PlanStatus = "unreachable"
)

// Plan represents a Terraform Enterprise plan.
type Plan struct {
	ID                   string                `jsonapi:"primary,plans"`
	HasChanges           bool                  `jsonapi:"attr,has-changes"`
	LogReadURL           string                `jsonapi:"attr,log-read-url"`
	ResourceAdditions    int                   `jsonapi:"attr,resource-additions"`
	ResourceChanges      int                   `jsonapi:"attr,resource-changes"`
	ResourceDestructions int                   `jsonapi:"attr,resource-destructions"`
	Status               PlanStatus            `jsonapi:"attr,status"`
	StatusTimestamps     *PlanStatusTimestamps `jsonapi:"attr,status-timestamps"`

	// Relations
	Exports []*PlanExport `jsonapi:"relation,exports"`
}

// PlanStatusTimestamps holds the timestamps for individual plan statuses.
type PlanStatusTimestamps struct {
	CanceledAt      *time.Time `json:"canceled-at,rfc3339,omitempty"`
	ErroredAt       *time.Time `json:"errored-at,rfc3339,omitempty"`
	FinishedAt      *time.Time `json:"finished-at,rfc3339,omitempty"`
	ForceCanceledAt *time.Time `json:"force-canceled-at,rfc3339,omitempty"`
	QueuedAt        *time.Time `json:"queued-at,rfc3339,omitempty"`
	StartedAt       *time.Time `json:"started-at,rfc3339,omitempty"`
}

// Read a plan by its ID.
func (s *plans) Read(ctx context.Context, planID string) (*Plan, error) {
	if !validStringID(&planID) {
		return nil, errors.New("invalid value for plan ID")
	}

	u := fmt.Sprintf("plans/%s", url.QueryEscape(planID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	p := &Plan{}
	err = s.client.do(ctx, req, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Logs retrieves the logs of a plan.
func (s *plans) Logs(ctx context.Context, planID string) (io.Reader, error) {
	if !validStringID(&planID) {
		return nil, errors.New("invalid value for plan ID")
	}

	// Get the plan to make sure it exists.
	p, err := s.Read(ctx, planID)
	if err != nil {
		return nil, err
	}

	// Return an error if the log URL is empty.
	if p.LogReadURL == "" {
		return nil, fmt.Errorf("plan %s does not have a log URL", planID)
	}

	u, err := url.Parse(p.LogReadURL)
	if err != nil {
		return nil, fmt.Errorf("invalid log URL: %v", err)
	}

	done := func() (bool, error) {
		p, err := s.Read(ctx, p.ID)
		if err != nil {
			return false, err
		}

		switch p.Status {
		case PlanCanceled, PlanErrored, PlanFinished, PlanUnreachable:
			return true, nil
		default:
			return false, nil
		}
	}

	return &LogReader{
		client: s.client,
		ctx:    ctx,
		done:   done,
		logURL: u,
	}, nil
}

// Retrieve the JSON execution plan
func (s *plans) JSONOutput(ctx context.Context, planID string) ([]byte, error) {
	if !validStringID(&planID) {
		return nil, errors.New("invalid value for plan ID")
	}

	u := fmt.Sprintf("plans/%s/json-output", url.QueryEscape(planID))
	req, err := s.client.newRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = s.client.do(ctx, req, &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

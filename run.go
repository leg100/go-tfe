package tfe

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"
)

var (
	// Compile-time proof of interface implementation.
	_ Runs = (*runs)(nil)

	PlanBinaryFormat = "binary"
	PlanJSONFormat   = "json"
)

// Runs describes all the run related methods that the Terraform Enterprise
// API supports.
//
// TFE API docs: https://www.terraform.io/docs/enterprise/api/run.html
type Runs interface {
	// List all the runs of the given workspace.
	List(ctx context.Context, workspaceID string, options RunListOptions) (*RunList, error)

	// Create a new run with the given options.
	Create(ctx context.Context, options RunCreateOptions) (*Run, error)

	// Read a run by its ID.
	Read(ctx context.Context, runID string) (*Run, error)

	// ReadWithOptions reads a run by its ID using the options supplied
	ReadWithOptions(ctx context.Context, runID string, options RunReadOptions) (*Run, error)

	// Apply a run by its ID.
	Apply(ctx context.Context, runID string, options RunApplyOptions) error

	// Cancel a run by its ID.
	Cancel(ctx context.Context, runID string, options RunCancelOptions) error

	// Force-cancel a run by its ID.
	ForceCancel(ctx context.Context, runID string, options RunForceCancelOptions) error

	// Discard a run by its ID.
	Discard(ctx context.Context, runID string, options RunDiscardOptions) error

	// GetPlanFile gets the plan file for a run by its run ID
	GetPlanFile(ctx context.Context, runID string, options PlanFileOptions) ([]byte, error)

	// UploadPlanFile uploads the plan file for a run by its run ID
	UploadPlanFile(ctx context.Context, runID string, plan []byte, options PlanFileOptions) error

	// UploadLogs uploads logs for a run. For use by an agent rather than user.
	UploadLogs(ctx context.Context, runID string, chunk []byte, options RunUploadLogsOptions) error
}

// runs implements Runs.
type runs struct {
	client *Client
}

// RunStatus represents a run state.
type RunStatus string

//List all available run statuses.
const (
	RunApplied            RunStatus = "applied"
	RunApplyQueued        RunStatus = "apply_queued"
	RunApplying           RunStatus = "applying"
	RunCanceled           RunStatus = "canceled"
	RunConfirmed          RunStatus = "confirmed"
	RunCostEstimated      RunStatus = "cost_estimated"
	RunCostEstimating     RunStatus = "cost_estimating"
	RunDiscarded          RunStatus = "discarded"
	RunErrored            RunStatus = "errored"
	RunPending            RunStatus = "pending"
	RunPlanQueued         RunStatus = "plan_queued"
	RunPlanned            RunStatus = "planned"
	RunPlannedAndFinished RunStatus = "planned_and_finished"
	RunPlanning           RunStatus = "planning"
	RunPolicyChecked      RunStatus = "policy_checked"
	RunPolicyChecking     RunStatus = "policy_checking"
	RunPolicyOverride     RunStatus = "policy_override"
	RunPolicySoftFailed   RunStatus = "policy_soft_failed"
)

// RunSource represents a source type of a run.
type RunSource string

// List all available run sources.
const (
	RunSourceAPI                  RunSource = "tfe-api"
	RunSourceConfigurationVersion RunSource = "tfe-configuration-version"
	RunSourceUI                   RunSource = "tfe-ui"
)

// RunList represents a list of runs.
type RunList struct {
	*Pagination
	Items []*Run
}

// Run represents a Terraform Enterprise run.
type Run struct {
	ID                     string               `jsonapi:"primary,runs"`
	Actions                *RunActions          `jsonapi:"attr,actions"`
	CreatedAt              time.Time            `jsonapi:"attr,created-at,iso8601"`
	ForceCancelAvailableAt time.Time            `jsonapi:"attr,force-cancel-available-at,iso8601"`
	HasChanges             bool                 `jsonapi:"attr,has-changes"`
	IsDestroy              bool                 `jsonapi:"attr,is-destroy"`
	Message                string               `jsonapi:"attr,message"`
	Permissions            *RunPermissions      `jsonapi:"attr,permissions"`
	PositionInQueue        int                  `jsonapi:"attr,position-in-queue"`
	Refresh                bool                 `jsonapi:"attr,refresh"`
	RefreshOnly            bool                 `jsonapi:"attr,refresh-only"`
	ReplaceAddrs           []string             `jsonapi:"attr,replace-addrs,omitempty"`
	Source                 RunSource            `jsonapi:"attr,source"`
	Status                 RunStatus            `jsonapi:"attr,status"`
	StatusTimestamps       *RunStatusTimestamps `jsonapi:"attr,status-timestamps"`
	TargetAddrs            []string             `jsonapi:"attr,target-addrs,omitempty"`

	// Relations
	Apply                *Apply                `jsonapi:"relation,apply"`
	ConfigurationVersion *ConfigurationVersion `jsonapi:"relation,configuration-version"`
	CostEstimate         *CostEstimate         `jsonapi:"relation,cost-estimate"`
	CreatedBy            *User                 `jsonapi:"relation,created-by"`
	Plan                 *Plan                 `jsonapi:"relation,plan"`
	PolicyChecks         []*PolicyCheck        `jsonapi:"relation,policy-checks"`
	Workspace            *Workspace            `jsonapi:"relation,workspace"`
}

// RunActions represents the run actions.
type RunActions struct {
	IsCancelable      bool `json:"is-cancelable"`
	IsConfirmable     bool `json:"is-confirmable"`
	IsDiscardable     bool `json:"is-discardable"`
	IsForceCancelable bool `json:"is-force-cancelable"`
}

// RunPermissions represents the run permissions.
type RunPermissions struct {
	CanApply        bool `json:"can-apply"`
	CanCancel       bool `json:"can-cancel"`
	CanDiscard      bool `json:"can-discard"`
	CanForceCancel  bool `json:"can-force-cancel"`
	CanForceExecute bool `json:"can-force-execute"`
}

// RunStatusTimestamps holds the timestamps for individual run statuses.
type RunStatusTimestamps struct {
	AppliedAt            *time.Time `json:"applied-at,rfc3339,omitempty"`
	ApplyQueuedAt        *time.Time `json:"apply-queued-at,rfc3339,omitempty"`
	ApplyingAt           *time.Time `json:"applying-at,rfc3339,omitempty"`
	CanceledAt           *time.Time `json:"canceled-at,rfc3339,omitempty"`
	ConfirmedAt          *time.Time `json:"confirmed-at,rfc3339,omitempty"`
	CostEstimatedAt      *time.Time `json:"cost-estimated-at,rfc3339,omitempty"`
	CostEstimatingAt     *time.Time `json:"cost-estimating-at,rfc3339,omitempty"`
	DiscardedAt          *time.Time `json:"discarded-at,rfc3339,omitempty"`
	ErroredAt            *time.Time `json:"errored-at,rfc3339,omitempty"`
	ForceCanceledAt      *time.Time `json:"force-canceled-at,rfc3339,omitempty"`
	PlanQueueableAt      *time.Time `json:"plan-queueable-at,rfc3339,omitempty"`
	PlanQueuedAt         *time.Time `json:"plan-queued-at,rfc3339,omitempty"`
	PlannedAndFinishedAt *time.Time `json:"planned-and-finished-at,rfc3339,omitempty"`
	PlannedAt            *time.Time `json:"planned-at,rfc3339,omitempty"`
	PlanningAt           *time.Time `json:"planning-at,rfc3339,omitempty"`
	PolicyCheckedAt      *time.Time `json:"policy-checked-at,rfc3339,omitempty"`
	PolicySoftFailedAt   *time.Time `json:"policy-soft-failed-at,rfc3339,omitempty"`
}

// RunListOptions represents the options for listing runs.
type RunListOptions struct {
	ListOptions

	// A list of relations to include. See available resources:
	// https://www.terraform.io/docs/cloud/api/run.html#available-related-resources
	Include *string `schema:"include"`
}

// List all the runs of the given workspace.
func (s *runs) List(ctx context.Context, workspaceID string, options RunListOptions) (*RunList, error) {
	if !validStringID(&workspaceID) {
		return nil, ErrInvalidWorkspaceID
	}

	u := fmt.Sprintf("workspaces/%s/runs", url.QueryEscape(workspaceID))
	req, err := s.client.newRequest("GET", u, &options)
	if err != nil {
		return nil, err
	}

	rl := &RunList{}
	err = s.client.do(ctx, req, rl)
	if err != nil {
		return nil, err
	}

	return rl, nil
}

// RunCreateOptions represents the options for creating a new run.
type RunCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,runs"`

	// Specifies if this plan is a destroy plan, which will destroy all
	// provisioned resources.
	IsDestroy *bool `jsonapi:"attr,is-destroy,omitempty"`

	// Refresh determines if the run should
	// update the state prior to checking for differences
	Refresh *bool `jsonapi:"attr,refresh,omitempty"`

	// RefreshOnly determines whether the run should ignore config changes
	// and refresh the state only
	RefreshOnly *bool `jsonapi:"attr,refresh-only,omitempty"`

	// Specifies the message to be associated with this run.
	Message *string `jsonapi:"attr,message,omitempty"`

	// Specifies the configuration version to use for this run. If the
	// configuration version object is omitted, the run will be created using the
	// workspace's latest configuration version.
	ConfigurationVersion *ConfigurationVersion `jsonapi:"relation,configuration-version"`

	// Specifies the workspace where the run will be executed.
	Workspace *Workspace `jsonapi:"relation,workspace"`

	// If non-empty, requests that Terraform should create a plan including
	// actions only for the given objects (specified using resource address
	// syntax) and the objects they depend on.
	//
	// This capability is provided for exceptional circumstances only, such as
	// recovering from mistakes or working around existing Terraform
	// limitations. Terraform will generally mention the -target command line
	// option in its error messages describing situations where setting this
	// argument may be appropriate. This argument should not be used as part
	// of routine workflow and Terraform will emit warnings reminding about
	// this whenever this property is set.
	TargetAddrs []string `jsonapi:"attr,target-addrs,omitempty"`

	// If non-empty, requests that Terraform create a plan that replaces
	// (destroys and then re-creates) the objects specified by the given
	// resource addresses.
	ReplaceAddrs []string `jsonapi:"attr,replace-addrs,omitempty"`
}

func (o RunCreateOptions) valid() error {
	if o.Workspace == nil {
		return errors.New("workspace is required")
	}
	return nil
}

// Create a new run with the given options.
func (s *runs) Create(ctx context.Context, options RunCreateOptions) (*Run, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}

	req, err := s.client.newRequest("POST", "runs", &options)
	if err != nil {
		return nil, err
	}

	r := &Run{}
	err = s.client.do(ctx, req, r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// Read a run by its ID.
func (s *runs) Read(ctx context.Context, runID string) (*Run, error) {
	return s.ReadWithOptions(ctx, runID, RunReadOptions{})
}

// RunReadOptions represents the options for reading a run.
type RunReadOptions struct {
	Include string `schema:"include"`
}

// Read a run by its ID with the given options.
func (s *runs) ReadWithOptions(ctx context.Context, runID string, options RunReadOptions) (*Run, error) {
	if !validStringID(&runID) {
		return nil, ErrInvalidRunID
	}

	u := fmt.Sprintf("runs/%s", url.QueryEscape(runID))
	req, err := s.client.newRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	r := &Run{}
	err = s.client.do(ctx, req, r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// RunApplyOptions represents the options for applying a run.
type RunApplyOptions struct {
	// An optional comment about the run.
	Comment *string `jsonapi:"attr,comment,omitempty"`
}

// Apply a run by its ID.
func (s *runs) Apply(ctx context.Context, runID string, options RunApplyOptions) error {
	if !validStringID(&runID) {
		return ErrInvalidRunID
	}

	u := fmt.Sprintf("runs/%s/actions/apply", url.QueryEscape(runID))
	req, err := s.client.newRequest("POST", u, &options)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// RunCancelOptions represents the options for canceling a run.
type RunCancelOptions struct {
	// An optional explanation for why the run was canceled.
	Comment *string `jsonapi:"attr,comment,omitempty"`
}

// Cancel a run by its ID.
func (s *runs) Cancel(ctx context.Context, runID string, options RunCancelOptions) error {
	if !validStringID(&runID) {
		return ErrInvalidRunID
	}

	u := fmt.Sprintf("runs/%s/actions/cancel", url.QueryEscape(runID))
	req, err := s.client.newRequest("POST", u, &options)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// RunForceCancelOptions represents the options for force-canceling a run.
type RunForceCancelOptions struct {
	// An optional comment explaining the reason for the force-cancel.
	Comment *string `jsonapi:"attr,comment,omitempty"`
}

// ForceCancel is used to forcefully cancel a run by its ID.
func (s *runs) ForceCancel(ctx context.Context, runID string, options RunForceCancelOptions) error {
	if !validStringID(&runID) {
		return ErrInvalidRunID
	}

	u := fmt.Sprintf("runs/%s/actions/force-cancel", url.QueryEscape(runID))
	req, err := s.client.newRequest("POST", u, &options)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// RunDiscardOptions represents the options for discarding a run.
type RunDiscardOptions struct {
	// An optional explanation for why the run was discarded.
	Comment *string `jsonapi:"attr,comment,omitempty"`
}

// Discard a run by its ID.
func (s *runs) Discard(ctx context.Context, runID string, options RunDiscardOptions) error {
	if !validStringID(&runID) {
		return ErrInvalidRunID
	}

	u := fmt.Sprintf("runs/%s/actions/discard", url.QueryEscape(runID))
	req, err := s.client.newRequest("POST", u, &options)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// PlanFileOptions represents the options for getting the plan file for a run.
type PlanFileOptions struct {
	// Format of plan file. Valid values are json and binary.
	Format string `schema:"format"`
}

// GetPlanFile gets the plan file for a run.
func (s *runs) GetPlanFile(ctx context.Context, runID string, options PlanFileOptions) ([]byte, error) {
	u := fmt.Sprintf("/runs/%s/plan", url.QueryEscape(runID))

	req, err := s.client.newRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := s.client.do(ctx, req, &buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UploadPlan uploads the plan file for a run.
func (s *runs) UploadPlanFile(ctx context.Context, runID string, plan []byte, options PlanFileOptions) error {
	q := url.Values{}
	if err := encoder.Encode(options, q); err != nil {
		return err
	}

	u := url.URL{
		Path:     fmt.Sprintf("/runs/%s/plan", url.QueryEscape(runID)),
		RawQuery: q.Encode(),
	}

	req, err := s.client.newRequest("PUT", u.String(), plan)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// RunUploadLogsOptions represents the options for uploading logs for a run.
type RunUploadLogsOptions struct {
	// End indicates this is the last and final chunk
	End bool `schema:"end"`
}

// UploadLogs uploads logs for a run.
func (s *runs) UploadLogs(ctx context.Context, runID string, plan []byte, options RunUploadLogsOptions) error {
	q := url.Values{}
	if err := encoder.Encode(options, q); err != nil {
		return err
	}

	u := url.URL{
		Path:     fmt.Sprintf("/runs/%s/logs", url.QueryEscape(runID)),
		RawQuery: q.Encode(),
	}

	req, err := s.client.newRequest("PUT", u.String(), plan)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

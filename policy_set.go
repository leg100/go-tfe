package tfe

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// Compile-time proof of interface implementation.
var _ PolicySets = (*policySets)(nil)

// PolicySets describes all the policy set related methods that the Terraform
// Enterprise API supports.
//
// TFE API docs: https://www.terraform.io/docs/enterprise/api/policies.html
type PolicySets interface {
	// List all the policy sets for a given organization.
	List(ctx context.Context, organization string, options PolicySetListOptions) (*PolicySetList, error)

	// Create a policy set and associate it with an organization.
	Create(ctx context.Context, organization string, options PolicySetCreateOptions) (*PolicySet, error)

	// Read a policy set by its ID.
	Read(ctx context.Context, policySetID string) (*PolicySet, error)

	// ReadWithOptions reads a policy set by its ID using the options supplied.
	ReadWithOptions(ctx context.Context, policySetID string, options *PolicySetReadOptions) (*PolicySet, error)

	// Update an existing policy set.
	Update(ctx context.Context, policySetID string, options PolicySetUpdateOptions) (*PolicySet, error)

	// Add policies to a policy set. This function can only be used when
	// there is no VCS repository associated with the policy set.
	AddPolicies(ctx context.Context, policySetID string, options PolicySetAddPoliciesOptions) error

	// Remove policies from a policy set. This function can only be used
	// when there is no VCS repository associated with the policy set.
	RemovePolicies(ctx context.Context, policySetID string, options PolicySetRemovePoliciesOptions) error

	// Add workspaces to a policy set.
	AddWorkspaces(ctx context.Context, policySetID string, options PolicySetAddWorkspacesOptions) error

	// Remove workspaces from a policy set.
	RemoveWorkspaces(ctx context.Context, policySetID string, options PolicySetRemoveWorkspacesOptions) error

	// Delete a policy set by its ID.
	Delete(ctx context.Context, policyID string) error
}

// policySets implements PolicySets.
type policySets struct {
	client *Client
}

// PolicySetList represents a list of policy sets.
type PolicySetList struct {
	*Pagination
	Items []*PolicySet
}

// PolicySet represents a Terraform Enterprise policy set.
type PolicySet struct {
	ID             string    `jsonapi:"primary,policy-sets"`
	Name           string    `jsonapi:"attr,name"`
	Description    string    `jsonapi:"attr,description"`
	Global         bool      `jsonapi:"attr,global"`
	PoliciesPath   string    `jsonapi:"attr,policies-path"`
	PolicyCount    int       `jsonapi:"attr,policy-count"`
	VCSRepo        *VCSRepo  `jsonapi:"attr,vcs-repo"`
	WorkspaceCount int       `jsonapi:"attr,workspace-count"`
	CreatedAt      time.Time `jsonapi:"attr,created-at,iso8601"`
	UpdatedAt      time.Time `jsonapi:"attr,updated-at,iso8601"`

	// Relations
	// The organization to which the policy set belongs to.
	Organization *Organization `jsonapi:"relation,organization"`
	// The workspaces to which the policy set applies.
	Workspaces []*Workspace `jsonapi:"relation,workspaces"`
	// Individually managed policies which are associated with the policy set.
	Policies []*Policy `jsonapi:"relation,policies"`
	// The most recently created policy set version, regardless of status.
	// Note that this relationship may include an errored and unusable version,
	// and is intended to allow checking for errors.
	NewestVersion *PolicySetVersion `jsonapi:"relation,newest-version"`
	// The most recent successful policy set version.
	CurrentVersion *PolicySetVersion `jsonapi:"relation,current-version"`
}

// PolicySetListOptions represents the options for listing policy sets.
type PolicySetListOptions struct {
	ListOptions

	// A search string (partial policy set name) used to filter the results.
	Search *string `schema:"search[name],omitempty"`
}

// List all the policies for a given organization.
func (s *policySets) List(ctx context.Context, organization string, options PolicySetListOptions) (*PolicySetList, error) {
	if !validStringID(&organization) {
		return nil, ErrInvalidOrg
	}

	u := fmt.Sprintf("organizations/%s/policy-sets", url.QueryEscape(organization))
	req, err := s.client.newRequest("GET", u, &options)
	if err != nil {
		return nil, err
	}

	psl := &PolicySetList{}
	err = s.client.do(ctx, req, psl)
	if err != nil {
		return nil, err
	}

	return psl, nil
}

// PolicySetCreateOptions represents the options for creating a new policy set.
type PolicySetCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,policy-sets"`

	// The name of the policy set.
	Name *string `jsonapi:"attr,name"`

	// The description of the policy set.
	Description *string `jsonapi:"attr,description,omitempty"`

	// Whether or not the policy set is global.
	Global *bool `jsonapi:"attr,global,omitempty"`

	// The sub-path within the attached VCS repository to ingress. All
	// files and directories outside of this sub-path will be ignored.
	// This option may only be specified when a VCS repo is present.
	PoliciesPath *string `jsonapi:"attr,policies-path,omitempty"`

	// The initial members of the policy set.
	Policies []*Policy `jsonapi:"relation,policies,omitempty"`

	// VCS repository information. When present, the policies and
	// configuration will be sourced from the specified VCS repository
	// instead of being defined within the policy set itself. Note that
	// this option is mutually exclusive with the Policies option and
	// both cannot be used at the same time.
	VCSRepo *VCSRepoOptions `jsonapi:"attr,vcs-repo,omitempty"`

	// The initial list of workspaces for which the policy set should be enforced.
	Workspaces []*Workspace `jsonapi:"relation,workspaces,omitempty"`
}

func (o PolicySetCreateOptions) valid() error {
	if !validString(o.Name) {
		return ErrRequiredName
	}
	if !validStringID(o.Name) {
		return ErrInvalidName
	}
	return nil
}

// Create a policy set and associate it with an organization.
func (s *policySets) Create(ctx context.Context, organization string, options PolicySetCreateOptions) (*PolicySet, error) {
	if !validStringID(&organization) {
		return nil, ErrInvalidOrg
	}
	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("organizations/%s/policy-sets", url.QueryEscape(organization))
	req, err := s.client.newRequest("POST", u, &options)
	if err != nil {
		return nil, err
	}

	ps := &PolicySet{}
	err = s.client.do(ctx, req, ps)
	if err != nil {
		return nil, err
	}

	return ps, err
}

// PolicySetReadOptions are read options.
// For a full list of relations, please see:
// https://www.terraform.io/docs/cloud/api/policy-sets.html#relationships
type PolicySetReadOptions struct {
	Include string `schema:"include"`
}

// Read a policy set by its ID.
func (s *policySets) Read(ctx context.Context, policySetID string) (*PolicySet, error) {
	return s.ReadWithOptions(ctx, policySetID, nil)
}

// ReadWithOptions reads a policy by its ID using the options supplied.
func (s *policySets) ReadWithOptions(ctx context.Context, policySetID string, options *PolicySetReadOptions) (*PolicySet, error) {
	if !validStringID(&policySetID) {
		return nil, errors.New("invalid value for policy set ID")
	}

	u := fmt.Sprintf("policy-sets/%s", url.QueryEscape(policySetID))
	req, err := s.client.newRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	ps := &PolicySet{}
	err = s.client.do(ctx, req, ps)
	if err != nil {
		return nil, err
	}

	return ps, err
}

// PolicySetUpdateOptions represents the options for updating a policy set.
type PolicySetUpdateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,policy-sets"`

	/// The name of the policy set.
	Name *string `jsonapi:"attr,name,omitempty"`

	// The description of the policy set.
	Description *string `jsonapi:"attr,description,omitempty"`

	// Whether or not the policy set is global.
	Global *bool `jsonapi:"attr,global,omitempty"`

	// The sub-path within the attached VCS repository to ingress. All
	// files and directories outside of this sub-path will be ignored.
	// This option may only be specified when a VCS repo is present.
	PoliciesPath *string `jsonapi:"attr,policies-path,omitempty"`

	// VCS repository information. When present, the policies and
	// configuration will be sourced from the specified VCS repository
	// instead of being defined within the policy set itself. Note that
	// specifying this option may only be used on policy sets with no
	// directly-attached policies (*PolicySet.Policies). Specifying this
	// option when policies are already present will result in an error.
	VCSRepo *VCSRepoOptions `jsonapi:"attr,vcs-repo,omitempty"`
}

func (o PolicySetUpdateOptions) valid() error {
	if o.Name != nil && !validStringID(o.Name) {
		return ErrInvalidName
	}
	return nil
}

// Update an existing policy set.
func (s *policySets) Update(ctx context.Context, policySetID string, options PolicySetUpdateOptions) (*PolicySet, error) {
	if !validStringID(&policySetID) {
		return nil, errors.New("invalid value for policy set ID")
	}
	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("policy-sets/%s", url.QueryEscape(policySetID))
	req, err := s.client.newRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	ps := &PolicySet{}
	err = s.client.do(ctx, req, ps)
	if err != nil {
		return nil, err
	}

	return ps, err
}

// PolicySetAddPoliciesOptions represents the options for adding policies
// to a policy set.
type PolicySetAddPoliciesOptions struct {
	/// The policies to add to the policy set.
	Policies []*Policy
}

func (o PolicySetAddPoliciesOptions) valid() error {
	if o.Policies == nil {
		return errors.New("policies is required")
	}
	if len(o.Policies) == 0 {
		return errors.New("must provide at least one policy")
	}
	return nil
}

// Add policies to a policy set
func (s *policySets) AddPolicies(ctx context.Context, policySetID string, options PolicySetAddPoliciesOptions) error {
	if !validStringID(&policySetID) {
		return errors.New("invalid value for policy set ID")
	}
	if err := options.valid(); err != nil {
		return err
	}

	u := fmt.Sprintf("policy-sets/%s/relationships/policies", url.QueryEscape(policySetID))
	req, err := s.client.newRequest("POST", u, options.Policies)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// PolicySetRemovePoliciesOptions represents the options for removing
// policies from a policy set.
type PolicySetRemovePoliciesOptions struct {
	/// The policies to remove from the policy set.
	Policies []*Policy
}

func (o PolicySetRemovePoliciesOptions) valid() error {
	if o.Policies == nil {
		return errors.New("policies is required")
	}
	if len(o.Policies) == 0 {
		return errors.New("must provide at least one policy")
	}
	return nil
}

// Remove policies from a policy set
func (s *policySets) RemovePolicies(ctx context.Context, policySetID string, options PolicySetRemovePoliciesOptions) error {
	if !validStringID(&policySetID) {
		return errors.New("invalid value for policy set ID")
	}
	if err := options.valid(); err != nil {
		return err
	}

	u := fmt.Sprintf("policy-sets/%s/relationships/policies", url.QueryEscape(policySetID))
	req, err := s.client.newRequest("DELETE", u, options.Policies)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// PolicySetAddWorkspacesOptions represents the options for adding workspaces
// to a policy set.
type PolicySetAddWorkspacesOptions struct {
	/// The workspaces to add to the policy set.
	Workspaces []*Workspace
}

func (o PolicySetAddWorkspacesOptions) valid() error {
	if o.Workspaces == nil {
		return errors.New("workspaces is required")
	}
	if len(o.Workspaces) == 0 {
		return errors.New("must provide at least one workspace")
	}
	return nil
}

// Add workspaces to a policy set.
func (s *policySets) AddWorkspaces(ctx context.Context, policySetID string, options PolicySetAddWorkspacesOptions) error {
	if !validStringID(&policySetID) {
		return errors.New("invalid value for policy set ID")
	}
	if err := options.valid(); err != nil {
		return err
	}

	u := fmt.Sprintf("policy-sets/%s/relationships/workspaces", url.QueryEscape(policySetID))
	req, err := s.client.newRequest("POST", u, options.Workspaces)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// PolicySetRemoveWorkspacesOptions represents the options for removing
// workspaces from a policy set.
type PolicySetRemoveWorkspacesOptions struct {
	/// The workspaces to remove from the policy set.
	Workspaces []*Workspace
}

func (o PolicySetRemoveWorkspacesOptions) valid() error {
	if o.Workspaces == nil {
		return errors.New("workspaces is required")
	}
	if len(o.Workspaces) == 0 {
		return errors.New("must provide at least one workspace")
	}
	return nil
}

// Remove workspaces from a policy set.
func (s *policySets) RemoveWorkspaces(ctx context.Context, policySetID string, options PolicySetRemoveWorkspacesOptions) error {
	if !validStringID(&policySetID) {
		return errors.New("invalid value for policy set ID")
	}
	if err := options.valid(); err != nil {
		return err
	}

	u := fmt.Sprintf("policy-sets/%s/relationships/workspaces", url.QueryEscape(policySetID))
	req, err := s.client.newRequest("DELETE", u, options.Workspaces)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// Delete a policy set by its ID.
func (s *policySets) Delete(ctx context.Context, policySetID string) error {
	if !validStringID(&policySetID) {
		return errors.New("invalid value for policy set ID")
	}

	u := fmt.Sprintf("policy-sets/%s", url.QueryEscape(policySetID))
	req, err := s.client.newRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

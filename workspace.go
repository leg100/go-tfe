package tfe

import (
	"errors"
)

// Workspace encapsulates all data fields of a workspace in TFE.
type Workspace struct {
	// Unique ID of this workspace. This ID is guaranteed unique within the
	// context of the TFE instance.
	ID *string `json:"id,omitempty"`

	// Name of the workspace. This value is only guaranteed unique within
	// an organization.
	Name *string `json:"name,omitempty"`

	// Creation time of the workspace.
	CreatedAt *string `json:"created-at,omitempty"`

	// Indicates if plans will be automatically applied (without confirmation).
	AutoApply *bool `json:"auto-apply,omitempty"`

	// The working directory used by Terraform during runs.
	WorkingDirectory *string `json:"working-directory,omitempty"`

	// The version of Terraform which will be used to execute plan and
	// apply operations for this workspace.
	TerraformVersion *string `json:"terraform-version,omitempty"`

	// VCSRepo holds the VCS settings for this workspace.
	VCSRepo *VCSRepo `json:"vcs-repo,omitempty"`
}

// Workspaces returns all of the workspaces within an organization.
func (c *Client) Workspaces(organization string) ([]*Workspace, error) {
	var result jsonapiWorkspaces

	if _, err := c.do(&request{
		method: "GET",
		path:   "/api/v2/organizations/" + organization + "/workspaces",
		output: &result,
	}); err != nil {
		return nil, err
	}

	output := make([]*Workspace, len(result))
	for i, ws := range result {
		output[i] = ws.Workspace
	}

	return output, nil
}

// Workspace returns the workspace identified by the given org and name.
func (c *Client) Workspace(organization, workspace string) (*Workspace, error) {
	var output jsonapiWorkspace

	if _, err := c.do(&request{
		method: "GET",
		path:   "/api/v2/organizations/" + organization + "/workspaces/" + workspace,
		output: &output,
	}); err != nil {
		return nil, err
	}

	return output.Workspace, nil
}

// CreateWorkspaceInput contains the parameters used for creating new
// new workspaces within an existing organization.
type CreateWorkspaceInput struct {
	// The organization name to create the workspace in.
	Organization *string

	// The name of the workspace
	Name *string

	// Determines if plans should automatically apply. Use this option with
	// caution - unexpected changes could be deployed to your infrastructure
	// if this is set to true.
	AutoApply *bool

	// The Terraform version number to run this workspace's configuration.
	// Setting this to "latest" will track the latest available version of
	// Terraform known to the TFE instance.
	TerraformVersion *string

	// An optional subdirectory to use as the "root" of the Terraform
	// configuration. TFE will change to this directory before running any
	// Terraform CLI commands against the configuration.
	WorkingDirectory *string

	VCSRepo *VCSRepo
}

// CreateWorkspaceOutput holds the return values from a workspace creation
// request.
type CreateWorkspaceOutput struct {
	// A reference to the newly created workspace.
	Workspace *Workspace
}

// CreateWorkspace is used to create a new workspace with the given parameters.
func (c *Client) CreateWorkspace(input *CreateWorkspaceInput) (
	*CreateWorkspaceOutput, error) {

	if input.Organization == nil || input.Name == nil {
		return nil, errors.New("Organization and Name are required")
	}

	// Create the special JSONAPI payload.
	jsonapiParams := jsonapiWorkspace{
		Workspace: &Workspace{
			Name:             input.Name,
			AutoApply:        input.AutoApply,
			TerraformVersion: input.TerraformVersion,
			WorkingDirectory: input.WorkingDirectory,
			VCSRepo:          input.VCSRepo,
		},
	}

	var output jsonapiWorkspace

	// Send the request.
	if _, err := c.do(&request{
		method: "POST",
		path:   "/api/v2/organizations/" + *input.Organization + "/workspaces",
		input:  jsonapiParams,
		output: &output,
	}); err != nil {
		return nil, err
	}

	return &CreateWorkspaceOutput{
		Workspace: output.Workspace,
	}, nil
}

// DeleteWorkspaceInput carries the parameters used for deleting workspaces.
type DeleteWorkspaceInput struct {
	// Organization is the name of the organization in which the workspace
	// exists.
	Organization *string

	// Name is the name of the workspace to delete.
	Name *string
}

// DeleteWorkspaceOutput holds the return values from deleting a workspace.
type DeleteWorkspaceOutput struct{}

// DeleteWorkspace is used to delete a single workspace.
func (c *Client) DeleteWorkspace(input *DeleteWorkspaceInput) (
	*DeleteWorkspaceOutput, error) {

	if input.Organization == nil || input.Name == nil {
		return nil, errors.New("Organization and Name are required")
	}

	if _, err := c.do(&request{
		method: "DELETE",
		path:   "/api/v2/organizations/" + *input.Organization + "/workspaces/" + *input.Name,
	}); err != nil {
		return nil, err
	}

	return &DeleteWorkspaceOutput{}, nil
}

// WorkspaceOrganizationSort provides sorting by the workspace name.
type WorkspaceNameSort []*Workspace

func (w WorkspaceNameSort) Len() int           { return len(w) }
func (w WorkspaceNameSort) Less(a, b int) bool { return *w[a].Name < *w[b].Name }
func (w WorkspaceNameSort) Swap(a, b int)      { w[a], w[b] = w[b], w[a] }

// Internal type to satisfy the jsonapi interface for a single workspace.
type jsonapiWorkspace struct{ *Workspace }

func (jsonapiWorkspace) GetName() string    { return "workspaces" }
func (jsonapiWorkspace) GetID() string      { return "" }
func (jsonapiWorkspace) SetID(string) error { return nil }
func (jsonapiWorkspace) SetToOneReferenceID(a, b string) error {
	return nil
}

// Internal type to satisfy the jsonapi interface for workspace indexes.
type jsonapiWorkspaces []jsonapiWorkspace

func (jsonapiWorkspaces) GetName() string    { return "workspaces" }
func (jsonapiWorkspaces) GetID() string      { return "" }
func (jsonapiWorkspaces) SetID(string) error { return nil }
func (jsonapiWorkspaces) SetToOneReferenceID(a, b string) error {
	return nil
}

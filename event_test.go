package tfe

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvents(t *testing.T) {
	client := testClient(t)

	sub, err := client.Events.Subscribe("dummy-id")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, sub.Close())
	}()

	_, rTestCleanup := createPlannedRun(t, client, nil)
	defer rTestCleanup()

	orgCreated := <-sub.C()
	assert.Equal(t, EventOrganizationCreated, orgCreated.Type)

	workspaceCreated := <-sub.C()
	assert.Equal(t, EventWorkspaceCreated, workspaceCreated.Type)

	runCreated := <-sub.C()
	assert.Equal(t, EventRunCreated, runCreated.Type)

	runQueued := <-sub.C()
	assert.Equal(t, EventPlanQueued, runQueued.Type)

	runPlanned := <-sub.C()
	assert.Equal(t, EventRunPlanned, runPlanned.Type)
}

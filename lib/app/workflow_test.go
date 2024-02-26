package app_test

import (
	"sync"
	"testing"

	"github.com/stackup-app/stackup/lib/app"
	"github.com/stackup-app/stackup/lib/gateway"
	"github.com/stretchr/testify/assert"
)

func TestCreateWorkflow(t *testing.T) {
	gw := &gateway.Gateway{} // Mock or initialize as needed
	processMap := &sync.Map{}
	workflow := app.CreateWorkflow(gw, processMap)

	assert.NotNil(t, workflow)
	assert.NotNil(t, workflow.Settings)
	assert.NotNil(t, workflow.Preconditions)
	assert.NotNil(t, workflow.Tasks)
	assert.NotNil(t, workflow.Includes)
	assert.NotNil(t, workflow.Gateway)
	assert.NotNil(t, workflow.ProcessMap)
	assert.NotNil(t, workflow.Integrations)
}

func TestAsContract(t *testing.T) {
	workflow := &app.StackupWorkflow{}
	contract := workflow.AsContract()

	assert.NotNil(t, contract)
}

func TestFindTaskById(t *testing.T) {
	workflow := &app.StackupWorkflow{
		Tasks: []*app.Task{
			{Id: "test1"},
			{Id: "test2"},
		},
	}

	taskAny, found := workflow.FindTaskById("test1")
	var task *app.Task = taskAny.(*app.Task)

	assert.True(t, found)
	assert.Equal(t, "test1", task.Id)

	_, notFound := workflow.FindTaskById("test3")
	assert.False(t, notFound)
}

package app_test

import (
	"testing"

	"github.com/stackup-app/stackup/lib/app"
	"github.com/stretchr/testify/assert"
)

func TestGetDisplayName(t *testing.T) {
	assert := assert.New(t)

	// Test case: When Include field has a value with "https://"
	task1 := app.Task{
		Include: "https://example.com",
		Name:    "TestName",
		Id:      "TestId",
		Uuid:    "TestUuid",
	}
	assert.Equal("example.com", task1.GetDisplayName())

	// Test case: When Include field has a value without "https://"
	task2 := app.Task{
		Include: "example.com",
		Name:    "TestName",
		Id:      "TestId",
		Uuid:    "TestUuid",
	}
	assert.Equal("example.com", task2.GetDisplayName())

	// Test case: When Name field has a value
	task3 := app.Task{
		Name: "TestName",
		Id:   "TestId",
		Uuid: "TestUuid",
	}
	assert.Equal("TestName", task3.GetDisplayName())

	// Test case: When Id field has a value
	task4 := app.Task{
		Id:   "TestId",
		Uuid: "TestUuid",
	}
	assert.Equal("TestId", task4.GetDisplayName())

	// Test case: When only Uuid field has a value
	task5 := app.Task{
		Uuid: "TestUuid",
	}
	assert.Equal("TestUuid", task5.GetDisplayName())

	// Test case: When no fields have values
	task6 := app.Task{}
	assert.Equal("", task6.GetDisplayName())
}

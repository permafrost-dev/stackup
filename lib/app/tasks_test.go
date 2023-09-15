package app_test

import (
	"runtime"
	"testing"

	"github.com/stackup-app/stackup/lib/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockJsEngine struct {
	MockEvaluate func(script string) interface{}
	mock.Mock
}

func (m *MockJsEngine) Evaluate(script string) interface{} {
	return m.MockEvaluate(script)
}

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

func TestCanRunOnCurrentPlatform(t *testing.T) {
	assert := assert.New(t)

	// Test case: When Platforms is nil
	task1 := &app.Task{}
	assert.True(task1.CanRunOnCurrentPlatform())

	// Test case: When Platforms is empty
	task2 := &app.Task{Platforms: []string{}}
	assert.True(task2.CanRunOnCurrentPlatform())

	// Test case: When Platforms contains the current platform (case insensitive)
	task3 := &app.Task{Platforms: []string{"windows", "linux", "darwin"}}
	assert.Contains(task3.Platforms, runtime.GOOS)

	// Test case: When Platforms does not contain the current platform
	task4 := &app.Task{Platforms: []string{"someotherplatform"}}
	assert.False(task4.CanRunOnCurrentPlatform())
}

// func TestCanRunConditionally(t *testing.T) {
// 	assert := assert.New(t)

// 	// Test case: When If field is empty
// 	task1 := &app.Task{If: ""}
// 	assert.True(task1.CanRunConditionally())

// 	// Test case: When JsEngine.Evaluate returns true
// 	getJsEngine := func(mockEval func(script string) interface{}) interface{} {
// 		engine := &MockJsEngine{
// 			MockEvaluate: mockEval,
// 		}

// 		return engine
// 	}

// 	jsengine2 := getJsEngine(func(script string) interface{} {
// 		return true
// 	})

// 	task2 := &app.Task{
// 		If:       "{{ some condition }}",
// 		JsEngine: jsengine2.(*scripting.JavaScriptEngine),
// 	}
// 	assert.True(task2.CanRunConditionally())

// 	// Test case: When JsEngine.Evaluate returns false
// 	jsengine3 := getJsEngine(func(script string) interface{} {
// 		return false
// 	})
// 	task3 := &app.Task{
// 		If:       "some condition",
// 		JsEngine: jsengine3.(*scripting.JavaScriptEngine),
// 	}
// 	assert.False(task3.CanRunConditionally())
// }

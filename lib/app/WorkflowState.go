package app

import (
	lls "github.com/emirpasic/gods/stacks/linkedliststack"
)

type WorkflowState struct {
	CurrentTask *Task
	Stack       *lls.Stack
	History     *lls.Stack
}

type CleanupCallback = func()
type SetActiveTaskCallback = func(task *Task) CleanupCallback

// sets the current task, and pushes the previous task onto the stack if it was still running.
// returns a cleanup function callback that restores the state to its previous value.
func (ws *WorkflowState) SetCurrent(task *Task) CleanupCallback {
	if ws.CurrentTask != nil {
		ws.Stack.Push(ws.CurrentTask)
	}

	ws.CurrentTask = task

	if task == nil {
		return func() {}
	}

	ws.History.Push(task.Uuid)

	return func() {
		ws.CurrentTask = nil

		value, ok := ws.Stack.Pop()
		if ok {
			ws.CurrentTask = value.(*Task)
		}
	}
}

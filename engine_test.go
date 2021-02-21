package workflow

import (
	"fmt"
	"sync"
	"testing"
)

const (
	//sample workflow names
	cmdWF     string = "command"
	customWF  string = "custom"
	nonStopWF string = "non-stop"
	errorWF   string = "error"

	//sample workflow states
	state1     = "state1"
	state2     = "state2"
	state3     = "state3"
	state4     = "state4"
	stateError = "stateError"

	//context map test keys
	outputKey = "output"
)

//dummy work done by every step
func dummyWork(ev Event) string {
	out := ""
	if temp := ev.Context(outputKey); temp != nil {
		out = temp.(string)
	}
	out = out + ev.State() + "+"
	return out
}

//step function
var stepA = func(ev Event) {
	ev.SetContext(outputKey, dummyWork(ev))
	ev.SetState(state1)
}

//step function
var stepB = func(ev Event) {
	ev.SetContext(outputKey, dummyWork(ev))
	ev.SetState(state2)
}

//step function
var stepC = func(ev Event) {
	ev.SetContext(outputKey, dummyWork(ev))
	ev.SetState(state3)
}

//step function
var stepD = func(ev Event) {
	ev.SetContext(outputKey, dummyWork(ev))
	ev.SetState(state4)
}

//step function
var stepE = func(ev Event) {
	ev.SetContext(errKey, fmt.Errorf("failed in stepE"))
	ev.SetState(stateError)
}

//step function
var stepF = func(ev Event) {
	ev.SetContext(outputKey, dummyWork(ev))
	ev.SetState(StateStop)
}

//cmdWorkflow - returns a sample command workflow
func cmdWorkflow() (*Workflow, error) {
	actions := make(map[string]StepFunc, 4)
	actions[StateStart] = stepA
	actions[state1] = stepB
	actions[state2] = stepC
	actions[state3] = stepD
	return New(actions)
}

//customWorkflow - returns a sample custom workflow
func customWorkflow() (*Workflow, error) {
	actions := make(map[string]StepFunc, 4)
	actions[StateStart] = stepD
	actions[state4] = stepC
	actions[state3] = stepB
	actions[state2] = stepA
	return New(actions)
}

//workflowWithoutStartHandler - returns a workflow
//missing start handler
func workflowWithoutStartHandler() (*Workflow, error) {
	actions := make(map[string]StepFunc, 4)
	actions[state4] = stepC
	actions[state3] = stepB
	actions[state2] = stepA
	return New(actions)
}

//errorWorkflow - returns a workflow that runs into an error
func errorWorkflow() (*Workflow, error) {
	actions := make(map[string]StepFunc, 4)
	actions[StateStart] = stepA
	actions[state1] = stepB
	actions[state2] = stepE
	actions[state3] = stepD
	actions[stateError] = stepF
	return New(actions)
}

//a test event for cmd
type testEventCmd struct {
	BaseEvent
}

func (te *testEventCmd) ShouldStop() bool {
	//workflow should stop if the final state is reached
	//or stop state is reached
	if te.state == state4 || te.state == StateStop {
		return true
	}
	return false
}

//GetDefaultEnrichmentID returns the default enrichment id for the event type
func (te *testEventCmd) GetDefaultEnrichmentID() (string, error) {
	return "SomeID", nil
}

func newTestEventCmd(name string) *testEventCmd {
	be := BaseEvent{
		name:    name,
		state:   StateStart,
		context: make(map[string]interface{}),
	}
	return &testEventCmd{be}
}

//a test event for custom
type testEventCustom struct {
	BaseEvent
}

func (te *testEventCustom) ShouldStop() bool {
	//workflow should stop if the final state is reached
	//or stop state is reached
	if te.state == state1 || te.state == StateStop {
		return true
	}
	return false
}

//GetDefaultEnrichmentID returns the default enrichment id for the event type
func (te *testEventCustom) GetDefaultEnrichmentID() (string, error) {
	return "SomeID", nil
}

func newTestEventCustom(name string) *testEventCustom {
	be := BaseEvent{
		name:    name,
		state:   StateStart,
		context: make(map[string]interface{}),
	}
	return &testEventCustom{be}
}

//An even with incorrect stop defined
type testEventNonStop struct {
	BaseEvent
}

func (te *testEventNonStop) ShouldStop() bool {
	if te.state == "Infinite" {
		return true
	}
	return false
}

//GetDefaultEnrichmentID returns the default enrichment id for the event type
func (te *testEventNonStop) GetDefaultEnrichmentID() (string, error) {
	return "SomeID", nil
}

func newTestEventNonStop(name string) *testEventNonStop {
	be := BaseEvent{
		name:    name,
		state:   StateStart,
		context: make(map[string]interface{}),
	}
	return &testEventNonStop{be}
}

func TestRun(t *testing.T) {

	t.Run("Success: two workflows in parallel", func(t *testing.T) {
		cmd := newTestEventCmd(cmdWF)
		cmdFlow, err := cmdWorkflow()
		if err != nil {
			t.Errorf("Failed to instantiate command workflow")
			return
		}
		GetEngine().Add(cmdWF, cmdFlow)

		custom := newTestEventCustom(customWF)
		customFlow, err := customWorkflow()
		if err != nil {
			t.Errorf("Failed to instantiate custom workflow")
			return
		}
		GetEngine().Add(customWF, customFlow)

		var wg sync.WaitGroup
		wg.Add(2)

		var (
			err1, err2 error
		)

		go func() {
			err1 = GetEngine().Run(cmdWF, cmd)
			wg.Done()
		}()

		go func() {
			err2 = GetEngine().Run(customWF, custom)
			wg.Done()
		}()

		wg.Wait()

		if err1 != nil {
			t.Errorf("expected nil, got %v", err1)
			return
		}
		if err2 != nil {
			t.Errorf("expected nil, got %v", err2)
			return
		}
		expectedCmd := "STATE_START+state1+state2+state3+"
		expectedCustom := "STATE_START+state4+state3+state2+"
		if cmd.Context(outputKey).(string) != expectedCmd {
			t.Errorf("expected: %v | got: %v", expectedCmd, cmd.Context(outputKey).(string))
		}
		if custom.Context(outputKey).(string) != expectedCustom {
			t.Errorf("expected: %v | got: %v", expectedCustom, custom.Context(outputKey).(string))
		}
	})
	t.Run("A workflow with incorrect stop", func(t *testing.T) {
		ne := newTestEventNonStop(nonStopWF)
		cmdFlow, err := cmdWorkflow()
		if err != nil {
			t.Errorf("Failed to instantiate command workflow")
			return
		}
		GetEngine().Add(nonStopWF, cmdFlow)
		err = GetEngine().Run(nonStopWF, ne)
		if err == nil {
			t.Errorf("Expected error got nil")
		}
	})
	t.Run("A workflow that runs into an errors", func(t *testing.T) {
		ne := newTestEventCmd(errorWF)
		flow, err := errorWorkflow()
		if err != nil {
			t.Errorf("Failed to instantiate command workflow")
			return
		}
		GetEngine().Add(errorWF, flow)
		err = GetEngine().Run(errorWF, ne)
		if ne.Context(errKey) == nil {
			t.Errorf("Expected error key to be set")
		}
	})

	t.Run("A workflow that runs into an errors - custom error", func(t *testing.T) {
		ne := newTestEventCmd(errorWF)
		flow, err := errorWorkflow()
		if err != nil {
			t.Errorf("Failed to instantiate command workflow")
			return
		}
		ne.state = state4
		ne.SetContext("error", Error{
			Code:      BadRequestPayload,
			Message:   "error",
			Retriable: false,
		})
		GetEngine().Add(errorWF, flow)
		err = GetEngine().Run(errorWF, ne)
		if ne.Context(errKey) == nil {
			t.Errorf("Expected error key to be set")
		}
	})
}

func TestNewWorkflow(t *testing.T) {
	t.Run("A workflow with start handler", func(t *testing.T) {
		_, err := cmdWorkflow()
		if err != nil {
			t.Errorf("Expected nil got %v", err)
		}
	})
	t.Run("A workflow without start handler", func(t *testing.T) {
		_, err := workflowWithoutStartHandler()
		if err == nil {
			t.Errorf("Expected error got nil")
		}
	})
}

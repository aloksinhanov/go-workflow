package workflow

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

//States that the workflow engine needs to function
const (
	//StateStart needed to start a workflow
	StateStart string = "STATE_START"

	//StateStop needed to stop a workflow
	StateStop string = "STATE_STOP"

	//StateError needed to go to error handling step
	StateError string = "STATE_ERROR"

	//State is the key to map to the current state of the workflow
	State = "state"

	errKey = "error"
)

var (
	once   sync.Once
	engine *Engine
)

//StepFunc is a type defnition for a function
//that accpets a context map, processes and returns the context map
type StepFunc func(Event)

//Workflow is a  struct to represent a workflow mapping
type Workflow struct {
	actions map[string]StepFunc
}

//New returns an instnace of a workflow
func New(steps map[string]StepFunc) (*Workflow, error) {
	if steps[StateStart] == nil {
		return nil, errors.New("NewWorkflow: Missing mandatory mapping in the workflow for start")
	}
	return &Workflow{actions: steps}, nil
}

func (wf *Workflow) run(ev Event) error {
	for {
		if ev.ShouldStop() {
			if err := ev.Context(errKey); err != nil {
				errs, ok := err.(Error)
				if ok && !errs.Retriable {
					log.Printf(ev.GetTransactionID(), "workflow-error", fmt.Sprintf("#Notify Workflow: run | error in workflow: %v", ev.Context(errKey)))
					return nil
				}

				return fmt.Errorf("#Notify Workflow: run | error in workflow: %v", ev.Context(errKey))
			}
			break
		}
		fn, ok := wf.actions[ev.State()]
		if !ok {
			return fmt.Errorf("Workflow:run | No handler found for %v", ev.Context(State))
		}

		log.Printf(ev.GetTransactionID(), fmt.Sprintf("Worflow - current workflow context %v", ev.Context(State)))
		log.Printf(ev.GetTransactionID(), "Workflow - starting "+ev.State())
		fn(ev)
		log.Printf(ev.GetTransactionID(), "Workflow - Ending "+ev.State())
	}
	return nil
}

//Engine is a map of instantiated  workflows, where
//the key is the workflow name
type Engine struct {
	workflows map[string]*Workflow
	mutex     sync.Mutex
}

//GetEngine returns the wrapper for working with workflows
func GetEngine() *Engine {
	once.Do(func() {
		engine = &Engine{workflows: make(map[string]*Workflow)}
	})
	return engine
}

//Add - used to add a new workflow to the engine
func (e *Engine) Add(name string, wf *Workflow) {
	e.mutex.Lock()
	e.workflows[name] = wf
	e.mutex.Unlock()
}

//Run - used to run a worklflow stored with the engine
func (e *Engine) Run(name string, ev Event) error {

	wf, ok := e.workflows[name]

	if !ok {
		return fmt.Errorf("Workflow %v not found", name)
	}
	return wf.run(ev)
}

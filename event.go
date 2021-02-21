package workflow

//Event interface to be imlmented by every new workflow type
//to store the workflow state
type Event interface {
	SetContext(key string, val interface{})
	Context(key string) interface{}
	ContextMap() map[string]interface{}
	SetState(state string)
	State() string
	ShouldStop() bool
	GetDefaultEnrichmentID() (string, error)
	GetTransactionID() string
}

//BaseEvent is a struct that can be used to extend and create
//a new event type
type BaseEvent struct {
	name          string
	context       map[string]interface{}
	state         string
	transactionID string
}

func NewBaseEvent(name string, state, transactionID string) BaseEvent {
	return BaseEvent{name: name, context: make(map[string]interface{}), state: state, transactionID: transactionID}
}

//GetTransactionID - return the transactionID associate with current context
func (be *BaseEvent) GetTransactionID() string {
	return be.transactionID
}

//SetContext - sets a context in the map
func (be *BaseEvent) SetContext(key string, val interface{}) {
	be.context[key] = val
}

//Context - gets a specific context
func (be *BaseEvent) Context(key string) interface{} {
	return be.context[key]
}

//ContextMap - returns the context map for this event
func (be *BaseEvent) ContextMap() map[string]interface{} {
	return be.context
}

//SetState - sets the state for the worklow being executed
func (be *BaseEvent) SetState(state string) {
	be.state = state
}

//State - gets a specific context
func (be *BaseEvent) State() string {
	return be.state
}

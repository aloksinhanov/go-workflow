package workflow

const (
	BadRequestPayload = "BAD_REQUEST"
)

type Error struct {
	Code      string
	Message   string
	Retriable bool
}

func (e Error) String() string {
	return e.Message
}

package workflow

type WorkflowStatus int

const (
	Pending WorkflowStatus = iota
	Error
	Finished
)

type workflowState struct {
	ID     ID             `json:"id"`
	Status WorkflowStatus `json:"status"`
}

type Repository interface {
}

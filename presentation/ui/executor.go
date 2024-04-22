package ui

// An Executor takes a task and executes it eventually.
type Executor interface {
	Execute(task func())
}

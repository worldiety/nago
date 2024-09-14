// Package workflow provides some basic infrastructure components to implement a workflow.
// We define a workflow as a part of a bigger business process, which can implemented and fulfilled within
// a single technical process instance.
// Each distinct step within a workflow is called stage. A stage may be an accepted usecase or just something
// smaller due to divide-and-conquer strategies (identified and accepted after discussion by the domain experts).
package workflow

package workflow

import (
	"fmt"
	pubsub2 "go.wdy.de/nago/pkg/pubsub"
	"go.wdy.de/nago/pkg/xreflect"
	"maps"
	"slices"
	"strconv"
	"strings"
)

// A ID is a unique identifier for a workflow instance.
type ID string

// A MessageID is a unique message identifier.
type MessageID int64

// A Message represents a single message for a distinct workflow instance. Messages are replayed until committed.
// A commit takes place, as soon as a workflow stage has finished. This mechanic is similar to what Kafka does
// with its cursor per consumer.
type Message struct {
	ID               MessageID
	WorkflowInstance ID
	Value            []byte
}

type Instance struct {
	id            ID
	messages      []MessageID
	lastCommitted MessageID
}

func (i *Instance) ID() ID {
	return i.id
}

type Workflow struct {
	pubsub             *pubsub2.PubSub
	name               string
	version            int
	onDestroyObservers []func()
	stages             []stageDescriptor
}

type stageDescriptor struct {
	usecase         xreflect.TypeID
	in              xreflect.TypeID
	out             xreflect.TypeID
	anyStageAdapter func(any) any
}

func (w *Workflow) Destroy() {
	for _, fn := range w.onDestroyObservers {
		fn()
	}
}

func (w *Workflow) NewInstance() *Instance {
	return &Instance{}
}

func (w *Workflow) startNodes() []stageDescriptor {
	candidates := map[xreflect.TypeID]stageDescriptor{}
	for _, stage := range w.stages {
		candidates[stage.in] = stage
	}

	for _, stage := range w.stages {
		delete(candidates, stage.out)
	}

	return slices.Collect(maps.Values(candidates))
}

func (w *Workflow) endNodes() []stageDescriptor {
	candidates := map[xreflect.TypeID]stageDescriptor{}
	for _, stage := range w.stages {
		candidates[stage.out] = stage
	}

	for _, stage := range w.stages {
		delete(candidates, stage.in)
	}

	return slices.Collect(maps.Values(candidates))
}

func (w *Workflow) String() any {
	var sb strings.Builder
	sb.WriteString("digraph G {\n")
	sb.WriteString(fmt.Sprintf("\t_start [shape=circle;label=Start];\n"))
	sb.WriteString(fmt.Sprintf("\t_end [shape=circle;label=Ende];\n"))
	/*for _, stage := range w.stages {
		in := string(stage.in)
		if doc, ok := xreflect.TypeDoc(stage.in); ok {
			in = doc
		}

		usecase := string(stage.usecase)
		if doc, ok := xreflect.TypeDoc(stage.usecase); ok {
			usecase = doc
		}

		out := string(stage.out)
		if doc, ok := xreflect.TypeDoc(stage.out); ok {
			out = doc
		}
		writeNode(&sb, in, "Ereignis:\n"+in, "oval")
		writeNode(&sb, out, "Ereignis:\n"+out, "oval")
		writeNode(&sb, usecase, "Anwendungsfall:\n"+usecase, "rect")

		sb.WriteString(fmt.Sprintf("%s -> %s\n", in, usecase))
		sb.WriteString(fmt.Sprintf("%s -> %s\n", usecase, out))
	}*/

	/*for _, stage := range w.startNodes() {
		in := string(stage.in)
		if doc, ok := xreflect.TypeDoc(stage.in); ok {
			in = doc
		}
		sb.WriteString(fmt.Sprintf("_start -> %s\n", in))
	}

	for _, stage := range w.endNodes() {
		out := string(stage.out)
		if doc, ok := xreflect.TypeDoc(stage.out); ok {
			out = doc
		}
		sb.WriteString(fmt.Sprintf("%s -> _end\n", out))
	}*/

	sb.WriteString("}")

	return sb.String()
}

func writeNode(dst *strings.Builder, name, label, shape string) {
	dst.WriteString(fmt.Sprintf("%s [\n  label=%s;\n  shape=%s]\n", name, strconv.Quote(label), shape))
}

func Create(pubSub *pubsub2.PubSub, workflowName string, version int, configure func(wf *Workflow)) *Workflow {
	wf := &Workflow{
		pubsub:  pubSub,
		name:    workflowName,
		version: version,
	}

	configure(wf)
	return wf
}

type Stage[In, Out any] func(In) Out

func Subscribe[S ~func(In) Out, In, Out any](workflow *Workflow, stage S) {
	stageType := xreflect.TypeIDOf[S]()
	typeId := xreflect.TypeIDOf[In]()
	pubsub := workflow.pubsub
	if strings.HasPrefix(string(stageType), "func(") {
		panic(fmt.Errorf("a stage must be a concrete use case func type to gurantee correct introspection"))
	}

	unsubscribe := pubsub.Subscribe(typeId, func(value any) {
		res := stage(value.(In))
		pubsub.Publish(res)
	})
	workflow.onDestroyObservers = append(workflow.onDestroyObservers, unsubscribe)

	workflow.stages = append(workflow.stages, stageDescriptor{
		usecase: stageType,
		in:      typeId,
		out:     xreflect.TypeIDOf[Out](),
		anyStageAdapter: func(a any) any {
			return stage(a.(In))
		},
	})
}

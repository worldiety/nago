---
# Content is auto generated
# Manual changes will be overwritten!
title: Stepper
---
It visually represents a sequence of steps in a process, highlighting
completed, current, and upcoming steps with distinct colors and styles. Each step can display a label, and the layout adapts between simple
or full-sized step representations.

## Constructors
### Stepper
Stepper creates a new stepper with the given steps

---
## Methods
| Method | Description |
|--------| ------------|
| `CompletedTextPattern(pattern string)` | CompletedTextPattern overwrites the default text pattern for completed simple steppers. Use %d for the total number of steps |
| `InputValue(state *core.State[int])` | InputValue sets a step index state, that will be used instead of the fixed value of the component |
| `Layout(layout StepperLayout)` | Layout sets a fixed layout for the stepper |
| `Lines(b bool)` | Lines defines whether to display lines in the stepper with the simple or simple list layout |
| `Numbers(b bool)` | Numbers defines whether to display step numbers in the stepper |
| `SimpleTextPattern(pattern string)` | SimpleTextPattern overwrites the default text pattern for the simple stepper layout. Use %d for the current step and another %d for the total number of steps |
| `Steps(steps ...)` | Steps sets the steps of the stepper. Previously set steps will be overwritten |
| `Value(value int)` | Value sets the current step index value |
---


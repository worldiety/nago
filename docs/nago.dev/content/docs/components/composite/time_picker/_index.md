---
# Content is auto generated
# Manual changes will be overwritten!
title: Time Picker
---
It lets users choose a duration with optional granularity for days,
hours, minutes, and seconds. The picker can be shown in a dialog,
bind to external state, and format its value either as a clock or
as decomposed units.

## Constructors
### Picker
Picker renders a time.Duration either in clock time format or in decomposed format.
Default is [ClockFormat]. By default, the Picker shows hours and minutes,
but you can be specific by setting the according flags.
Keep in mind, that the picker also clamps to the natural limits, e.g. you cannot set
25 hours, instead you must enable the day flag, so that the user can configure 1 day and 1 hour.

---
## Methods
| Method | Description |
|--------| ------------|
| `AccessibilityLabel(label string)` | AccessibilityLabel sets a label used by screen readers for accessibility. (currently not implemented) |
| `Border(border ui.Border)` | Border sets the border style of the time picker. (currently not implemented) |
| `Days(showDays bool)` | Days toggles whether the picker allows selecting days. |
| `Disabled(disabled bool)` | Disabled enables or disables user interaction with the time picker. |
| `ErrorText(text string)` | ErrorText sets the validation or error message displayed below the picker. |
| `Format(format PickerFormat)` | Format sets the display format for the duration value (clock or decomposed). |
| `Frame(frame ui.Frame)` | Frame sets the layout frame of the time picker, including size and positioning. |
| `Hours(showHours bool)` | Hours toggles whether the picker allows selecting hours. |
| `Minutes(showMinutes bool)` | Minutes toggles whether the picker allows selecting minutes. |
| `Padding(padding ui.Padding)` | Padding sets the inner spacing around the time picker content. (currently not implemented) |
| `Seconds(showSeconds bool)` | Seconds toggles whether the picker allows selecting seconds. |
| `SupportingText(text string)` | SupportingText sets helper or secondary text displayed below the picker label. |
| `Title(title string)` | Title sets the title of the picker, typically shown in dialogs. |
| `Visible(visible bool)` | Visible controls the visibility of the time picker; setting false hides it. (currently not implemented) |
| `WithFrame(fn func(ui.Frame) ui.Frame)` | WithFrame applies a transformation function to the picker's frame and returns the updated component. |
| `dayDown()` | dayDown decreases the number of days in the current selection, wrapping around to 99 if it goes below 0. |
| `dayUp()` | dayUp increases the number of days in the current selection, wrapping back to 0 if it exceeds 99. |
| `hourDown()` | hourDown decreases the hours in the current selection, wrapping around to 23 if it goes below 0. |
| `hourUp()` | hourUp increases the hours in the current selection, wrapping back to 0 if it reaches 24. |
| `minDown()` | minDown decreases the minutes in the current selection, wrapping around to 59 if it goes below 0. |
| `minUp()` | minUp increases the minutes in the current selection, wrapping back to 0 if it exceeds 59. |
| `renderPicker()` | renderPicker builds the interactive picker view for adjusting the duration. It shows increment and decrement buttons with numeric labels for each enabled unit (days, hours, minutes, seconds). If no units are explicitly enabled, the picker automatically decides which units to display based on the current duration. |
| `round(d time.Duration)` | round normalizes the given duration based on the picker's configuration. If seconds are not displayed, the duration is truncated to the nearest minute; otherwise, it is returned unchanged. |
| `secDown()` | secDown decreases the seconds in the current selection, wrapping around to 59 if it goes below 0. |
| `secUp()` | secUp increases the seconds in the current selection, wrapping back to 0 if it exceeds 59. |
---


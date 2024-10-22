package timepicker

import (
	"fmt"
	"go.wdy.de/nago/presentation/core"
	heroSolid "go.wdy.de/nago/presentation/icons/hero/solid"
	"go.wdy.de/nago/presentation/ui"
	"go.wdy.de/nago/presentation/ui/alert"
	"strings"
	"time"
)

type PickerFormat int

const (
	// ClockFormat interprets a duration as a clock e.g. 18:30. Note, that days and seconds
	// are also rendered the same way, so you may get also something like 01:18:30:42 to display.
	ClockFormat PickerFormat = iota

	// DecomposedFormat interprets a duration as distinct days, hours, minutes and seconds e.g.
	// 18 Std. 30 Min. At worst, you may get 1 T 18 Std. 30 Min. 42 Sek to display.
	DecomposedFormat
)

type TPicker struct {
	label                string
	supportingText       string
	errorText            string
	frame                ui.Frame
	pickerPresented      *core.State[bool]
	targetSelectedState  *core.State[time.Duration]
	currentSelectedState *core.State[time.Duration]
	title                string
	showDays             bool
	showHours            bool
	showMinutes          bool
	showSeconds          bool
	format               PickerFormat
	disabled             bool
	// why is the scale-to-seconds removed? because it limits the picker to be in seconds, however we also need at least millisecond resolution in the near future.
}

// Picker renders a time.Duration either in clock time format or in decomposed format.
// Default is [ClockFormat]. By default, the Picker shows hours and minutes,
// but you can be specific by setting the according flags.
// Keep in mind, that the picker also clamps to the natural limits, e.g. you cannot set
// 25 hours, instead you must enable the day flag, so that the user can configure 1 day and 1 hour.
func Picker(label string, selectedState *core.State[time.Duration]) TPicker {
	p := TPicker{
		label:               label,
		format:              ClockFormat,
		targetSelectedState: selectedState,
	}

	if selectedState != nil {
		p.pickerPresented = core.DerivedState[bool](selectedState, ".pck.pre")
		p.currentSelectedState = core.DerivedState[time.Duration](selectedState, ".pck.tmp").Init(func() time.Duration {
			return selectedState.Get()
		})

		p.currentSelectedState.Observe(func(newValue time.Duration) {
			p.targetSelectedState.Set(newValue)
		})
	}
	return p
}

func (c TPicker) Padding(padding ui.Padding) ui.DecoredView {
	//TODO implement me
	return c
}

func (c TPicker) Frame(frame ui.Frame) ui.DecoredView {
	c.frame = frame
	return c
}

func (c TPicker) Border(border ui.Border) ui.DecoredView {
	//TODO implement me
	return c
}

func (c TPicker) Visible(visible bool) ui.DecoredView {
	//TODO implement me
	return c
}

func (c TPicker) AccessibilityLabel(label string) ui.DecoredView {
	//TODO implement me
	return c
}

func (c TPicker) Disabled(disabled bool) TPicker {
	c.disabled = disabled
	return c
}

func (c TPicker) Title(title string) TPicker {
	c.title = title
	return c
}

func (c TPicker) Format(format PickerFormat) TPicker {
	c.format = format
	return c
}

func (c TPicker) SupportingText(text string) TPicker {
	c.supportingText = text
	return c
}

func (c TPicker) ErrorText(text string) TPicker {
	c.errorText = text
	return c
}

func (c TPicker) Hours(showHours bool) TPicker {
	c.showHours = showHours
	return c
}

func (c TPicker) Minutes(showMinutes bool) TPicker {
	c.showMinutes = showMinutes
	return c
}

func (c TPicker) Days(showDays bool) TPicker {
	c.showDays = showDays
	return c
}

func (c TPicker) Seconds(showSeconds bool) TPicker {
	c.showSeconds = showSeconds
	return c
}

func auto(showDays, showHours, showMinutes, showSeconds bool) bool {
	return !showDays && !showHours && !showMinutes && !showSeconds
}

func fmtDurationTime(showDays, showHours, showMinutes, showSeconds bool, d time.Duration) string {
	days, hours, minutes, seconds := FromDuration(d)
	if auto(showDays, showHours, showMinutes, showSeconds) {
		showDays = days != 0
		showHours = hours != 0
		showMinutes = minutes != 0
		showSeconds = seconds != 0
	}

	if auto(showDays, showHours, showMinutes, showSeconds) {
		showHours = true
		showMinutes = true
	}

	var segments []string
	if showDays {
		segments = append(segments, fmt.Sprintf("%d T", days))
	}

	if showHours {
		segments = append(segments, fmt.Sprintf("%d Std.", hours))
	}

	if showMinutes {
		segments = append(segments, fmt.Sprintf("%d Min.", minutes))
	}

	if showSeconds {
		segments = append(segments, fmt.Sprintf("%d Sek.", seconds))
	}

	return strings.Join(segments, " ")
}

func fmtClockTime(showDays, showHours, showMinutes, showSeconds bool, d time.Duration) string {
	days, hours, minutes, seconds := FromDuration(d)

	if auto(showDays, showHours, showMinutes, showSeconds) {
		showDays = days != 0
		showHours = hours != 0
		showMinutes = minutes != 0
		showSeconds = seconds != 0
	}

	if auto(showDays, showHours, showMinutes, showSeconds) {
		showHours = true
		showMinutes = true
	}

	var segments []string
	if showDays {
		segments = append(segments, fmt.Sprintf("%02d", days))
	}

	if showHours {
		segments = append(segments, fmt.Sprintf("%02d", hours))
	}

	if showMinutes {
		segments = append(segments, fmt.Sprintf("%02d", minutes))
	}

	if showSeconds {
		segments = append(segments, fmt.Sprintf("%02d", seconds))
	}

	return strings.Join(segments, "\u202F:\u202F") // use thin space instead of space
}

func (c TPicker) dayDown() {
	days, hours, minutes, seconds := FromDuration(c.currentSelectedState.Get())
	days--
	if days < 0 {
		days = 99
	}
	c.currentSelectedState.Set(Duration(days, hours, minutes, seconds))
}

func (c TPicker) dayUp() {
	days, hours, minutes, seconds := FromDuration(c.currentSelectedState.Get())
	days++
	if days > 99 {
		days = 0
	}
	c.currentSelectedState.Set(Duration(days, hours, minutes, seconds))
}

func (c TPicker) hourDown() {
	days, hours, minutes, seconds := FromDuration(c.currentSelectedState.Get())
	hours--
	if hours < 0 {
		hours = 23
	}
	c.currentSelectedState.Set(Duration(days, hours, minutes, seconds))
}

func (c TPicker) hourUp() {
	days, hours, minutes, seconds := FromDuration(c.currentSelectedState.Get())
	hours++
	if hours >= 24 {
		hours = 0
	}
	c.currentSelectedState.Set(Duration(days, hours, minutes, seconds))
}

func (c TPicker) minDown() {
	days, hours, minutes, seconds := FromDuration(c.currentSelectedState.Get())
	minutes--
	if minutes < 0 {
		minutes = 59
	}
	c.currentSelectedState.Set(Duration(days, hours, minutes, seconds))
}

func (c TPicker) minUp() {
	days, hours, minutes, seconds := FromDuration(c.currentSelectedState.Get())
	minutes++
	if minutes > 59 {
		minutes = 0
	}
	c.currentSelectedState.Set(Duration(days, hours, minutes, seconds))
}

func (c TPicker) secDown() {
	days, hours, minutes, seconds := FromDuration(c.currentSelectedState.Get())
	seconds--
	if seconds < 0 {
		seconds = 59
	}
	c.currentSelectedState.Set(Duration(days, hours, minutes, seconds))
}

func (c TPicker) secUp() {
	days, hours, minutes, seconds := FromDuration(c.currentSelectedState.Get())
	seconds++
	if seconds > 59 {
		seconds = 0
	}
	c.currentSelectedState.Set(Duration(days, hours, minutes, seconds))
}

func (c TPicker) renderPicker() core.View {
	days, hours, minutes, seconds := FromDuration(c.currentSelectedState.Get())

	if auto(c.showDays, c.showHours, c.showMinutes, c.showSeconds) {
		c.showDays = days != 0
		c.showHours = hours != 0
		c.showMinutes = minutes != 0
		c.showSeconds = seconds != 0
	}

	if auto(c.showDays, c.showHours, c.showMinutes, c.showSeconds) {
		c.showHours = true
		c.showMinutes = true
	}

	segments := make([]core.View, 0, 8)
	if c.showDays {
		segments = append(segments,
			ui.VStack(
				ui.TertiaryButton(c.dayUp).PreIcon(heroSolid.ChevronUp),
				ui.Text(fmt.Sprintf("%02d", days)),
				ui.TertiaryButton(c.dayDown).PreIcon(heroSolid.ChevronDown),
			),
			ui.Text("Tage"),
		)
	}

	if c.showHours {
		segments = append(segments,
			ui.VStack(
				ui.TertiaryButton(c.hourUp).PreIcon(heroSolid.ChevronUp),
				ui.Text(fmt.Sprintf("%02d", hours)),
				ui.TertiaryButton(c.hourDown).PreIcon(heroSolid.ChevronDown),
			),
			ui.Text("Std."),
		)
	}

	if c.showMinutes {
		segments = append(segments,
			ui.VStack(
				ui.TertiaryButton(c.minUp).PreIcon(heroSolid.ChevronUp),
				ui.Text(fmt.Sprintf("%02d", minutes)),
				ui.TertiaryButton(c.minDown).PreIcon(heroSolid.ChevronDown),
			),
			ui.Text("Min."),
		)
	}

	if c.showSeconds {
		segments = append(segments,
			ui.VStack(
				ui.TertiaryButton(c.secUp).PreIcon(heroSolid.ChevronUp),
				ui.Text(fmt.Sprintf("%02d", seconds)),
				ui.TertiaryButton(c.secDown).PreIcon(heroSolid.ChevronDown),
			),
			ui.Text("Sek."),
		)
	}

	return ui.HStack(segments...).Frame(ui.Frame{}.FullWidth())
}

func (c TPicker) Render(ctx core.RenderContext) core.RenderNode {
	durationText := Format(c.showDays, c.showHours, c.showMinutes, c.showSeconds, c.format, c.currentSelectedState.Get())

	colors := core.Colors[ui.Colors](ctx.Window())
	inner := ui.HStack(
		alert.Dialog(c.title, c.renderPicker(), c.pickerPresented, alert.Cancel(func() {
			c.currentSelectedState.Set(c.targetSelectedState.Get())
		}), alert.Custom(func(close func(closeDlg bool)) core.View {
			// positive case
			return ui.PrimaryButton(func() {
				c.targetSelectedState.Set(c.currentSelectedState.Get())
				c.targetSelectedState.Notify() // invoke observers
				close(true)
			}).Title(fmt.Sprintf("übernehmen"))
		})),
		ui.Text(durationText),
		ui.Spacer(),
		ui.Image().Embed(heroSolid.ChevronDown).Frame(ui.Frame{}.Size(ui.L16, ui.L16)),
	).Action(func() {
		if c.disabled {
			return
		}
		c.pickerPresented.Set(true)
	}).HoveredBorder(ui.Border{}.Color(colors.I1.WithBrightness(75)).Width(ui.L1).Radius("0.375rem")).
		Gap(ui.L8).
		Frame(ui.Frame{}.FullWidth()).
		Border(ui.Border{}.Color(ui.M8).Width(ui.L1).Radius("0.375rem")).
		Padding(ui.Padding{}.All(ui.L8))

	return ui.VStack(
		ui.IfElse(c.errorText == "",
			ui.Text(c.label).Font(ui.Font{Size: ui.L16}),
			ui.HStack(
				ui.Image().StrokeColor(ui.SE0).Embed(heroSolid.XMark).Frame(ui.Frame{}.Size(ui.L20, ui.L20)),
				ui.Text(c.label).Font(ui.Font{Size: ui.L16}).Color(ui.SE0),
			),
		),
		inner,
		ui.IfElse(c.errorText == "",
			ui.Text(c.supportingText).Font(ui.Font{Size: "0.75rem"}).Color(ui.ST0),
			ui.Text(c.errorText).Font(ui.Font{Size: "0.75rem"}).Color(ui.SE0),
		),
	).Alignment(ui.Leading).
		Gap(ui.L4).
		Frame(c.frame).
		Render(ctx)
}

// FromDuration returns the days, hours, minutes and seconds from the given duration.
func FromDuration(d time.Duration) (days, hours, minutes, seconds int) {
	days = int(d / (time.Hour * 24))
	d = d % (time.Hour * 24)

	hours = int(d / time.Hour)
	d = d % time.Hour

	minutes = int(d / time.Minute)
	d = d % time.Minute

	seconds = int(d / time.Second)

	return
}

// Duration creates time.Duration based on the decomposed individual durations.
func Duration(days, hours, minutes, seconds int) time.Duration {
	return time.Hour*24*time.Duration(days) + time.Hour*time.Duration(hours) + time.Minute*time.Duration(minutes) + time.Second*time.Duration(seconds)
}

func Format(showDays, showHours, showMinutes, showSeconds bool, format PickerFormat, duration time.Duration) string {
	dur := duration

	var durationText string
	switch format {
	case DecomposedFormat:
		durationText = fmtDurationTime(showDays, showHours, showMinutes, showSeconds, dur)
	default:
		durationText = fmtClockTime(showDays, showHours, showMinutes, showSeconds, dur)
	}

	return durationText
}

// Minutes returns only the exact truncated minutes fraction of the duration.
func Minutes(duration time.Duration) int {
	_, _, m, _ := FromDuration(duration)
	return m
}

// Hours returns only the exact truncated hours fraction of the duration.
func Hours(duration time.Duration) int {
	_, h, _, _ := FromDuration(duration)
	return h
}

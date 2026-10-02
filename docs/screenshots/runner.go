// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

//go:embed page.js
var pageJS string

type runner struct {
	root     string
	site     string
	manifest Manifest
	headful  bool
}

func (r *runner) renderExample(ctx context.Context, example string, shots []Shot) error {
	tmp, err := os.MkdirTemp("", "nago-shot-"+example+"-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	log.Printf("%s: building", example)
	bin := filepath.Join(tmp, "app")
	build := exec.CommandContext(ctx, "go", "build", "-o", bin, "./example/cmd/"+example)
	build.Dir = r.root
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build failed: %w\n%s", err, out)
	}

	port, err := freePort()
	if err != nil {
		return err
	}

	stateDir := filepath.Join(tmp, "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		return err
	}

	logFile, err := os.Create(filepath.Join(tmp, "app.log"))
	if err != nil {
		return err
	}
	defer logFile.Close()

	app := exec.CommandContext(ctx, bin)
	app.Dir = tmp
	app.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", port), "STATE_DIRECTORY="+stateDir)
	app.Stdout = logFile
	app.Stderr = logFile
	if err := app.Start(); err != nil {
		return err
	}
	defer stop(app)

	base := fmt.Sprintf("http://localhost:%d", port)
	if err := waitHTTP(ctx, base, 60*time.Second); err != nil {
		return r.withLog(err, logFile)
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", !r.headful),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("force-color-profile", "srgb"),
		chromedp.Flag("font-render-hinting", "none"),
		// the frontend does not boot for crawlers and the default headless user agent is detected as one
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36"),
	)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	defer cancelAlloc()

	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	defer cancelBrowser()

	for _, shot := range shots {
		log.Printf("%s: rendering %s", example, shot.Out)
		if err := r.renderShot(browserCtx, base, shot); err != nil {
			return r.withLog(fmt.Errorf("%s: %w", shot.Out, err), logFile)
		}
	}

	return nil
}

func (r *runner) renderShot(browserCtx context.Context, base string, shot Shot) error {
	tabCtx, cancelTab := chromedp.NewContext(browserCtx)
	defer cancelTab()

	ctx, cancel := context.WithTimeout(tabCtx, 60*time.Second)
	defer cancel()

	scheme := "light"
	if shot.Dark != nil && *shot.Dark {
		scheme = "dark"
	}

	err := chromedp.Run(ctx,
		emulation.SetDeviceMetricsOverride(int64(shot.Width), int64(shot.Height), shot.Scale, false),
		emulation.SetEmulatedMedia().WithFeatures([]*emulation.MediaFeature{
			{Name: "prefers-color-scheme", Value: scheme},
			{Name: "prefers-reduced-motion", Value: "reduce"},
		}),
		chromedp.Navigate(base+shot.Path),
		chromedp.Evaluate(pageJS, nil),
	)
	if err != nil {
		return err
	}

	if err := settle(ctx); err != nil {
		return err
	}

	for _, step := range shot.Steps {
		if err := runStep(ctx, base, step); err != nil {
			return err
		}

		if err := settle(ctx); err != nil {
			return err
		}
	}

	clip, err := cropRect(ctx, shot)
	if err != nil {
		return err
	}

	format := page.CaptureScreenshotFormatPng
	if strings.EqualFold(filepath.Ext(shot.Out), ".webp") {
		format = page.CaptureScreenshotFormatWebp
	}

	var buf []byte
	err = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		cap := page.CaptureScreenshot().
			WithFormat(format).
			WithCaptureBeyondViewport(true).
			WithClip(&page.Viewport{X: clip.X, Y: clip.Y, Width: clip.W, Height: clip.H, Scale: 1})
		if format != page.CaptureScreenshotFormatPng {
			cap = cap.WithQuality(int64(shot.Quality))
		}

		var err error
		buf, err = cap.Do(ctx)
		return err
	}))
	if err != nil {
		return err
	}

	dst := filepath.Join(r.site, shot.Out)
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	return os.WriteFile(dst, buf, 0644)
}

type rect struct {
	X, Y, W, H float64
}

// cropRect asks page.js for the region to capture. Padding is applied here and the result is clamped
// to the document, so that a crop never shows anything outside of the page.
func cropRect(ctx context.Context, shot Shot) (rect, error) {
	var res struct {
		Rect rect    `json:"rect"`
		DocW float64 `json:"docW"`
		DocH float64 `json:"docH"`
	}

	expr := fmt.Sprintf("window.__nagoShot.crop(%q)", shot.Crop)
	if err := chromedp.Run(ctx, chromedp.Evaluate(expr, &res)); err != nil {
		return rect{}, err
	}

	if res.Rect.W <= 0 || res.Rect.H <= 0 {
		return rect{}, fmt.Errorf("crop %q matched nothing visible", shot.Crop)
	}

	r := res.Rect
	if shot.Crop != "viewport" && shot.Crop != "page" {
		p := float64(*shot.Padding)
		r = rect{X: r.X - p, Y: r.Y - p, W: r.W + 2*p, H: r.H + 2*p}
	}

	x0 := math.Max(0, math.Floor(r.X))
	y0 := math.Max(0, math.Floor(r.Y))
	x1 := math.Min(res.DocW, math.Ceil(r.X+r.W))
	y1 := math.Min(res.DocH, math.Ceil(r.Y+r.H))

	return rect{X: x0, Y: y0, W: x1 - x0, H: y1 - y0}, nil
}

func runStep(ctx context.Context, base string, step Step) error {
	switch {
	case step.Click != "":
		return clickAt(ctx, fmt.Sprintf("window.__nagoShot.centerOf(%q)", step.Click), step.Click)
	case step.ClickText != "":
		return clickAt(ctx, fmt.Sprintf("window.__nagoShot.centerOfText(%q)", step.ClickText), step.ClickText)
	case step.Hover != "":
		pt, err := pointOf(ctx, fmt.Sprintf("window.__nagoShot.centerOf(%q)", step.Hover), step.Hover)
		if err != nil {
			return err
		}

		return chromedp.Run(ctx, input.DispatchMouseEvent(input.MouseMoved, pt.X, pt.Y))
	case step.Type != nil:
		return chromedp.Run(ctx,
			chromedp.Click(step.Type.Selector, chromedp.ByQuery),
			chromedp.SendKeys(step.Type.Selector, step.Type.Text, chromedp.ByQuery),
		)
	case step.Wait != "":
		d, _ := time.ParseDuration(step.Wait) // validated by loadManifest
		return chromedp.Run(ctx, chromedp.Sleep(d))
	case step.JS != "":
		return chromedp.Run(ctx, chromedp.Evaluate(step.JS, nil))
	case step.Goto != "":
		return chromedp.Run(ctx, chromedp.Navigate(base+step.Goto), chromedp.Evaluate(pageJS, nil))
	default:
		return fmt.Errorf("empty step")
	}
}

type point struct {
	X, Y float64
	OK   bool
}

func pointOf(ctx context.Context, expr, what string) (point, error) {
	var pt point
	if err := chromedp.Run(ctx, chromedp.Evaluate(expr, &pt)); err != nil {
		return pt, err
	}

	if !pt.OK {
		return pt, fmt.Errorf("nothing visible found for %q", what)
	}

	return pt, nil
}

// clickAt dispatches real mouse events, because the nago frontend reacts on pointer events and not only on
// synthetic click events.
func clickAt(ctx context.Context, expr, what string) error {
	pt, err := pointOf(ctx, expr, what)
	if err != nil {
		return err
	}

	return chromedp.Run(ctx, chromedp.MouseClickXY(pt.X, pt.Y))
}

// settle waits until the frontend has rendered, all fonts and images are loaded and the DOM did not change
// for a moment, so that animations and late server updates are part of the capture.
func settle(ctx context.Context) error {
	var ok bool
	if err := chromedp.Run(ctx, chromedp.Evaluate("window.__nagoShot.settle()", &ok, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
		return p.WithAwaitPromise(true)
	})); err != nil {
		return err
	}

	if !ok {
		var body string
		_ = chromedp.Run(ctx, chromedp.Evaluate("document.body.innerHTML.slice(0, 2000)", &body))
		return fmt.Errorf("page did not render any content:\n%s", body)
	}

	return nil
}

func (r *runner) withLog(err error, f *os.File) error {
	buf, _ := os.ReadFile(f.Name())
	lines := strings.Split(strings.TrimSpace(string(buf)), "\n")
	if len(lines) > 20 {
		lines = lines[len(lines)-20:]
	}

	return fmt.Errorf("%w\n--- last lines of the app log:\n%s", err, strings.Join(lines, "\n"))
}

func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}

func waitHTTP(ctx context.Context, url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if res, err := http.DefaultClient.Do(req); err == nil {
			res.Body.Close()
			if res.StatusCode == http.StatusOK {
				return nil
			}
		}

		time.Sleep(250 * time.Millisecond)
	}

	return fmt.Errorf("app did not answer on %s within %v", url, timeout)
}

func stop(cmd *exec.Cmd) {
	_ = cmd.Process.Signal(syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
		<-done
	}
}

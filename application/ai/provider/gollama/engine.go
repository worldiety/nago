// Copyright (c) 2026 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: Custom-License

package gollama

import (
	"fmt"
	"strings"
	"sync"

	gollama "github.com/dianlight/gollama.cpp"
	"go.wdy.de/nago/application/ai/completion"
)

const (
	// defaultMaxTokens caps generated tokens when the request does not specify a limit.
	defaultMaxTokens = 2048
	// defaultCtxSize is the context window used when neither the catalog entry nor the provider settings
	// configure one. It is intentionally conservative to bound memory use on local machines.
	defaultCtxSize = 4096
)

// engine owns the llama.cpp lifecycle: it initialises the backend exactly once, keeps a cache of loaded
// (memory-mapped) models keyed by file path, and runs the greedy generation loop. Models are expensive to
// load and immutable once loaded, so they are shared across requests; a fresh context (and therefore a fresh
// KV cache) is created per request so concurrent completions never share decoder state.
type engine struct {
	cfg Settings

	initOnce sync.Once
	initErr  error
	sampler  gollama.LlamaSampler

	mu     sync.Mutex // guards models and serialises model loading
	models map[string]*loadedModel

	dlMu sync.Mutex // serialises downloads so two requests never fetch the same file twice
}

// loadedModel is a model held in memory together with the metadata-derived family adapter.
type loadedModel struct {
	path    string
	handle  gollama.LlamaModel
	meta    ggufMetadata
	adapter adapter
}

func newEngine(cfg Settings) *engine {
	return &engine{cfg: cfg, models: map[string]*loadedModel{}}
}

// ensureBackend initialises the llama.cpp backend and the shared greedy sampler exactly once. The greedy
// sampler is stateless and therefore safe to reuse across concurrent generations.
//
// Before initialising the backend we explicitly load the llama.cpp shared library that ships embedded in the
// gollama.cpp module (build b6862). This matters because the Go binding's struct layouts (notably
// llama_context_params) are generated for that exact build: loading a different llama.cpp build leads to an
// ABI mismatch and an intermittent SIGSEGV inside llama_init_from_model. Backend_init on its own does NOT
// extract the embedded library; it only scans the cache directory and otherwise falls back to a bare
// dlopen("libllama.dylib"), which resolves to whatever the system dynamic loader finds first (e.g. a
// Homebrew-installed llama.cpp of a different, incompatible build). LoadLibraryWithVersion("") extracts the
// embedded build into the same cache directory that Backend_init subsequently scans, guaranteeing the binding
// loads the library it was compiled against. We surface a load failure as an error rather than letting
// Backend_init silently fall back to an ABI-incompatible system library.
func (e *engine) ensureBackend() error {
	e.initOnce.Do(func() {
		if err := gollama.LoadLibraryWithVersion(""); err != nil {
			e.initErr = fmt.Errorf("load embedded llama.cpp library: %w", err)
			return
		}
		if err := gollama.Backend_init(); err != nil {
			e.initErr = fmt.Errorf("llama.cpp backend init: %w", err)
			return
		}
		e.sampler = gollama.Sampler_init_greedy()
	})
	return e.initErr
}

// load resolves the model file for a catalog entry (downloading it if necessary), loads it into memory once
// and returns the cached handle.
func (e *engine) load(entry catalogEntry) (*loadedModel, error) {
	if err := e.ensureBackend(); err != nil {
		return nil, err
	}

	path, err := e.resolveModel(entry)
	if err != nil {
		return nil, err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if lm, ok := e.models[path]; ok {
		return lm, nil
	}

	meta, err := readGGUFMetadata(path)
	if err != nil {
		return nil, fmt.Errorf("read gguf metadata %q: %w", path, err)
	}

	mp := gollama.Model_default_params()
	if e.cfg.GpuLayers > 0 {
		mp.NGpuLayers = int32(e.cfg.GpuLayers)
	}

	handle, err := gollama.Model_load_from_file(path, mp)
	if err != nil {
		return nil, fmt.Errorf("load model %q: %w", path, err)
	}

	lm := &loadedModel{
		path:    path,
		handle:  handle,
		meta:    meta,
		adapter: selectAdapter(entry.Family, meta),
	}
	e.models[path] = lm
	return lm, nil
}

// effectiveCtx resolves the context window (in tokens) the llama context is created with. It is both the
// physical KV-cache size (passed as n_ctx, see generate) and the logical bound used by the prompt-overflow
// pre-check and the decode-loop limit.
//
// The window is the configured size capped by the model's trained context length: a model cannot be used
// beyond what it was trained for, so any larger request is clamped to the model maximum. The configured size
// is the catalog entry's CtxSize, else the provider's CtxSize, else [defaultCtxSize]. A configured value of 0
// therefore means "use the built-in default", NOT "use the model's full context": local models routinely
// advertise very large context lengths (Qwen3 reports 40960) whose KV cache would consume several gigabytes,
// so we default conservatively and let operators opt into a larger window explicitly. When the model does not
// advertise a context length we fall back to the configured size (or the default).
func (e *engine) effectiveCtx(entry catalogEntry, meta ggufMetadata) int {
	physical := meta.contextLength

	want := defaultCtxSize
	if entry.CtxSize > 0 {
		want = entry.CtxSize
	} else if e.cfg.CtxSize > 0 {
		want = e.cfg.CtxSize
	}

	if physical > 0 && physical < want {
		return physical
	}
	return want
}

// genOutput is the raw result of one generation pass, before the family adapter parses the text into content
// blocks.
type genOutput struct {
	text         string
	stop         completion.StopReason
	promptTokens int
	outputTokens int
	// emitted is the number of bytes of text already surfaced through the emit callback. The completion layer
	// uses it to flush any text that was held back by tool-call masking but turned out not to be a tool call.
	emitted int
}

// generate runs the greedy decode loop for a single prompt. emit, when non-nil, is invoked with incremental
// text segments as they become safe to surface (stop strings and tool-call markers are held back); it returns
// false to abort generation early. The returned text is the full assistant output with any trailing stop
// string removed.
func (e *engine) generate(lm *loadedModel, prompt string, stops, toolMarkers []string, maxTokens, nCtx int, emit func(string) bool) (genOutput, error) {
	promptTokens, err := gollama.Tokenize(lm.handle, prompt, true, true)
	if err != nil {
		return genOutput{}, fmt.Errorf("tokenize prompt: %w", err)
	}
	if len(promptTokens) == 0 {
		return genOutput{}, fmt.Errorf("empty prompt after tokenization")
	}
	if nCtx > 0 && len(promptTokens) >= nCtx {
		return genOutput{}, completion.ContextWindowError{Limit: nCtx, Tokens: len(promptTokens)}
	}

	// The gollama.cpp b6862 binding ships a buggy llama_context_params definition: both the Go
	// LlamaContextParams struct and its libffi type descriptor carry a leading `seed uint32` field that
	// upstream llama.cpp removed years ago. The struct and the descriptor agree with each other, so the
	// binding round-trips its own defaults correctly, but relative to the real C layout every field is shifted
	// by one uint32. This is verified empirically against the loaded dylib: writing the Go `Seed` field changes
	// the context's C `n_ctx` (and the KV-cache size) accordingly. We therefore configure each parameter
	// through the Go field that sits one position BEFORE the intended C field and leave everything else at the
	// (consistently shifted) defaults returned by Context_default_params:
	//
	//	C n_ctx           <- Go Seed
	//	C n_threads       <- Go NSeqMax
	//	C n_threads_batch <- Go NThreads
	//
	// Using a field's natural Go name would instead write the NEXT C field (e.g. Go NCtx -> C n_batch), so the
	// shifted names below are deliberate. The phantom-seed bug is also why we cannot simply set NCtx; doing so
	// leaves n_ctx at its 512 default and silently truncates longer conversations.
	cp := gollama.Context_default_params()
	if nCtx > 0 {
		cp.Seed = uint32(nCtx) // -> C n_ctx
	}
	if e.cfg.Threads > 0 {
		cp.NSeqMax = uint32(e.cfg.Threads) // -> C n_threads
		cp.NThreads = int32(e.cfg.Threads) // -> C n_threads_batch
	}

	cctx, err := gollama.Init_from_model(lm.handle, cp)
	if err != nil {
		return genOutput{}, fmt.Errorf("create context: %w", err)
	}
	defer gollama.Free(cctx)

	if err := gollama.Decode(cctx, gollama.Batch_get_one(promptTokens)); err != nil {
		return genOutput{}, fmt.Errorf("decode prompt: %w", err)
	}

	limit := maxTokens
	if limit <= 0 {
		limit = defaultMaxTokens
	}

	dec := newPieceDecoder(lm.meta)
	gate := newStreamGate(stops, toolMarkers, emit)
	stop := completion.StopMaxTokens
	nPast := len(promptTokens)

	var produced int
	for produced = 0; produced < limit; produced++ {
		tok := gollama.Sampler_sample(e.sampler, cctx, -1)
		if tok == gollama.LLAMA_TOKEN_NULL {
			stop = completion.StopEndTurn
			break
		}
		if lm.meta.isEOG(int(tok)) {
			stop = completion.StopEndTurn
			break
		}

		gate.push(dec.decode(gollama.Token_to_piece(lm.handle, tok, true)))
		if gate.aborted {
			break
		}
		if gate.stopped {
			stop = completion.StopStopSequence
			break
		}

		if err := gollama.Decode(cctx, gollama.Batch_get_one([]gollama.LlamaToken{tok})); err != nil {
			return genOutput{}, fmt.Errorf("decode token: %w", err)
		}
		nPast++
		if nPast >= nCtx {
			stop = completion.StopMaxTokens
			break
		}
	}

	gate.push(dec.flush())
	text := gate.finish()

	return genOutput{
		text:         text,
		stop:         stop,
		promptTokens: len(promptTokens),
		outputTokens: produced,
		emitted:      gate.emitted,
	}, nil
}

// streamGate buffers decoded text so that (a) stop strings can trim the output and never be surfaced and
// (b) tool-call markers suppress incremental text emission (the raw tool-call syntax must not appear as
// visible assistant text). It emits only the portion of text that cannot be part of a pending marker, holding
// back a window equal to the longest marker minus one byte.
type streamGate struct {
	stops   []string
	markers []string
	maxHold int
	emit    func(string) bool
	full    string
	emitted int
	masked  bool
	stopped bool
	aborted bool
}

func newStreamGate(stops, markers []string, emit func(string) bool) *streamGate {
	hold := 1
	for _, s := range append(append([]string{}, stops...), markers...) {
		if len(s) > hold {
			hold = len(s)
		}
	}
	return &streamGate{stops: stops, markers: markers, maxHold: hold, emit: emit}
}

func (g *streamGate) push(s string) {
	// Once stopped (stop string hit) or aborted (consumer left) the text is frozen.
	if g.aborted || g.stopped || s == "" {
		return
	}
	g.full += s

	if at, ok := firstIndexOf(g.full, g.stops); ok {
		if !g.masked {
			g.flushTo(at)
		}
		g.full = g.full[:at]
		g.stopped = true
		return
	}

	if g.masked {
		return
	}

	if at, ok := firstIndexOf(g.full, g.markers); ok {
		g.flushTo(at)
		g.masked = true
		return
	}

	g.flushTo(len(g.full) - (g.maxHold - 1))
}

// finish flushes any remaining safe text (unless masked or stopped) and returns the full accumulated text.
func (g *streamGate) finish() string {
	if !g.masked && !g.stopped {
		g.flushTo(len(g.full))
	}
	return g.full
}

func (g *streamGate) flushTo(n int) {
	if n > len(g.full) {
		n = len(g.full)
	}
	if n <= g.emitted {
		return
	}
	seg := g.full[g.emitted:n]
	g.emitted = n
	if g.emit != nil && seg != "" {
		if !g.emit(seg) {
			g.aborted = true
		}
	}
}

// firstIndexOf returns the earliest index at which any of the needles occurs in s.
func firstIndexOf(s string, needles []string) (int, bool) {
	best := -1
	for _, n := range needles {
		if n == "" {
			continue
		}
		if i := strings.Index(s, n); i >= 0 && (best < 0 || i < best) {
			best = i
		}
	}
	if best < 0 {
		return 0, false
	}
	return best, true
}

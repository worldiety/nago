/*
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { type ProgramChannel, attach } from '@/components/gostrict/channels';
import { type InputSample } from '@/shared/inputEvents';
import { GostrictMessage, GostrictProgram, InputEventTypeValues } from '@/shared/proto/nprotoc_gen';

/**
 * A program of the gostrict subset, see package pkg/gostrict, runs in the frontend of a component. The backend
 * serves its translated script with the schema of its boundary as a manifest under an immutable URL. A session
 * evaluates the script, calls its exported functions and passes the messages of the backend and the input events of
 * the component to the handlers of the program between two frames.
 *
 * The values cross the protocol as JSON, see package presentation/ui/program: a record is an object with the
 * names of its fields, which the translated program reads as they are, a list is an array, and the values, which
 * JSON can not express, are objects with a key starting with $: {"$m": entries} for a map, {"$": tag, ...fields}
 * for a variant of a sealed interface, {"$f": "NaN"} for NaN, the infinities and -0, {"$b": "123"} for a big.Int,
 * {"$e": message} for an error
 * and {"$r": ref} for a state of the engine of the browser. So a reviver of JSON.parse and a replacer of
 * JSON.stringify convert all values without a schema.
 */

export interface TypeJSON {
	kind: string;
	name?: string;
	elem?: TypeJSON;
	key?: TypeJSON;
}

interface FuncJSON {
	params: TypeJSON[] | null;
	results: TypeJSON[] | null;
}

interface RecordJSON {
	fields: { name: string; js: string; type: TypeJSON }[] | null;
}

export interface Manifest {
	hash: string;
	js: string;
	exports: Record<string, FuncJSON>;
	records: Record<string, RecordJSON>;
	unions: Record<string, string[]>;
	handlers: Record<string, string>;
	// states maps the full name of each state to the hash of its structure, which is empty, if a new version can
	// not take it over
	states: Record<string, string>;
	// sourceMap maps the script to the Go sources
	sourceMap?: unknown;
}

/**
 * Previous is the state of the former version of a program, which a new version takes over, if the structure of
 * the state has not changed, e.g. after an edit of the program.
 */
export interface Previous {
	stateType: string;
	hash: string;
	state: unknown;
}

// inputEvent is the record of the input events, see package lib/input
const inputEvent = 'go.wdy.de/nago/pkg/gostrict/lib/input.Event';

const manifests = new Map<string, Promise<Manifest>>();

/**
 * loadManifest fetches the manifest once per URL. A failed request is not cached, so a later mount retries it.
 */
export function loadManifest(url: string): Promise<Manifest> {
	let p = manifests.get(url);
	if (!p) {
		p = fetch(url, { credentials: 'same-origin' }).then((res) => {
			if (!res.ok) throw new Error(`cannot load the program ${url}: ${res.status}`);
			return res.json() as Promise<Manifest>;
		});
		p.catch(() => manifests.delete(url));
		manifests.set(url, p);
	}
	return p;
}

type Factory = (...libs: unknown[]) => (host: object) => Record<string, any>;

// factories caches the evaluated scripts by their hash and the names of the libraries, which they receive
const factories = new Map<string, Factory>();

function factory(m: Manifest, libNames: string[]): Factory {
	const key = m.hash + '\x00' + libNames.join(',');
	let f = factories.get(key);
	if (!f) {
		// The script is the translation of a checked program, which the server has compiled, and refers to the
		// libraries, e.g. THREE, through its bindings. The indirect eval evaluates it in the global scope, so it can
		// not see the variables of this module. The wrapper starts on the first line of the script, so the lines of
		// the source map stay valid, and the browser shows the Go sources in its debugger.
		let src =
			'(function (' +
			libNames.join(', ') +
			') { return ' +
			m.js +
			'\n})\n//# sourceURL=gostrict/' +
			m.hash +
			'.js\n';
		if (m.sourceMap) {
			src +=
				'//# sourceMappingURL=data:application/json;charset=utf-8;base64,' +
				base64(JSON.stringify(m.sourceMap)) +
				'\n';
		}
		f = (0, eval)(src) as Factory;
		factories.set(key, f);
	}
	return f;
}

function base64(s: string): string {
	const bytes = new TextEncoder().encode(s);
	let bin = '';
	for (let i = 0; i < bytes.length; i += 0x8000) {
		bin += String.fromCharCode(...bytes.subarray(i, i + 0x8000));
	}
	return btoa(bin);
}

type Queued =
	| { kind: 'msg'; msg: GostrictMessage }
	| { kind: 'input'; sample: InputSample }
	| { kind: 'event'; type: string; values: Record<string, unknown> };

export interface SessionOptions {
	// id is the id of the component, which the backend uses for its messages
	id: string;
	program: GostrictProgram;
	manifest: Manifest;
	// libs are the libraries, which the bindings of the program use, e.g. { THREE }
	libs: Record<string, unknown>;
	// host is the object, which the bindings call, e.g. show(scene, camera)
	host: object;
	// before is called before the program handles anything in a frame, e.g. to reset the transformation of a canvas
	before?: () => void;
	// changed is called after the program has handled messages or input events outside of a frame
	changed?: () => void;
	// previous is the state of the former version of the program
	previous?: Previous | null;
}

/**
 * Session runs one instance of a program. The values of a checked program can not fail, but the bindings of a
 * library may throw, so a session stops at the first exception and logs it.
 */
export class Session {
	readonly manifest: Manifest;
	private readonly program: GostrictProgram;
	private readonly exports: Record<string, any>;
	// methods maps the full name of each variant of a sealed interface to the methods of its type, which box it
	private readonly methods = new Map<string, unknown>();
	// tags maps the methods of the type of each variant back to its full name
	private readonly tags = new Map<unknown, string>();
	private readonly channel: ProgramChannel;
	private readonly before?: () => void;
	private readonly changed?: () => void;
	private readonly previous: Previous | null;
	private state: unknown = undefined;
	private stateType = '';
	private queue: Queued[] = [];
	// moveAt is the index in the queue of the last pointer move of each pointer, which may be replaced
	private moveAt = new Map<number, number>();
	private scheduled = 0;
	private outbox: GostrictMessage[] = [];
	private failed = false;
	private closed = false;

	constructor(opts: SessionOptions) {
		this.manifest = opts.manifest;
		this.program = opts.program;
		this.before = opts.before;
		this.changed = opts.changed;
		this.previous = opts.previous ?? null;
		const names = Object.keys(opts.libs);
		const make = factory(opts.manifest, names);
		this.exports = make(...names.map((n) => opts.libs[n]))(opts.host);
		const unions: Record<string, Record<string, [unknown, unknown]>> = this.exports.$unions ?? {};
		for (const variants of Object.values(unions)) {
			for (const [tag, [methods]] of Object.entries(variants)) {
				this.methods.set(tag, methods);
				this.tags.set(methods, tag);
			}
		}
		this.channel = attach(opts.id, (msgs) => this.enqueueMessages(msgs));
	}

	/**
	 * init calls the Init function of the program with its arguments and keeps the state, which it returns.
	 */
	init() {
		const name = this.program.init ?? '';
		if (!name) return;
		const f = this.manifest.exports[name];
		if (!f) {
			this.fail(new Error(`the program does not export ${name}`));
			return;
		}
		const args = this.program.args ? (this.decode(this.program.args) as unknown[]) : [];
		const results = this.invoke(name, args);
		const res = f.results ?? [];
		if (res.length > 0 && res[0].kind === 'state') {
			this.state = results[0];
			this.stateType = res[0].name ?? '';
			const prev = this.previous;
			const hash = this.manifest.states?.[this.stateType] ?? '';
			if (prev && prev.stateType === this.stateType && hash !== '' && prev.hash === hash) {
				// Init has built the scene of the new version, the state of the former one continues
				this.state = prev.state;
			}
		}
		this.flushOutbox();
	}

	/**
	 * takeOver returns the state for a new version of the program.
	 */
	takeOver(): Previous | null {
		if (!this.stateType || this.failed) return null;
		return { stateType: this.stateType, hash: this.manifest.states?.[this.stateType] ?? '', state: this.state };
	}

	/**
	 * callRaw calls an exported function with the values of the program and returns its results, without sending
	 * the messages, which it returns. It throws, if a binding throws.
	 */
	callRaw(name: string, args: unknown[]): unknown[] {
		const f = this.manifest.exports[name];
		if (!f) throw new Error(`the program does not export ${name}`);
		const ret = this.exports[name](...args);
		return (f.results ?? []).length > 1 ? (ret as unknown[]) : [ret];
	}

	/**
	 * animated tells, whether the program has a function for each frame.
	 */
	get animated(): boolean {
		return !!this.program.frame && !!this.manifest.exports[this.program.frame];
	}

	/**
	 * frame handles the queued messages and input events and calls the Frame function with the time in seconds.
	 */
	frame(t: number) {
		this.before?.();
		this.flushQueue();
		if (this.animated) this.call(this.program.frame!, [t]);
		this.flushOutbox();
	}

	/**
	 * resize calls the Resize function with the size of the area in CSS pixels.
	 */
	resize(width: number, height: number, dpr: number) {
		const name = this.program.resize ?? '';
		const f = this.manifest.exports[name];
		if (!f || height <= 0) return;
		const n = (f.params ?? []).filter((p) => !this.isState(p)).length;
		const args = n === 1 ? [width / height] : n === 2 ? [width, height] : [width, height, dpr];
		this.call(name, args);
		this.flushOutbox();
	}

	/**
	 * input queues an input event of the component for the handler of input.Event. Only the last move of a
	 * pointer and the sum of the wheel events since the last frame are kept.
	 */
	input(s: InputSample) {
		if (!this.manifest.handlers[inputEvent]) return;
		const isMove = s.type === InputEventTypeValues.InputEventPointerMove;
		if (isMove) {
			const at = this.moveAt.get(s.pointerId);
			if (at !== undefined) {
				const q = this.queue[at];
				if (q.kind === 'input') {
					q.sample = s;
					return;
				}
			}
		}
		if (s.type === InputEventTypeValues.InputEventMouseWheel) {
			const last = this.queue[this.queue.length - 1];
			if (last?.kind === 'input' && last.sample.type === s.type) {
				s.deltaX += last.sample.deltaX;
				s.deltaY += last.sample.deltaY;
				s.deltaZ += last.sample.deltaZ;
				last.sample = s;
				return;
			}
		}
		this.queue.push({ kind: 'input', sample: s });
		if (isMove) {
			this.moveAt.set(s.pointerId, this.queue.length - 1);
		} else {
			// a later move must not overtake this event
			this.moveAt.delete(s.pointerId);
		}
		this.schedule();
	}

	/**
	 * emit queues an event of the host, e.g. a loaded image, for the handler of the record with the full name. The
	 * values are the fields of the record by their name in Go. A program without such a handler ignores it.
	 */
	emit(type: string, values: Record<string, unknown>) {
		if (!this.manifest.handlers[type]) return;
		this.queue.push({ kind: 'event', type, values });
		this.schedule();
	}

	/**
	 * wantsInput tells, whether the program handles input events.
	 */
	get wantsInput(): boolean {
		return !!this.manifest.handlers[inputEvent];
	}

	close() {
		this.closed = true;
		this.channel.detach();
		if (this.scheduled) cancelAnimationFrame(this.scheduled);
		this.queue = [];
	}

	private enqueueMessages(msgs: GostrictMessage[]) {
		for (const msg of msgs) this.queue.push({ kind: 'msg', msg });
		this.schedule();
	}

	// schedule handles the queue in the next frame, if the program has no frame function, which does it anyway
	private schedule() {
		if (this.animated || this.scheduled || this.closed) return;
		this.scheduled = requestAnimationFrame(() => {
			this.scheduled = 0;
			this.before?.();
			this.flushQueue();
			this.flushOutbox();
			this.changed?.();
		});
	}

	private flushQueue() {
		const queue = this.queue;
		this.queue = [];
		this.moveAt.clear();
		for (const q of queue) {
			switch (q.kind) {
				case 'msg':
					this.handleMessage(q.msg);
					break;
				case 'input':
					this.handleInput(q.sample);
					break;
				case 'event':
					this.handleEvent(q.type, q.values);
					break;
			}
		}
	}

	private handleMessage(m: GostrictMessage) {
		const type = m.type ?? '';
		const name = this.manifest.handlers[type];
		if (!name) {
			console.warn('gostrict: the program does not handle the message', type);
			return;
		}
		let record: unknown;
		try {
			record = this.decode(m.json ?? '');
		} catch (e) {
			console.warn('gostrict: cannot decode the message', type, e);
			return;
		}
		this.handle(name, type, record);
	}

	// handle passes the record to its handler, which may take it as a variant of a sealed interface
	private handle(name: string, type: string, record: unknown) {
		const param = (this.manifest.exports[name].params ?? [])[1];
		this.call(name, [param.kind === 'union' ? this.box(type, record) : record]);
	}

	// box returns the value of a sealed interface, which holds the record of the variant
	private box(tag: string, record: unknown): unknown {
		const m = this.methods.get(tag);
		if (m === undefined) throw new Error(`${tag} is no variant of a sealed interface of the program`);
		return { v: record, m };
	}

	private handleEvent(type: string, values: Record<string, unknown>) {
		const rec = this.manifest.records[type];
		const name = this.manifest.handlers[type];
		if (!rec || !name) return;
		const record: Record<string, unknown> = {};
		for (const f of rec.fields ?? []) record[f.js] = values[f.name];
		this.handle(name, type, record);
	}

	private handleInput(s: InputSample) {
		this.handleEvent(inputEvent, {
			Type: s.type,
			X: s.x,
			Y: s.y,
			Code: s.code,
			PointerID: s.pointerId | 0,
			Buttons: s.buttons >>> 0,
			Pressure: s.pressure,
			Shift: s.shift,
			Ctrl: s.ctrl,
			Alt: s.alt,
			Meta: s.meta,
			DeltaX: s.deltaX,
			DeltaY: s.deltaY,
			DeltaZ: s.deltaZ,
			Key: s.key,
			Repeat: s.repeat,
		});
	}

	private isState(t: TypeJSON): boolean {
		return t.kind === 'state' && t.name === this.stateType;
	}

	// call calls an exported function with the state, if it takes one, and keeps the messages, which it returns
	private call(name: string, args: unknown[]): unknown[] {
		const f = this.manifest.exports[name];
		if (!f) return [];
		const params = f.params ?? [];
		if (params.length > 0 && this.isState(params[0])) args = [this.state, ...args];
		return this.invoke(name, args);
	}

	private invoke(name: string, args: unknown[]): unknown[] {
		if (this.failed || this.closed) return [];
		const f = this.manifest.exports[name];
		let ret: unknown;
		try {
			ret = this.exports[name](...args);
		} catch (e) {
			this.fail(e);
			return [];
		}
		const results = f.results ?? [];
		const values = results.length > 1 ? (ret as unknown[]) : [ret];
		results.forEach((t, i) => {
			if (t.kind === 'out') this.collect(values[i] as unknown[], t.elem!);
		});
		return values;
	}

	private collect(out: unknown[], elem: TypeJSON) {
		for (const v of out) {
			if (elem.kind === 'union') {
				const box = v as { v: unknown; m: unknown };
				this.outbox.push(new GostrictMessage(this.tags.get(box.m), this.encode(box.v)));
				continue;
			}
			this.outbox.push(new GostrictMessage(elem.name, this.encode(v)));
		}
	}

	private flushOutbox() {
		if (this.outbox.length === 0) return;
		const msgs = this.outbox;
		this.outbox = [];
		this.channel.send(msgs);
	}

	private fail(e: unknown) {
		this.failed = true;
		console.error('gostrict: the program has stopped', e);
	}

	/**
	 * decode parses the JSON of values of the program. A state of the engine of the browser is resolved by state.
	 */
	decode(json: string, state?: (ref: number) => unknown): unknown {
		return JSON.parse(json, (_key, v) => {
			if (v === null || typeof v !== 'object' || Array.isArray(v)) return v;
			if ('$f' in v) return special[v.$f as string] ?? 0;
			if ('$b' in v) return BigInt(v.$b);
			if ('$m' in v) return new Map(v.$m);
			if ('$r' in v && state) return state(v.$r);
			if ('$' in v) {
				const tag = v.$ as string;
				delete v.$;
				return this.box(tag, v);
			}
			return v;
		});
	}

	/**
	 * encode returns the JSON of a value of the program.
	 */
	encode(value: unknown): string {
		const tags = this.tags;
		return JSON.stringify(value, (_key, v) => {
			if (typeof v === 'bigint') return { $b: v.toString() };
			if (typeof v === 'number') {
				if (v !== v) return { $f: 'NaN' };
				if (v === Infinity) return { $f: 'Inf' };
				if (v === -Infinity) return { $f: '-Inf' };
				if (v === 0 && 1 / v < 0) return { $f: '-0' };
				return v;
			}
			if (v instanceof Map) return { $m: Array.from(v.entries()) };
			if (v !== null && typeof v === 'object' && 'm' in v && 'v' in v && Object.keys(v).length === 2) {
				const tag = tags.get(v.m);
				if (tag !== undefined) return { ...v.v, $: tag };
				if (typeof v.m?.Error === 'function') return { $e: v.m.Error(v.v) };
			}
			return v;
		});
	}
}

// special are the floats, which JSON can not express
const special: Record<string, number> = { NaN: NaN, Inf: Infinity, '-Inf': -Infinity, '-0': -0 };

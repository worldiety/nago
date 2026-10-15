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
import {
	GostrictBool,
	GostrictFloat,
	GostrictInt,
	GostrictList,
	GostrictMessage,
	GostrictNegZero,
	GostrictProgram,
	GostrictStr,
	type GostrictValue,
	GostrictValues,
	GostrictVariant,
	InputEventTypeValues,
} from '@/shared/proto/nprotoc_gen';

/**
 * A program of the gostrict subset, see package pkg/gostrict, runs in the frontend of a component. The backend
 * serves its translated script with the schema of its boundary as a manifest under an immutable URL. A session
 * evaluates the script, calls its exported functions, converts the values of the protocol with the schema and
 * passes the messages of the backend and the input events of the component to the handlers of the program between
 * two frames.
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
	private readonly news: Record<string, (...fields: unknown[]) => unknown>;
	private readonly unions: Record<string, Record<string, [unknown, (...fields: unknown[]) => unknown]>>;
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
		this.news = this.exports.$new ?? {};
		this.unions = this.exports.$unions ?? {};
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
		const params = f.params ?? [];
		const wire = this.program.args?.value ?? [];
		const args = params.map((t, i) => this.fromWire(wire[i], t));
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
		this.handle(name, type, this.recordFields(type, m.fields?.value ?? []));
	}

	// handle passes the record to its handler, which may take it as a variant of a sealed interface
	private handle(name: string, type: string, fields: unknown[]) {
		const param = (this.manifest.exports[name].params ?? [])[1];
		const arg = param.kind === 'union' ? this.unions[param.name!][type][1](...fields) : this.news[type](...fields);
		this.call(name, [arg]);
	}

	private handleEvent(type: string, values: Record<string, unknown>) {
		const rec = this.manifest.records[type];
		const name = this.manifest.handlers[type];
		if (!rec || !name) return;
		this.handle(
			name,
			type,
			(rec.fields ?? []).map((f) => values[f.name])
		);
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
				const variant = this.toWire(v, elem) as GostrictVariant;
				this.outbox.push(new GostrictMessage(variant.tag, variant.fields));
				continue;
			}
			this.outbox.push(new GostrictMessage(elem.name, (this.toWire(v, elem) as GostrictList).v));
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

	private recordFields(name: string, values: GostrictValue[]): unknown[] {
		const rec = this.manifest.records[name];
		return (rec?.fields ?? []).map((f, i) => this.fromWire(values[i], f.type));
	}

	// fromWire converts a value of the protocol into the value of the program
	fromWire(v: GostrictValue | undefined, t: TypeJSON): unknown {
		switch (t.kind) {
			case 'bool':
				return v instanceof GostrictBool ? !!v.v : false;
			case 'int32':
				return v instanceof GostrictInt ? (v.v ?? 0) | 0 : 0;
			case 'uint32':
				return v instanceof GostrictInt ? (v.v ?? 0) >>> 0 : 0;
			case 'float64':
				if (v instanceof GostrictNegZero) return -0;
				return v instanceof GostrictFloat ? (v.v ?? 0) : 0;
			case 'string':
				return v instanceof GostrictStr ? (v.v ?? '') : '';
			case 'list':
			case 'out':
				return this.items(v).map((e) => this.fromWire(e, t.elem!));
			case 'map':
				return new Map(
					this.items(v).map((e) => {
						const kv = this.items(e);
						return [this.fromWire(kv[0], t.key!), this.fromWire(kv[1], t.elem!)];
					})
				);
			case 'record':
				return this.news[t.name!](...this.recordFields(t.name!, this.items(v)));
			case 'union': {
				if (!(v instanceof GostrictVariant)) throw new Error(`expected a variant of ${t.name}`);
				const tag = v.tag ?? '';
				const variant = this.unions[t.name!]?.[tag];
				if (!variant) throw new Error(`${tag} is no variant of ${t.name}`);
				return variant[1](...this.recordFields(tag, v.fields?.value ?? []));
			}
		}
		throw new Error(`${t.kind} can not cross the boundary`);
	}

	private items(v: GostrictValue | undefined): GostrictValue[] {
		return v instanceof GostrictList ? (v.v?.value ?? []) : [];
	}

	// toWire converts a value of the program into the protocol
	toWire(v: unknown, t: TypeJSON): GostrictValue {
		switch (t.kind) {
			case 'bool':
				return new GostrictBool(v as boolean);
			case 'int32':
			case 'uint32':
				return new GostrictInt(v as number);
			case 'float64':
				return Object.is(v, -0) ? new GostrictNegZero() : new GostrictFloat(v as number);
			case 'string':
				return new GostrictStr(v as string);
			case 'list':
			case 'out':
				return new GostrictList(new GostrictValues((v as unknown[]).map((e) => this.toWire(e, t.elem!))));
			case 'map':
				return new GostrictList(
					new GostrictValues(
						Array.from((v as Map<unknown, unknown>).entries()).map(
							([k, e]) =>
								new GostrictList(new GostrictValues([this.toWire(k, t.key!), this.toWire(e, t.elem!)]))
						)
					)
				);
			case 'record': {
				const obj = v as Record<string, unknown>;
				const rec = this.manifest.records[t.name!];
				return new GostrictList(
					new GostrictValues((rec.fields ?? []).map((f) => this.toWire(obj[f.js], f.type)))
				);
			}
			case 'union': {
				const box = v as { v: unknown; m: unknown };
				for (const tag of this.manifest.unions[t.name!] ?? []) {
					if (this.unions[t.name!][tag][0] === box.m) {
						const fields = this.toWire(box.v, { kind: 'record', name: tag }) as GostrictList;
						return new GostrictVariant(tag, fields.v);
					}
				}
				throw new Error(`the value is no variant of ${t.name}`);
			}
		}
		throw new Error(`${t.kind} can not cross the boundary`);
	}
}

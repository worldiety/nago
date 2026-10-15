/*
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { Session, type TypeJSON, loadManifest } from '@/components/gostrict/session';
import {
	GostrictCall,
	GostrictError,
	GostrictNil,
	GostrictProgram,
	GostrictRef,
	GostrictReturn,
	type GostrictValue,
	GostrictValues,
} from '@/shared/proto/nprotoc_gen';

/**
 * The engine of package presentation/ui/program runs programs in the browser on behalf of the backend: each
 * GostrictCall calls an exported function of an instance, which is created on its first call. A state stays in
 * the instance and crosses the protocol as a reference.
 */

interface Instance {
	session: Session;
	refs: Map<number, unknown>;
	ids: Map<unknown, number>;
	next: number;
}

const instances = new Map<string, Promise<Instance>>();

async function instance(id: string, bundle: string): Promise<Instance> {
	let p = instances.get(id);
	if (!p) {
		p = loadManifest(bundle).then((manifest) => ({
			session: new Session({ id: '$engine:' + id, program: new GostrictProgram(), manifest, libs: {}, host: {} }),
			refs: new Map(),
			ids: new Map(),
			next: 1,
		}));
		p.catch(() => instances.delete(id));
		instances.set(id, p);
	}
	return p;
}

/**
 * call answers a GostrictCall.
 */
export async function call(c: GostrictCall): Promise<GostrictReturn> {
	const id = c.instance ?? '';
	if (!c.name) {
		const p = instances.get(id);
		instances.delete(id);
		p?.then((i) => i.session.close()).catch(() => {});
		return new GostrictReturn();
	}
	try {
		const inst = await instance(id, c.bundle ?? '');
		const f = inst.session.manifest.exports[c.name];
		if (!f) throw new Error(`the program does not export ${c.name}`);
		const wire = c.args?.value ?? [];
		const args = (f.params ?? []).map((t, i) => fromWire(inst, wire[i], t));
		const results = inst.session.callRaw(c.name, args);
		const values = (f.results ?? []).map((t, i) => toWire(inst, results[i], t));
		return new GostrictReturn(new GostrictValues(values));
	} catch (e) {
		return new GostrictReturn(undefined, String((e as Error)?.message ?? e));
	}
}

function fromWire(inst: Instance, v: GostrictValue | undefined, t: TypeJSON): unknown {
	if (t.kind === 'state') {
		const ref = v instanceof GostrictRef ? (v.v ?? 0) : 0;
		if (!inst.refs.has(ref)) throw new Error('unknown state');
		return inst.refs.get(ref);
	}
	return inst.session.fromWire(v, t);
}

function toWire(inst: Instance, v: unknown, t: TypeJSON): GostrictValue {
	switch (t.kind) {
		case 'state': {
			let ref = inst.ids.get(v);
			if (ref === undefined) {
				ref = inst.next++;
				inst.ids.set(v, ref);
				inst.refs.set(ref, v);
			}
			return new GostrictRef(ref);
		}
		case 'error': {
			if (v === null || v === undefined) return new GostrictNil();
			const box = v as { v: unknown; m: { Error(v: unknown): string } };
			return new GostrictError(box.m.Error(box.v));
		}
	}
	return inst.session.toWire(v, t);
}

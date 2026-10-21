/*
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { Session, loadManifest } from '@/components/gostrict/session';
import { GostrictCall, GostrictProgram, GostrictReturn } from '@/shared/proto/nprotoc_gen';

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
		const args = inst.session.decode(c.args || '[]', (ref) => {
			if (!inst.refs.has(ref)) throw new Error('unknown state');
			return inst.refs.get(ref);
		}) as unknown[];
		const results = inst.session.callRaw(c.name, args);
		// a state stays in the instance and crosses the protocol as its reference
		const values = (f.results ?? []).map((t, i) =>
			t.kind === 'state' ? { $r: ref(inst, results[i]) } : results[i]
		);
		return new GostrictReturn(inst.session.encode(values));
	} catch (e) {
		return new GostrictReturn(undefined, String((e as Error)?.message ?? e));
	}
}

function ref(inst: Instance, state: unknown): number {
	let r = inst.ids.get(state);
	if (r === undefined) {
		r = inst.next++;
		inst.ids.set(state, r);
		inst.refs.set(r, state);
	}
	return r;
}

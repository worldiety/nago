/*
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { type Previous, Session, loadManifest } from '@/components/gostrict/session';
import { listenInput } from '@/shared/inputEvents';
import { GostrictProgram, InputEventTypeValues } from '@/shared/proto/nprotoc_gen';

export interface StartOptions {
	// id is the id of the component
	id: string;
	program: GostrictProgram;
	// root is the element, whose size the program fills
	root: HTMLElement;
	// input is the element, whose input events go to the program
	input: HTMLElement;
	libs: Record<string, unknown>;
	host: object;
	// resize adapts the drawing surface to the size of root in CSS pixels before the program is told about it
	resize(width: number, height: number, dpr: number): void;
	// render draws the result of the program after each frame or change
	render(): void;
	// before is called before the program handles anything in a frame
	before?: () => void;
	// attached receives the session before Init is called, e.g. for the events of the host
	attached?: (s: Session) => void;
	// previous is the state of the former version of the program, see Session.takeOver
	previous?: Previous | null;
}

export interface ProgramHandle {
	session: Session;
	stop(): void;
}

/**
 * startProgram loads the program of a component and runs it: it calls Init, tells it the size of the component,
 * passes the input events and calls the Frame function in each animation frame, until stop is called. The root
 * element shows the average time of a frame of the program in milliseconds in its attribute data-frame-ms.
 */
export async function startProgram(opts: StartOptions): Promise<ProgramHandle | null> {
	let manifest;
	try {
		manifest = await loadManifest(opts.program.bundle!);
	} catch (e) {
		console.error('gostrict: cannot load the program', e);
		return null;
	}

	const session = new Session({
		id: opts.id,
		program: opts.program,
		manifest,
		libs: opts.libs,
		host: opts.host,
		before: opts.before,
		changed: () => opts.render(),
		previous: opts.previous,
	});
	opts.attached?.(session);

	const size = () => {
		const rect = opts.root.getBoundingClientRect();
		const dpr = window.devicePixelRatio || 1;
		opts.resize(rect.width, rect.height, dpr);
		session.resize(rect.width, rect.height, dpr);
	};

	// the size comes first, because Init may build a scene for the aspect ratio
	session.init();
	size();
	opts.render();

	const observer = new ResizeObserver(() => {
		size();
		if (!session.animated) opts.render();
	});
	observer.observe(opts.root);

	let stopInput = () => {};
	if (session.wantsInput) {
		const types = new Set<InputEventTypeValues>(opts.program.inputs?.value ?? []);
		stopInput = listenInput(opts.input, types, (s) => session.input(s), {
			preventWheel: true,
			capturePointer: true,
		});
	}

	let raf = 0;
	if (session.animated) {
		const t0 = performance.now();
		let avg = 0;
		let n = 0;
		const tick = (now: number) => {
			raf = requestAnimationFrame(tick);
			const start = performance.now();
			session.frame((now - t0) / 1000);
			opts.render();
			// an exponential average of the time, which the program and the rendering take per frame
			const ms = performance.now() - start;
			avg = n < 10 ? (avg * n + ms) / (n + 1) : avg * 0.95 + ms * 0.05;
			n++;
			if (n % 30 === 0) opts.root.dataset.frameMs = avg.toFixed(3);
		};
		raf = requestAnimationFrame(tick);
	}

	return {
		session,
		stop() {
			cancelAnimationFrame(raf);
			observer.disconnect();
			stopInput();
			session.close();
		},
	};
}

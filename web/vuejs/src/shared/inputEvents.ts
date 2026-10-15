/*
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { InputEventTypeValues } from '@/shared/proto/nprotoc_gen';

/**
 * InputSample is an input event at an element. It has the fields of the protocol type InputEvent, which the
 * backend receives through a listener, and of the record input.Event, which a gostrict program receives in the
 * frontend. Both use the same DOM listeners, see listenInput.
 */
export interface InputSample {
	type: InputEventTypeValues;
	// x and y are the position in CSS pixels relative to the element
	x: number;
	y: number;
	// code is the physical key, e.g. KeyA
	code: string;
	pointerId: number;
	buttons: number;
	pressure: number;
	shift: boolean;
	ctrl: boolean;
	alt: boolean;
	meta: boolean;
	deltaX: number;
	deltaY: number;
	deltaZ: number;
	// key is the value of the key in the layout of the keyboard, e.g. a or Enter
	key: string;
	repeat: boolean;
}

export function emptySample(type: InputEventTypeValues): InputSample {
	return {
		type,
		x: 0,
		y: 0,
		code: '',
		pointerId: 0,
		buttons: 0,
		pressure: 0,
		shift: false,
		ctrl: false,
		alt: false,
		meta: false,
		deltaX: 0,
		deltaY: 0,
		deltaZ: 0,
		key: '',
		repeat: false,
	};
}

const pointerMappings: [string, InputEventTypeValues][] = [
	['pointerdown', InputEventTypeValues.InputEventPointerDown],
	['pointerup', InputEventTypeValues.InputEventPointerUp],
	['pointermove', InputEventTypeValues.InputEventPointerMove],
	['pointercancel', InputEventTypeValues.InputEventPointerCancel],
	['pointerleave', InputEventTypeValues.InputEventPointerLeave],
	['pointerout', InputEventTypeValues.InputEventPointerOut],
	['pointerover', InputEventTypeValues.InputEventPointerOver],
];

const mouseMappings: [string, InputEventTypeValues][] = [
	['mousedown', InputEventTypeValues.InputEventMouseDown],
	['mouseup', InputEventTypeValues.InputEventMouseUp],
	['mousemove', InputEventTypeValues.InputEventMouseMove],
	['mouseenter', InputEventTypeValues.InputEventMouseEnter],
	['mouseleave', InputEventTypeValues.InputEventMouseLeave],
	['mouseover', InputEventTypeValues.InputEventMouseOver],
	['mouseout', InputEventTypeValues.InputEventMouseOut],
];

const keyMappings: [string, InputEventTypeValues][] = [
	['keydown', InputEventTypeValues.InputEventKeyDown],
	['keyup', InputEventTypeValues.InputEventKeyUp],
];

function modifiers(s: InputSample, e: MouseEvent | KeyboardEvent): InputSample {
	s.shift = e.shiftKey;
	s.ctrl = e.ctrlKey;
	s.alt = e.altKey;
	s.meta = e.metaKey;
	return s;
}

function mouseSample(elem: Element, type: InputEventTypeValues, e: MouseEvent): InputSample {
	const rect = elem.getBoundingClientRect();
	const s = modifiers(emptySample(type), e);
	s.x = e.clientX - rect.left;
	s.y = e.clientY - rect.top;
	s.buttons = e.buttons;
	return s;
}

export interface ListenOptions {
	// preventWheel stops the page from scrolling, while the wheel turns over the element
	preventWheel?: boolean;
	// capturePointer keeps the pointer at the element from pointerdown to pointerup, e.g. for dragging, and focuses
	// the element for the keys
	capturePointer?: boolean;
}

/**
 * listenInput registers the DOM listeners for the requested types at the element and passes each event to the
 * sink. It returns the function, which removes the listeners.
 */
export function listenInput(
	elem: Element,
	requested: Set<InputEventTypeValues>,
	sink: (s: InputSample) => void,
	opts: ListenOptions = {}
): () => void {
	const listeners: Array<{ domEvent: string; fn: EventListener; options?: AddEventListenerOptions }> = [];
	const add = (domEvent: string, fn: EventListener, options?: AddEventListenerOptions) => {
		elem.addEventListener(domEvent, fn, options);
		listeners.push({ domEvent, fn, options });
	};

	if (requested.has(InputEventTypeValues.InputEventInvalidate)) {
		add('invalidated', () => sink(emptySample(InputEventTypeValues.InputEventInvalidate)));
	}

	for (const [domEvent, evtType] of pointerMappings) {
		if (!requested.has(evtType)) continue;
		add(domEvent, (e) => {
			const pe = e as PointerEvent;
			if (opts.capturePointer && domEvent === 'pointerdown') {
				try {
					elem.setPointerCapture(pe.pointerId);
				} catch {
					// the pointer may already be gone
				}
				// the keys go to the element, which the user has touched last
				(elem as HTMLElement).focus?.({ preventScroll: true });
			}
			const s = mouseSample(elem, evtType, pe);
			s.pointerId = pe.pointerId;
			s.pressure = pe.pressure;
			sink(s);
		});
	}

	for (const [domEvent, evtType] of mouseMappings) {
		if (!requested.has(evtType)) continue;
		add(domEvent, (e) => sink(mouseSample(elem, evtType, e as MouseEvent)));
	}

	if (requested.has(InputEventTypeValues.InputEventMouseWheel)) {
		add(
			'wheel',
			(e) => {
				const we = e as WheelEvent;
				if (opts.preventWheel) we.preventDefault();
				const s = mouseSample(elem, InputEventTypeValues.InputEventMouseWheel, we);
				// lines and pages are converted into pixels, so that a program needs only one unit
				const unit = we.deltaMode === 1 ? 16 : we.deltaMode === 2 ? elem.clientHeight : 1;
				s.deltaX = we.deltaX * unit;
				s.deltaY = we.deltaY * unit;
				s.deltaZ = we.deltaZ * unit;
				sink(s);
			},
			{ passive: !opts.preventWheel }
		);
	}

	if (requested.has(InputEventTypeValues.InputEventBlur)) {
		add('blur', () => sink(emptySample(InputEventTypeValues.InputEventBlur)));
	}

	for (const [domEvent, evtType] of keyMappings) {
		if (!requested.has(evtType)) continue;
		add(domEvent, (e) => {
			const ke = e as KeyboardEvent;
			const s = modifiers(emptySample(evtType), ke);
			s.code = ke.code;
			s.key = ke.key;
			s.repeat = ke.repeat;
			sink(s);
		});
	}

	return () => {
		for (const { domEvent, fn, options } of listeners) {
			elem.removeEventListener(domEvent, fn, options);
		}
	};
}

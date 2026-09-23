/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { domToCanvas } from 'modern-screenshot';
import { nextTick } from 'vue';

/**
 * DEFAULT_MAX_EDGE is the largest edge typical vision models process without downscaling it themselves.
 */
export const DEFAULT_MAX_EDGE = 1568;

export interface CaptureOptions {
	selector?: string;
	image: boolean;
	snapshot: boolean;
	maxEdge?: number;
}

export interface CaptureResult {
	pngBase64: string;
	width: number;
	height: number;
	snapshot: string;
}

/**
 * capture renders the current DOM (or the element matched by selector) into a PNG and/or an accessibility
 * snapshot. The PNG is produced by re-rendering the DOM through an SVG foreignObject, thus it is an
 * approximation: cross-origin images without CORS, iframes and videos may be missing.
 */
export async function capture(opts: CaptureOptions): Promise<CaptureResult> {
	await settle();

	let target: Element | null = document.body;
	if (opts.selector) {
		target = document.querySelector(opts.selector);
		if (!target) {
			throw new Error(`no element matches selector ${JSON.stringify(opts.selector)}`);
		}
	}

	const res: CaptureResult = { pngBase64: '', width: 0, height: 0, snapshot: '' };

	if (opts.snapshot) {
		res.snapshot = ariaSnapshot(target);
	}

	if (opts.image) {
		const el = target as HTMLElement;
		const rect = el.getBoundingClientRect();
		const w = Math.max(1, Math.ceil(Math.max(rect.width, el === document.body ? window.innerWidth : 0)));
		const h = Math.max(1, Math.ceil(Math.max(rect.height, el === document.body ? window.innerHeight : 0)));
		const maxEdge = opts.maxEdge && opts.maxEdge > 0 ? opts.maxEdge : DEFAULT_MAX_EDGE;
		const scale = Math.min(window.devicePixelRatio || 1, maxEdge / Math.max(w, h));

		const canvas = await domToCanvas(el, {
			width: w,
			height: h,
			scale,
			backgroundColor: backgroundOf(el),
		});

		const dataUrl = canvas.toDataURL('image/png');
		res.pngBase64 = dataUrl.substring(dataUrl.indexOf(',') + 1);
		res.width = canvas.width;
		res.height = canvas.height;
	}

	return res;
}

/**
 * settle waits until pending vue updates, layout, fonts and images are done, so that a capture requested right
 * after a render shows that render.
 */
async function settle(): Promise<void> {
	await nextTick();
	await new Promise((r) => requestAnimationFrame(() => requestAnimationFrame(r)));
	try {
		await document.fonts?.ready;
	} catch {
		// ignore
	}

	const pending = Array.from(document.images).filter((img) => !img.complete);
	if (pending.length > 0) {
		await Promise.race([
			Promise.all(pending.map((img) => new Promise((r) => img.addEventListener('load', r, { once: true })))),
			new Promise((r) => setTimeout(r, 2000)),
		]);
	}
}

function backgroundOf(el: Element): string | undefined {
	for (let e: Element | null = el; e; e = e.parentElement) {
		const bg = getComputedStyle(e).backgroundColor;
		if (bg && bg !== 'transparent' && bg !== 'rgba(0, 0, 0, 0)') {
			return bg;
		}
	}
	const html = getComputedStyle(document.documentElement).backgroundColor;
	return html && html !== 'rgba(0, 0, 0, 0)' ? html : '#ffffff';
}

// ---------------------------------------------------------------------------------------------------------------
// accessibility snapshot
// ---------------------------------------------------------------------------------------------------------------

/**
 * ariaSnapshot renders an indented YAML-like tree of the accessible structure below root, similar to the aria
 * snapshot of Playwright. Elements without a semantic role are flattened into their parent, so the result
 * contains what a user perceives and not how the DOM is nested.
 */
export function ariaSnapshot(root: Element): string {
	const lines: string[] = [];
	walkChildren(root, 0, lines);
	return lines.join('\n');
}

// roles whose accessible name is computed from their content; the content is therefore not repeated as children
const nameFromContent = new Set([
	'button',
	'link',
	'heading',
	'tab',
	'menuitem',
	'menuitemcheckbox',
	'menuitemradio',
	'option',
	'cell',
	'columnheader',
	'rowheader',
	'checkbox',
	'radio',
	'switch',
	'treeitem',
	'tooltip',
]);

// roles which are emitted only when they have visible descendants
const structural = new Set([
	'list',
	'listitem',
	'navigation',
	'main',
	'banner',
	'contentinfo',
	'complementary',
	'region',
	'form',
	'group',
	'table',
	'row',
	'rowgroup',
	'dialog',
	'alertdialog',
	'menu',
	'menubar',
	'tablist',
	'tabpanel',
	'toolbar',
	'tree',
	'grid',
	'alert',
	'status',
	'radiogroup',
]);

function walkChildren(parent: Node, depth: number, out: string[]) {
	const root = parent instanceof Element ? (parent.shadowRoot ?? parent) : parent;
	for (const child of Array.from(root.childNodes)) {
		walk(child, depth, out);
	}
}

function walk(node: Node, depth: number, out: string[]) {
	if (node.nodeType === Node.TEXT_NODE) {
		const text = collapse(node.textContent ?? '');
		if (text) {
			pushText(out, depth, text);
		}
		return;
	}

	if (!(node instanceof Element) || isHidden(node)) {
		return;
	}

	const tag = node.tagName.toLowerCase();
	if (tag === 'script' || tag === 'style' || tag === 'template' || tag === 'noscript') {
		return;
	}

	if (tag === 'svg') {
		const name = accessibleName(node, 'img');
		if (name) {
			out.push(`${indent(depth)}- img ${quote(name)}`);
		}
		return;
	}

	const role = roleOf(node);
	if (!role || role === 'generic' || role === 'presentation' || role === 'none') {
		walkChildren(node, depth, out);
		return;
	}

	const name = accessibleName(node, role);
	let line = `${indent(depth)}- ${role}`;
	if (name) {
		line += ` ${quote(name)}`;
	}
	line += states(node, role);

	const value = valueOf(node, role);

	if (nameFromContent.has(role) || isLeafRole(role)) {
		out.push(value !== undefined ? `${line}: ${value}` : line);
		return;
	}

	const children: string[] = [];
	walkChildren(node, depth + 1, children);

	if (children.length === 0) {
		if (structural.has(role) && !name) {
			return;
		}
		out.push(value !== undefined ? `${line}: ${value}` : line);
		return;
	}

	// a single text child is inlined, as Playwright does
	const onlyText = children.length === 1 && children[0].trimStart().startsWith('- text: ');
	if (onlyText && !name) {
		out.push(`${line}: ${children[0].trimStart().substring('- text: '.length)}`);
		return;
	}

	out.push(`${line}:`);
	out.push(...children);
}

function pushText(out: string[], depth: number, text: string) {
	// merge adjacent text fragments, which are common for inline formatting
	const prefix = `${indent(depth)}- text: `;
	const last = out[out.length - 1];
	if (last !== undefined && last.startsWith(prefix)) {
		out[out.length - 1] = `${last} ${text}`;
		return;
	}
	out.push(prefix + text);
}

function isLeafRole(role: string): boolean {
	return (
		role === 'textbox' ||
		role === 'searchbox' ||
		role === 'spinbutton' ||
		role === 'slider' ||
		role === 'combobox' ||
		role === 'img' ||
		role === 'progressbar' ||
		role === 'separator' ||
		role === 'meter'
	);
}

function isHidden(el: Element): boolean {
	if (el.getAttribute('aria-hidden') === 'true' || (el as HTMLElement).hidden) {
		return true;
	}

	const anyEl = el as any;
	if (typeof anyEl.checkVisibility === 'function') {
		// display:contents elements report false but their children are visible
		const style = getComputedStyle(el);
		if (style.display === 'contents') {
			return false;
		}
		return !anyEl.checkVisibility({ visibilityProperty: true, opacityProperty: false });
	}

	const style = getComputedStyle(el);
	return style.display === 'none' || style.visibility === 'hidden';
}

function roleOf(el: Element): string | undefined {
	const explicit = el.getAttribute('role');
	if (explicit) {
		return explicit.split(/\s+/)[0];
	}

	const tag = el.tagName.toLowerCase();
	switch (tag) {
		case 'a':
			return el.hasAttribute('href') ? 'link' : undefined;
		case 'button':
			return 'button';
		case 'h1':
		case 'h2':
		case 'h3':
		case 'h4':
		case 'h5':
		case 'h6':
			return 'heading';
		case 'input':
			return inputRole(el as HTMLInputElement);
		case 'textarea':
			return 'textbox';
		case 'select':
			return (el as HTMLSelectElement).multiple || (el as HTMLSelectElement).size > 1 ? 'listbox' : 'combobox';
		case 'option':
			return 'option';
		case 'img':
			return (el as HTMLImageElement).alt === '' && el.hasAttribute('alt') ? 'presentation' : 'img';
		case 'ul':
		case 'ol':
			return 'list';
		case 'li':
			return 'listitem';
		case 'table':
			return 'table';
		case 'tr':
			return 'row';
		case 'td':
			return 'cell';
		case 'th':
			return 'columnheader';
		case 'nav':
			return 'navigation';
		case 'main':
			return 'main';
		case 'dialog':
			return 'dialog';
		case 'hr':
			return 'separator';
		case 'progress':
			return 'progressbar';
		case 'meter':
			return 'meter';
		case 'p':
			return 'paragraph';
		case 'form':
			return el.hasAttribute('aria-label') || el.hasAttribute('aria-labelledby') ? 'form' : undefined;
		default:
			return undefined;
	}
}

function inputRole(el: HTMLInputElement): string | undefined {
	switch (el.type) {
		case 'hidden':
			return undefined;
		case 'checkbox':
			return 'checkbox';
		case 'radio':
			return 'radio';
		case 'range':
			return 'slider';
		case 'number':
			return 'spinbutton';
		case 'search':
			return 'searchbox';
		case 'button':
		case 'submit':
		case 'reset':
		case 'image':
			return 'button';
		case 'file':
			return 'button';
		default:
			return 'textbox';
	}
}

function accessibleName(el: Element, role: string): string {
	const labelledBy = el.getAttribute('aria-labelledby');
	if (labelledBy) {
		const text = labelledBy
			.split(/\s+/)
			.map((id) => document.getElementById(id))
			.filter((e): e is HTMLElement => !!e)
			.map((e) => collapse(e.textContent ?? ''))
			.join(' ')
			.trim();
		if (text) return text;
	}

	const ariaLabel = el.getAttribute('aria-label');
	if (ariaLabel && ariaLabel.trim()) {
		return collapse(ariaLabel);
	}

	if (el instanceof HTMLInputElement || el instanceof HTMLTextAreaElement || el instanceof HTMLSelectElement) {
		const labels = el.labels ? Array.from(el.labels) : [];
		const text = labels
			.map((l) => collapse(visibleText(l)))
			.join(' ')
			.trim();
		if (text) return text;
		if (el instanceof HTMLInputElement && ['button', 'submit', 'reset'].includes(el.type) && el.value) {
			return collapse(el.value);
		}
		const placeholder = el.getAttribute('placeholder');
		if (placeholder) return collapse(placeholder);
	}

	if (el instanceof HTMLImageElement) {
		return collapse(el.alt ?? '');
	}

	if (el.tagName.toLowerCase() === 'svg') {
		const title = el.querySelector('title');
		if (title?.textContent) return collapse(title.textContent);
	}

	if (nameFromContent.has(role)) {
		const text = collapse(visibleText(el));
		if (text) return truncate(text, 200);
	}

	const title = el.getAttribute('title');
	if (title) return collapse(title);

	return '';
}

/**
 * visibleText concatenates the text below el, skipping hidden subtrees and form controls.
 */
function visibleText(el: Element): string {
	let s = '';
	for (const child of Array.from(el.childNodes)) {
		if (child.nodeType === Node.TEXT_NODE) {
			s += ' ' + (child.textContent ?? '');
		} else if (child instanceof Element && !isHidden(child)) {
			const tag = child.tagName.toLowerCase();
			if (tag === 'input' || tag === 'textarea' || tag === 'select' || tag === 'script' || tag === 'style') {
				continue;
			}
			if (tag === 'img') {
				s += ' ' + ((child as HTMLImageElement).alt ?? '');
				continue;
			}
			const label = child.getAttribute('aria-label');
			s += ' ' + (label ?? visibleText(child));
		}
	}
	return s;
}

function states(el: Element, role: string): string {
	const s: string[] = [];

	if (role === 'heading') {
		const level = el.getAttribute('aria-level') ?? el.tagName.match(/^H([1-6])$/i)?.[1];
		if (level) s.push(`level=${level}`);
	}

	if (role === 'checkbox' || role === 'radio' || role === 'switch' || role === 'menuitemcheckbox') {
		const aria = el.getAttribute('aria-checked');
		if (aria === 'mixed') s.push('checked=mixed');
		else if (aria === 'true' || (el instanceof HTMLInputElement && el.checked)) s.push('checked');
	}

	if (el.getAttribute('aria-pressed') === 'true') s.push('pressed');
	if (el.getAttribute('aria-selected') === 'true' || (el instanceof HTMLOptionElement && el.selected))
		s.push('selected');

	const expanded = el.getAttribute('aria-expanded');
	if (expanded === 'true') s.push('expanded');
	else if (expanded === 'false') s.push('expanded=false');

	if (el.getAttribute('aria-disabled') === 'true' || (el as HTMLButtonElement).disabled) s.push('disabled');
	if (el.getAttribute('aria-invalid') === 'true') s.push('invalid');
	if (el.getAttribute('aria-required') === 'true' || (el as HTMLInputElement).required) s.push('required');
	if ((el as HTMLInputElement).readOnly) s.push('readonly');
	if (el === document.activeElement && el !== document.body) s.push('focused');

	return s.length > 0 ? ` [${s.join('] [')}]` : '';
}

function valueOf(el: Element, role: string): string | undefined {
	if (el instanceof HTMLInputElement) {
		if (el.type === 'password') {
			return el.value ? '"••••"' : undefined;
		}
		if (['textbox', 'searchbox', 'spinbutton', 'slider'].includes(role)) {
			return el.value ? quote(truncate(el.value, 500)) : undefined;
		}
		return undefined;
	}

	if (el instanceof HTMLTextAreaElement) {
		return el.value ? quote(truncate(el.value, 500)) : undefined;
	}

	if (el instanceof HTMLSelectElement && role === 'combobox') {
		const opt = el.selectedOptions[0];
		return opt ? quote(collapse(opt.textContent ?? '')) : undefined;
	}

	if (role === 'progressbar' || role === 'slider' || role === 'meter' || role === 'spinbutton') {
		const now = el.getAttribute('aria-valuetext') ?? el.getAttribute('aria-valuenow');
		return now ? quote(now) : undefined;
	}

	return undefined;
}

function collapse(s: string): string {
	return s.replace(/\s+/g, ' ').trim();
}

function truncate(s: string, max: number): string {
	return s.length > max ? s.substring(0, max) + '…' : s;
}

function quote(s: string): string {
	return JSON.stringify(s);
}

function indent(depth: number): string {
	return '  '.repeat(depth);
}

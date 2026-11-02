/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

const svgNS = 'http://www.w3.org/2000/svg';

export function downloadSVG(svg: SVGSVGElement, name: string) {
	download(new Blob([serialize(svg)], { type: 'image/svg+xml;charset=utf-8' }), name + '.svg');
}

// downloadPNG rasterizes the chart at twice its size.
export async function downloadPNG(svg: SVGSVGElement, name: string) {
	const width = svg.clientWidth;
	const height = svg.clientHeight;
	const url = URL.createObjectURL(new Blob([serialize(svg)], { type: 'image/svg+xml;charset=utf-8' }));
	try {
		const img = new Image();
		await new Promise((resolve, reject) => {
			img.onload = resolve;
			img.onerror = reject;
			img.src = url;
		});

		const scale = 2;
		const canvas = document.createElement('canvas');
		canvas.width = width * scale;
		canvas.height = height * scale;
		const ctx = canvas.getContext('2d');
		if (!ctx) {
			return;
		}

		ctx.scale(scale, scale);
		ctx.drawImage(img, 0, 0, width, height);
		const blob = await new Promise<Blob | null>((resolve) => canvas.toBlob(resolve, 'image/png'));
		if (blob) {
			download(blob, name + '.png');
		}
	} finally {
		URL.revokeObjectURL(url);
	}
}

export function downloadCSV(rows: string[][], name: string) {
	const field = (s: string) => (/[",\r\n]/.test(s) ? `"${s.replace(/"/g, '""')}"` : s);
	const text = rows.map((row) => row.map(field).join(',')).join('\r\n');
	// the byte order mark makes spreadsheets read the file as UTF-8
	download(new Blob(['﻿' + text], { type: 'text/csv;charset=utf-8' }), name + '.csv');
}

function download(blob: Blob, filename: string) {
	const url = URL.createObjectURL(blob);
	const a = document.createElement('a');
	a.href = url;
	a.download = filename;
	document.body.appendChild(a);
	a.click();
	a.remove();
	setTimeout(() => URL.revokeObjectURL(url), 1000);
}

// serialize returns the markup of the chart as a standalone image: the colors of the theme and the text color it
// inherits are resolved, the inspection is left out and the background of the page is put behind it.
function serialize(svg: SVGSVGElement): string {
	const clone = svg.cloneNode(true) as SVGSVGElement;
	const from = [svg, ...svg.querySelectorAll('*')];
	const to = [clone, ...clone.querySelectorAll('*')];
	from.forEach((el, i) => {
		const style = getComputedStyle(el);
		for (const attr of ['fill', 'stroke', 'stop-color']) {
			const v = to[i].getAttribute(attr);
			if (v && (v === 'currentColor' || v.includes('var(') || v.includes('color-mix('))) {
				to[i].setAttribute(attr, rgb(style.getPropertyValue(attr)));
			}
		}
	});

	clone.querySelectorAll('[data-export="false"]').forEach((el) => el.remove());
	const comments = document.createTreeWalker(clone, NodeFilter.SHOW_COMMENT);
	const remove: Node[] = [];
	while (comments.nextNode()) {
		remove.push(comments.currentNode);
	}

	remove.forEach((n) => n.parentNode?.removeChild(n));

	const style = getComputedStyle(svg);
	clone.setAttribute('xmlns', svgNS);
	clone.setAttribute('width', String(svg.clientWidth));
	clone.setAttribute('height', String(svg.clientHeight));
	clone.setAttribute('font-family', style.fontFamily);
	clone.removeAttribute('class');
	clone.removeAttribute('style');

	const bg = document.createElementNS(svgNS, 'rect');
	bg.setAttribute('width', '100%');
	bg.setAttribute('height', '100%');
	bg.setAttribute('fill', background(svg));
	clone.insertBefore(bg, clone.firstChild);

	return new XMLSerializer().serializeToString(clone);
}

// background returns the first opaque background behind the element.
function background(el: Element): string {
	for (let e: Element | null = el; e; e = e.parentElement) {
		const c = rgb(getComputedStyle(e).backgroundColor);
		if (c && c !== 'transparent' && !/^rgba\(.*,\s*0\)$/.test(c)) {
			return c;
		}
	}

	return '#ffffff';
}

// rgb converts a color in the srgb color space, which browsers resolve theme colors and mixes to, into the rgb
// notation, which every program understands that opens an SVG.
function rgb(c: string): string {
	const m = /^color\(srgb ([-\d.e]+) ([-\d.e]+) ([-\d.e]+)(?: \/ ([-\d.e]+))?\)$/.exec(c.trim());
	if (!m) {
		return c;
	}

	const [r, g, b] = [m[1], m[2], m[3]].map((v) => Math.round(Math.min(Math.max(parseFloat(v), 0), 1) * 255));
	if (m[4] === undefined) {
		return `rgb(${r}, ${g}, ${b})`;
	}

	return `rgba(${r}, ${g}, ${b}, ${parseFloat(m[4])})`;
}

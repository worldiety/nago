/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { Hct } from '@material/material-color-utilities/hct/hct.js';

// RGB holds the channels in the range of 0 to 255.
export interface RGB {
	r: number;
	g: number;
	b: number;
}

// HCT holds hue (0-360), chroma (0-~150) and tone (0-100) of the HCT color space. The server calculates it with
// a Go port of the same library, see presentation/ui/colorpicker/testdata/hct_ts.json.
export interface HCT {
	h: number;
	c: number;
	t: number;
}

// HSV holds hue (0-360), saturation (0-100) and value (0-100).
export interface HSV {
	h: number;
	s: number;
	v: number;
}

// HSL holds hue (0-360), saturation (0-100) and lightness (0-100).
export interface HSL {
	h: number;
	s: number;
	l: number;
}

// CMYK holds the naive device independent conversion (0-100), without any color profile.
export interface CMYK {
	c: number;
	m: number;
	y: number;
	k: number;
}

const clamp = (v: number, min: number, max: number) => Math.min(max, Math.max(min, v));

export function wrapHue(h: number): number {
	return ((h % 360) + 360) % 360;
}

// parseHex accepts #rgb, #rgba, #rrggbb and #rrggbbaa and returns the color and its alpha (0-1).
export function parseHex(hex: string | undefined): { rgb: RGB; alpha: number } | undefined {
	if (!hex) return undefined;
	let v = hex.trim();
	if (!v.startsWith('#')) return undefined;
	v = v.substring(1);
	if (!/^[0-9a-fA-F]+$/.test(v)) return undefined;
	if (v.length === 3 || v.length === 4) {
		v = v
			.split('')
			.map((c) => c + c)
			.join('');
	}

	if (v.length !== 6 && v.length !== 8) return undefined;
	const n = (i: number) => parseInt(v.substring(i, i + 2), 16);
	return { rgb: { r: n(0), g: n(2), b: n(4) }, alpha: v.length === 8 ? n(6) / 255 : 1 };
}

// toHex returns #rrggbb for an opaque color and #rrggbbaa otherwise, like the server normalizes it.
export function toHex(rgb: RGB, alpha = 1): string {
	const h = (v: number) =>
		Math.round(clamp(v, 0, 255))
			.toString(16)
			.padStart(2, '0');
	const a = Math.round(clamp(alpha, 0, 1) * 255);
	return `#${h(rgb.r)}${h(rgb.g)}${h(rgb.b)}${a === 255 ? '' : h(a)}`;
}

function argbOf(rgb: RGB): number {
	return ((0xff << 24) | (Math.round(rgb.r) << 16) | (Math.round(rgb.g) << 8) | Math.round(rgb.b)) >>> 0;
}

function rgbOfArgb(argb: number): RGB {
	return { r: (argb >> 16) & 0xff, g: (argb >> 8) & 0xff, b: argb & 0xff };
}

export function rgbToHct(rgb: RGB): HCT {
	const v = Hct.fromInt(argbOf(rgb));
	return { h: v.hue, c: v.chroma, t: v.tone };
}

// hctToRgb returns the color with the given hue and tone and the chroma, which comes closest to the requested one
// within sRGB.
export function hctToRgb(hct: HCT): RGB {
	return rgbOfArgb(Hct.from(hct.h, hct.c, hct.t).toInt());
}

// maxChroma returns the highest chroma, which can be displayed in sRGB for the given hue and tone.
export function maxChroma(h: number, t: number): number {
	return Hct.from(h, 200, t).chroma;
}

export function rgbToHsv({ r, g, b }: RGB): HSV {
	r /= 255;
	g /= 255;
	b /= 255;
	const max = Math.max(r, g, b);
	const d = max - Math.min(r, g, b);
	return { h: hueOf(r, g, b, max, d), s: max === 0 ? 0 : (d / max) * 100, v: max * 100 };
}

export function hsvToRgb({ h, s, v }: HSV): RGB {
	s /= 100;
	v /= 100;
	const f = (n: number) => {
		const k = (n + h / 60) % 6;
		return (v - v * s * Math.max(0, Math.min(k, 4 - k, 1))) * 255;
	};
	return { r: f(5), g: f(3), b: f(1) };
}

export function rgbToHsl({ r, g, b }: RGB): HSL {
	r /= 255;
	g /= 255;
	b /= 255;
	const max = Math.max(r, g, b);
	const min = Math.min(r, g, b);
	const d = max - min;
	const l = (max + min) / 2;
	const s = d === 0 ? 0 : d / (1 - Math.abs(2 * l - 1));
	return { h: hueOf(r, g, b, max, d), s: s * 100, l: l * 100 };
}

export function hslToRgb({ h, s, l }: HSL): RGB {
	s /= 100;
	l /= 100;
	const a = s * Math.min(l, 1 - l);
	const f = (n: number) => {
		const k = (n + h / 30) % 12;
		return (l - a * Math.max(-1, Math.min(k - 3, 9 - k, 1))) * 255;
	};
	return { r: f(0), g: f(8), b: f(4) };
}

export function rgbToCmyk({ r, g, b }: RGB): CMYK {
	r /= 255;
	g /= 255;
	b /= 255;
	const k = 1 - Math.max(r, g, b);
	if (k === 1) return { c: 0, m: 0, y: 0, k: 100 };
	return {
		c: ((1 - r - k) / (1 - k)) * 100,
		m: ((1 - g - k) / (1 - k)) * 100,
		y: ((1 - b - k) / (1 - k)) * 100,
		k: k * 100,
	};
}

export function cmykToRgb({ c, m, y, k }: CMYK): RGB {
	const f = (v: number) => 255 * (1 - v / 100) * (1 - k / 100);
	return { r: f(c), g: f(m), b: f(y) };
}

function hueOf(r: number, g: number, b: number, max: number, d: number): number {
	if (d === 0) return 0;
	let h: number;
	if (max === r) h = ((g - b) / d) % 6;
	else if (max === g) h = (b - r) / d + 2;
	else h = (r - g) / d + 4;
	return wrapHue(h * 60);
}

// parseNumbers reads the numbers of a notation like "4°, 82%, 56%" and ignores units and separators.
export function parseNumbers(text: string, count: number): number[] | undefined {
	const numbers = text.match(/-?\d+(?:\.\d+)?/g)?.map((n) => parseFloat(n));
	if (!numbers || numbers.length !== count || numbers.some((n) => !isFinite(n))) return undefined;
	return numbers;
}

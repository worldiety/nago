/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

// Axis maps the indices of rows or columns to positions. Most indices have the uniform size, only a few have
// their own size or are hidden, which is a size of zero. So a sheet with tens of thousands of rows needs no table
// of positions: a position is the uniform one plus the differences of the exceptions before it.
export class Axis {
	readonly count: number;
	private readonly uniform: number;
	private readonly indices: number[]; // sorted indices of the exceptions
	private readonly sizes: number[]; // size of each exception
	private readonly deltas: number[]; // sum of (size - uniform) of the exceptions up to and including i
	private readonly hidden: Set<number>;

	constructor(count: number, uniform: number, sizes: Map<number, number>, hidden: number[]) {
		this.count = count;
		this.uniform = uniform;
		this.hidden = new Set(hidden.filter((h) => h >= 0 && h < count));

		const exceptions = new Map<number, number>();
		for (const [i, size] of sizes) {
			if (i >= 0 && i < count && size !== uniform) exceptions.set(i, size);
		}
		for (const h of this.hidden) exceptions.set(h, 0);

		this.indices = [...exceptions.keys()].sort((a, b) => a - b);
		this.sizes = this.indices.map((i) => exceptions.get(i)!);
		this.deltas = [];
		let sum = 0;
		for (const size of this.sizes) {
			sum += size - uniform;
			this.deltas.push(sum);
		}
	}

	isHidden(index: number): boolean {
		return this.hidden.has(index);
	}

	size(index: number): number {
		const k = lowerBound(this.indices, index);
		return this.indices[k] === index ? this.sizes[k] : this.uniform;
	}

	// start is the position of the index, which is the end of all indices before it
	start(index: number): number {
		const k = lowerBound(this.indices, index);
		return index * this.uniform + (k > 0 ? this.deltas[k - 1] : 0);
	}

	get total(): number {
		return this.start(this.count);
	}

	// span is the size of the indices [from, from+count).
	span(from: number, count: number): number {
		return this.start(Math.min(this.count, from + count)) - this.start(from);
	}

	// indexAtPosition returns the visible index at the position, or the nearest one.
	indexAtPosition(pos: number): number {
		let lo = 0;
		let hi = this.count - 1;
		let res = 0;
		while (lo <= hi) {
			const mid = (lo + hi) >> 1;
			if (this.start(mid) <= pos) {
				res = mid;
				lo = mid + 1;
			} else {
				hi = mid - 1;
			}
		}

		// a hidden index has the start of the next one, so step back to a visible one at the end
		while (res > 0 && this.size(res) === 0) res--;
		while (res < this.count - 1 && this.size(res) === 0) res++;
		return res;
	}

	// next returns the next visible index in the direction or the given one, if there is none.
	next(index: number, step: number): number {
		let i = index;
		let moved = 0;
		const dir = Math.sign(step);
		while (moved < Math.abs(step)) {
			const candidate = i + dir;
			if (candidate < 0 || candidate >= this.count) break;
			i = candidate;
			if (!this.isHidden(i)) moved++;
		}
		while (this.isHidden(i) && i > 0) i--;
		return i;
	}

	first(): number {
		let i = 0;
		while (i < this.count - 1 && this.isHidden(i)) i++;
		return i;
	}

	last(): number {
		let i = this.count - 1;
		while (i > 0 && this.isHidden(i)) i--;
		return Math.max(0, i);
	}
}

function lowerBound(sorted: number[], value: number): number {
	let lo = 0;
	let hi = sorted.length;
	while (lo < hi) {
		const mid = (lo + hi) >> 1;
		if (sorted[mid] < value) lo = mid + 1;
		else hi = mid;
	}
	return lo;
}

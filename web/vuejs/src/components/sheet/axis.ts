/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */

// Axis maps the indices of rows or columns to positions, skipping hidden ones. Rows have a uniform size, so a sheet
// with tens of thousands of rows needs no table of positions; columns are few and may have individual sizes.
export class Axis {
	readonly count: number;
	private readonly hidden: number[]; // sorted
	private readonly hiddenSet: Set<number>;
	private readonly sizes?: number[]; // per index, only for columns
	private readonly uniform: number;
	private readonly offsets?: number[]; // start of the visible ordinal, only with sizes

	constructor(count: number, hidden: number[], uniform: number, sizes?: Map<number, number>) {
		this.count = count;
		this.hidden = [...hidden].filter((h) => h >= 0 && h < count).sort((a, b) => a - b);
		this.hiddenSet = new Set(this.hidden);
		this.uniform = uniform;
		if (sizes) {
			this.sizes = [];
			this.offsets = [0];
			for (let i = 0; i < count; i++) {
				const size = sizes.get(i) ?? uniform;
				this.sizes.push(size);
				if (!this.hiddenSet.has(i)) {
					this.offsets.push(this.offsets[this.offsets.length - 1] + size);
				}
			}
		}
	}

	isHidden(index: number): boolean {
		return this.hiddenSet.has(index);
	}

	get visibleCount(): number {
		return this.count - this.hidden.length;
	}

	// ordinal returns the position of the index among the visible ones.
	ordinal(index: number): number {
		return index - lowerBound(this.hidden, index);
	}

	// indexAt is the inverse of ordinal.
	indexAt(ordinal: number): number {
		let index = ordinal;
		for (const h of this.hidden) {
			if (h <= index) index++;
			else break;
		}
		return Math.min(index, this.count);
	}

	size(index: number): number {
		if (this.isHidden(index)) return 0;
		return this.sizes ? this.sizes[index] : this.uniform;
	}

	start(index: number): number {
		const ord = this.ordinal(index);
		return this.offsets ? this.offsets[ord] : ord * this.uniform;
	}

	get total(): number {
		return this.offsets ? this.offsets[this.offsets.length - 1] : this.visibleCount * this.uniform;
	}

	// span is the size of the visible indices of [from, from+count).
	span(from: number, count: number): number {
		let res = 0;
		const to = Math.min(this.count, from + count);
		if (!this.sizes) {
			return (this.ordinal(to) - this.ordinal(from)) * this.uniform;
		}
		for (let i = from; i < to; i++) res += this.size(i);
		return res;
	}

	// indexAtPosition returns the visible index at the position.
	indexAtPosition(pos: number): number {
		if (this.visibleCount === 0) return 0;
		let ord: number;
		if (this.offsets) {
			ord = Math.max(0, upperBound(this.offsets, pos) - 1);
		} else {
			ord = Math.floor(pos / this.uniform);
		}
		ord = Math.max(0, Math.min(ord, this.visibleCount - 1));
		return this.indexAt(ord);
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
		return this.indexAt(0);
	}

	last(): number {
		return this.indexAt(Math.max(0, this.visibleCount - 1));
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

function upperBound(sorted: number[], value: number): number {
	let lo = 0;
	let hi = sorted.length;
	while (lo < hi) {
		const mid = (lo + hi) >> 1;
		if (sorted[mid] <= value) lo = mid + 1;
		else hi = mid;
	}
	return lo;
}

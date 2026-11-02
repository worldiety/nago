/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import {
	type Label,
	type LegendEntry,
	type PlacedLegendEntry,
	type Rect,
	TextMeasurer,
	type Tick,
	flowLegend,
	labelHeight,
	niceDomain,
	numberTicks,
	pad,
	titleHeight,
	titleSize,
	toolbarSize,
} from '@/components/charts/chart';

export interface CartesianOptions {
	width: number;
	height: number;
	categories: string[];
	// lo and hi are the extent of the values.
	lo: number;
	hi: number;
	// includeZero keeps zero on the value axis, as bars and areas need it.
	includeZero: boolean;
	// horizontal puts the categories on the vertical axis and the values on the horizontal one.
	horizontal: boolean;
	xTitle?: string;
	yTitle?: string;
	legend: LegendEntry[];
	toolbar: boolean;
	measure: TextMeasurer;
}

export interface PositionedTick extends Tick {
	pos: number;
}

// Cartesian is the layout of a chart with a category axis and a value axis: the plot area, the ticks, the labels
// and titles around it and the legend below it. A vertical chart has its value axis on the trailing side, as
// Swift Charts has, a horizontal one its categories on the leading side.
export class Cartesian {
	readonly plot: Rect;
	readonly min: number;
	readonly max: number;
	readonly ticks: PositionedTick[];
	readonly tickLabels: Label[] = [];
	readonly categoryLabels: Label[] = [];
	readonly titles: Label[] = [];
	readonly legend: PlacedLegendEntry[];
	readonly bandSize: number;
	readonly horizontal: boolean;
	private readonly count: number;

	constructor(o: CartesianOptions) {
		const { width, height, measure } = o;
		this.horizontal = o.horizontal;
		this.count = Math.max(o.categories.length, 1);

		const legend = o.legend.length > 1 ? flowLegend(o.legend, width, measure) : { items: [], height: 0 };
		const header = Math.max(o.yTitle ? titleHeight : 0, o.toolbar ? toolbarSize : 0);
		const top = header > 0 ? header + 4 + (this.horizontal ? 0 : labelHeight / 2) : labelHeight / 2 + 2;
		let bottom = height - labelHeight - pad / 2;
		if (o.xTitle) {
			bottom -= titleHeight + 2;
		}

		if (legend.height > 0) {
			bottom -= legend.height + pad * 2;
		}

		bottom = Math.max(bottom, top + 1);

		const catWidths = o.categories.map((c) => measure.width(c));
		const widest = Math.max(0, ...catWidths);

		let left: number;
		let right: number;
		let lo = o.lo;
		let hi = o.hi;
		if (o.includeZero) {
			lo = Math.min(lo, 0);
			hi = Math.max(hi, 0);
			if (lo === 0 && hi === 0) {
				hi = 1;
			}
		}

		if (this.horizontal) {
			const catWidth = Math.min(widest, width * 0.3);
			left = catWidth + pad;
			right = Math.max(width - 8, left + 1);
			const count = Math.max(1, Math.floor((right - left) / 80));
			[this.min, this.max] = niceDomain(lo, hi, count);
			this.ticks = this.position(numberTicks(this.min, this.max, count), {
				x: left,
				y: top,
				w: right - left,
				h: bottom - top,
			});
		} else {
			const count = Math.max(1, Math.floor((bottom - top) / 36));
			[this.min, this.max] = niceDomain(lo, hi, count);
			const ticks = numberTicks(this.min, this.max, count);
			const tickWidth = Math.max(0, ...ticks.map((t) => measure.width(t.label)));
			left = 0;
			right = Math.max(width - tickWidth - pad, left + 1);
			this.ticks = this.position(ticks, { x: left, y: top, w: right - left, h: bottom - top });
		}

		this.plot = { x: left, y: top, w: right - left, h: bottom - top };
		this.bandSize = (this.horizontal ? this.plot.h : this.plot.w) / this.count;
		this.legend = legend.items.map((e) => ({ ...e, y: e.y + height - legend.height }));

		this.layoutTicks(width, measure);
		this.layoutCategories(o.categories, catWidths, width, measure);
		this.layoutTitles(o, width, header);
	}

	private position(ticks: Tick[], plot: Rect): PositionedTick[] {
		const span = this.max - this.min || 1;
		return ticks.map((t) => {
			const f = (t.value - this.min) / span;
			return { ...t, pos: this.horizontal ? plot.x + f * plot.w : plot.y + plot.h - f * plot.h };
		});
	}

	// valuePos returns the position of the value on the value axis.
	valuePos(v: number): number {
		const f = (v - this.min) / (this.max - this.min || 1);
		if (this.horizontal) {
			return this.plot.x + f * this.plot.w;
		}

		return this.plot.y + this.plot.h - f * this.plot.h;
	}

	// baseline is where bars and areas start: zero, or the edge of the plot nearest to it.
	baseline(): number {
		return this.valuePos(Math.min(Math.max(0, this.min), this.max));
	}

	// bandStart returns the start of the category on the category axis.
	bandStart(i: number): number {
		return (this.horizontal ? this.plot.y : this.plot.x) + i * this.bandSize;
	}

	bandCenter(i: number): number {
		return this.bandStart(i) + this.bandSize / 2;
	}

	// categoryAt returns the category under the position, or -1 if there is none. The labels of a category belong
	// to it as well.
	categoryAt(x: number, y: number): number {
		const p = this.plot;
		let at: number;
		if (this.horizontal) {
			if (y < p.y || y > p.y + p.h || x > p.x + p.w) {
				return -1;
			}

			at = y - p.y;
		} else {
			if (x < p.x || x > p.x + p.w || y > p.y + p.h + labelHeight + pad) {
				return -1;
			}

			at = x - p.x;
		}

		return Math.min(Math.max(Math.floor(at / this.bandSize), 0), this.count - 1);
	}

	private layoutTicks(width: number, measure: TextMeasurer) {
		const p = this.plot;
		let edge = -Infinity;
		for (const t of this.ticks) {
			if (this.horizontal) {
				// centered under the tick, kept within the chart and left out where it runs into the one before
				const w = measure.width(t.label);
				const x = Math.min(Math.max(t.pos, w / 2), width - w / 2);
				if (x - w / 2 < edge + pad) {
					continue;
				}

				edge = x + w / 2;
				this.tickLabels.push({ x, y: p.y + p.h + pad / 2 + labelHeight / 2, text: t.label, anchor: 'middle' });
			} else {
				this.tickLabels.push({ x: p.x + p.w + pad, y: t.pos, text: t.label, anchor: 'start' });
			}
		}
	}

	private layoutCategories(cats: string[], widths: number[], width: number, measure: TextMeasurer) {
		if (cats.length === 0) {
			return;
		}

		const p = this.plot;
		if (this.horizontal) {
			const every = Math.max(1, Math.ceil(labelHeight / this.bandSize));
			const room = p.x - pad;
			cats.forEach((c, i) => {
				if (i % every === 0) {
					this.categoryLabels.push({
						x: room,
						y: this.bandCenter(i),
						text: measure.truncate(c, room),
						anchor: 'end',
					});
				}
			});

			return;
		}

		// categories take every so many bands, so that the widest of them does not run into its neighbours
		const maxWidth = Math.min(Math.max(0, ...widths), 160);
		const every = Math.max(1, Math.ceil((maxWidth + pad * 2) / this.bandSize));
		const room = this.bandSize * every - pad;
		cats.forEach((c, i) => {
			if (i % every !== 0) {
				return;
			}

			const text = measure.truncate(c, Math.max(room, 24));
			const w = measure.width(text);
			const x = Math.min(Math.max(this.bandCenter(i), w / 2), width - w / 2);
			this.categoryLabels.push({ x, y: p.y + p.h + pad / 2 + labelHeight / 2, text, anchor: 'middle' });
		});
	}

	private layoutTitles(o: CartesianOptions, width: number, header: number) {
		const p = this.plot;
		if (o.xTitle) {
			const y = p.y + p.h + pad / 2 + labelHeight + 2 + titleHeight / 2;
			this.titles.push({ x: p.x + p.w, y, text: o.measure.truncate(o.xTitle, p.w, titleSize), anchor: 'end' });
		}

		if (o.yTitle) {
			// the title of the vertical axis stands above it, beside the download button
			const room = width - (o.toolbar ? toolbarSize + pad : 0);
			const text = o.measure.truncate(o.yTitle, room, titleSize);
			if (this.horizontal) {
				this.titles.push({ x: 0, y: header / 2, text, anchor: 'start' });
			} else {
				this.titles.push({ x: room, y: header / 2, text, anchor: 'end' });
			}
		}
	}
}

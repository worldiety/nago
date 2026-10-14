/**
 * Copyright (c) 2026 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import type { SheetCell, SheetStyle } from '@/shared/proto/nprotoc_gen';
import { SheetValueTypeValues } from '@/shared/proto/nprotoc_gen';

export interface CopyCell {
	cell?: SheetCell;
	style?: SheetStyle;
	rowSpan: number;
	colSpan: number;
}

export function decimalSeparator(locale: string): string {
	try {
		return new Intl.NumberFormat(locale).formatToParts(1.5).find((p) => p.type === 'decimal')?.value ?? '.';
	} catch {
		return '.';
	}
}

function rawValue(cell: SheetCell | undefined): string {
	if (!cell) return '';
	return cell.value || cell.text || '';
}

function isNumber(cell: SheetCell | undefined): boolean {
	return !!cell?.value && cell.type === SheetValueTypeValues.SheetValueNumber;
}

// tsvValue writes numbers with the decimal separator of the locale, so that a spreadsheet application parses them.
function tsvValue(cell: SheetCell | undefined, decimal: string): string {
	let v = rawValue(cell);
	if (isNumber(cell) && decimal !== '.') v = v.replace('.', decimal);
	if (/[\t\r\n"]/.test(v)) v = '"' + v.replaceAll('"', '""') + '"';
	return v;
}

// toTSV renders the rows of cells as tab separated values. A null entry is covered by a merge and left empty.
export function toTSV(rows: (CopyCell | null)[][], decimal: string): string {
	return rows.map((row) => row.map((c) => (c ? tsvValue(c.cell, decimal) : '')).join('\t')).join('\r\n');
}

function escapeHTML(s: string): string {
	return s.replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('>', '&gt;').replaceAll('"', '&quot;');
}

// toHTML renders the rows as a table, from which a spreadsheet application takes over colors, fonts and merges.
// x:num marks a number, so that Excel keeps it as a number although the text is formatted.
export function toHTML(rows: (CopyCell | null)[][], resolveColor: (c?: string) => string): string {
	const body = rows
		.map((row) => {
			const cells = row
				.filter((c): c is CopyCell => c !== null)
				.map((c) => {
					const css: string[] = [];
					const s = c.style;
					if (s?.backgroundColor) css.push(`background:${resolveColor(s.backgroundColor)}`);
					if (s?.color) css.push(`color:${resolveColor(s.color)}`);
					if (s?.bold) css.push('font-weight:bold');
					if (s?.italic) css.push('font-style:italic');
					if (s?.underline) css.push('text-decoration:underline');
					const attrs: string[] = [];
					if (css.length) attrs.push(`style="${css.join(';')}"`);
					if (c.rowSpan > 1) attrs.push(`rowspan="${c.rowSpan}"`);
					if (c.colSpan > 1) attrs.push(`colspan="${c.colSpan}"`);
					if (isNumber(c.cell)) attrs.push(`x:num="${escapeHTML(c.cell!.value!)}"`);
					return `<td ${attrs.join(' ')}>${escapeHTML(c.cell?.text ?? '')}</td>`;
				})
				.join('');
			return `<tr>${cells}</tr>`;
		})
		.join('');

	return `<html xmlns:x="urn:schemas-microsoft-com:office:excel"><body><table>${body}</table></body></html>`;
}

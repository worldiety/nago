<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { borderCSS } from '@/components/shared/border';
import { colorValue } from '@/components/shared/colors';
import { frameCSS } from '@/components/shared/frame';
import { cssLengthValue } from '@/components/shared/length';
import { paddingCSS } from '@/components/shared/padding';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { AlignmentValues, FunctionCallRequested, Table } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Table;
}>();

const serviceAdapter = useServiceAdapter();

function commonStyles(): string[] {
	let styles = frameCSS(props.ui.frame);

	// background handling
	if (props.ui.backgroundColor) {
		styles.push(`background-color: ${colorValue(props.ui.backgroundColor)}`);
	}

	// border handling
	styles.push(...borderCSS(props.ui.border));

	return styles;
}

const frameStyles = computed<string>(() => {
	let styles = commonStyles();

	return styles.join(';');
});

function rowStyles(idx: number): string {
	const styles: string[] = [];
	let row = props.ui.rows?.value?.at(idx)!;
	if (row.backgroundColor) {
		styles.push(`background-color: ${colorValue(row.backgroundColor)}`);
	}

	if (row.hovered && row.hoveredBackgroundColor) {
		styles.push(`background-color: ${colorValue(row.hoveredBackgroundColor)}`);
	}

	if (row.height) {
		styles.push(`height: ${cssLengthValue(row.height)}`);
	}

	if (idx > 0 && props.ui.rowDividerColor) {
		styles.push(
			'border-collapse: collapse',
			'border-top-width: 1px',
			`border-color: ${colorValue(props.ui.rowDividerColor)}`
		);
	}

	if (row.action) {
		styles.push('cursor: pointer');
	}

	return styles.join(';');
}

function headStyles() {
	const styles: string[] = [];
	if (props.ui.headerDividerColor) {
		styles.push(
			'border-collapse: collapse',
			'border-bottom-width: 2px',
			`border-color: ${colorValue(props.ui.headerDividerColor)}`
		);
	} else if (props.ui.rowDividerColor) {
		styles.push(
			'border-collapse: collapse',
			'border-bottom-width: 2px',
			`border-color: ${colorValue(props.ui.rowDividerColor)}`
		);
	}

	return styles.join(';');
}

function cellStyles(rowIdx: number, colIdx: number): string {
	const styles: string[] = [];
	let cell = props.ui.rows?.value.at(rowIdx)?.cells?.value.at(colIdx)!;
	if (cell.backgroundColor) {
		styles.push(`background-color: ${colorValue(cell.backgroundColor)}`);
	}

	if (cell.hovered && cell.hoveredBackgroundColor) {
		styles.push(`background-color: ${colorValue(cell.hoveredBackgroundColor)}`);
	}

	if (cell.alignment === undefined) {
		cell.alignment = AlignmentValues.Leading; // this is a different rule: if not defined, use leading, which is more common than centered in tables
	}

	switch (cell.alignment) {
		case AlignmentValues.Leading:
			styles.push('vertical-align: middle', 'text-align: start');
			break;
		case AlignmentValues.Trailing:
			styles.push('vertical-align: middle', 'text-align: end');
			break;
		case AlignmentValues.Center:
			styles.push('vertical-align: middle', 'text-align: center');
			break;
		case AlignmentValues.TopLeading:
			styles.push('vertical-align: top', 'text-align: start');
			break;
		case AlignmentValues.BottomLeading:
			styles.push('vertical-align: bottom', 'text-align: start');
			break;
		case AlignmentValues.TopTrailing:
			styles.push('vertical-align: top', 'text-align: end');
			break;
		case AlignmentValues.Top:
			styles.push('vertical-align: top', 'text-align: center');
			break;
		case AlignmentValues.BottomTrailing:
			styles.push('vertical-align: bottom', 'text-align: end');
			break;
		case AlignmentValues.Bottom:
			styles.push('vertical-align: bottom', 'text-align: center');
			break;
		default:
			// nothing, just default
			break;
	}

	if (!cell.padding) {
		// default cell padding from the entire table
		styles.push(...paddingCSS(props.ui.defaultCellPadding));
	} else {
		// specific cell padding takes precedence
		styles.push(...paddingCSS(cell.padding));
	}

	styles.push(...borderCSS(cell.border));

	if (cell.action) {
		styles.push('cursor: pointer');
	}

	return styles.join(';');
}

function headCellStyles(colIdx: number): string {
	const styles: string[] = [];
	let cell = props.ui.header?.columns?.value.at(colIdx)!;
	if (cell.cellBackgroundColor) {
		styles.push(`background-color: ${colorValue(cell.cellBackgroundColor)}`);
	}

	if (cell.cellHovered && cell.cellHoveredBackgroundColor) {
		styles.push(`background-color: ${colorValue(cell.cellHoveredBackgroundColor)}`);
	}

	if (cell.width) {
		styles.push(`width: ${cssLengthValue(cell.width)}`);
	}

	if (cell.alignment === undefined) {
		cell.alignment = AlignmentValues.Leading; // this is a different rule: if not defined, use leading, which is more common than centered in tables
	}

	switch (cell.alignment) {
		case AlignmentValues.Leading:
			styles.push('vertical-align: middle', 'text-align: start');
			break;
		case AlignmentValues.Trailing:
			styles.push('vertical-align: middle', 'text-align: end');
			break;
		case AlignmentValues.Center:
			styles.push('vertical-align: middle', 'text-align: center');
			break;
		case AlignmentValues.TopLeading:
			styles.push('vertical-align: top', 'text-align: start');
			break;
		case AlignmentValues.BottomLeading:
			styles.push('vertical-align: bottom', 'text-align: start');
			break;
		case AlignmentValues.TopTrailing:
			styles.push('vertical-align: top', 'text-align: end');
			break;
		case AlignmentValues.Top:
			styles.push('vertical-align: top', 'text-align: center');
			break;
		case AlignmentValues.BottomTrailing:
			styles.push('vertical-align: bottom', 'text-align: end');
			break;
		case AlignmentValues.Bottom:
			styles.push('vertical-align: bottom', 'text-align: center');
			break;
		default:
			// nothing, just default
			break;
	}

	if (!cell.cellPadding && props.ui.defaultCellPadding) {
		// default cell padding from the entire table
		styles.push(...paddingCSS(props.ui.defaultCellPadding));
	} else if (cell.cellPadding) {
		// specific cell padding takes precedence
		styles.push(...paddingCSS(cell.cellPadding));
	}

	styles.push(...borderCSS(cell.cellBorder));

	if (cell.cellAction) {
		styles.push('cursor: pointer');
	}

	return styles.join(';');
}

function onClickRow(rowIdx: number) {
	let row = props.ui.rows?.value?.at(rowIdx)!;
	if (row.action) {
		serviceAdapter.sendEvent(new FunctionCallRequested(row.action, nextRID()));
	}
}

function onClickCell(rowIdx: number, colIdx: number) {
	let row = props.ui.rows?.value?.at(rowIdx)!;
	let cell = row.cells?.value.at(colIdx)!;
	if (cell.action) {
		serviceAdapter.sendEvent(new FunctionCallRequested(cell.action, nextRID()));
	} else if (row.action) {
		serviceAdapter.sendEvent(new FunctionCallRequested(row.action, nextRID()));
	}
}

function onClickHeaderCell(colIdx: number) {
	let cell = props.ui.header?.columns?.value?.at(colIdx)!;
	if (cell.cellAction) {
		serviceAdapter.sendEvent(new FunctionCallRequested(cell.cellAction, nextRID()));
	}
}

function onCellMouseEnter(rowIdx: number, colIdx: number) {
	let cell = props.ui.rows?.value?.at(rowIdx)?.cells?.value.at(colIdx)!;
	cell.hovered = true;
}

function onCellMouseLeave(rowIdx: number, colIdx: number) {
	let cell = props.ui.rows?.value?.at(rowIdx)?.cells?.value.at(colIdx)!;
	cell.hovered = false;
}

function onHeadCellMouseEnter(colIdx: number) {
	let cell = props.ui.header?.columns?.value?.at(colIdx)!;
	cell.cellHovered = true;
}

function onHeadCellMouseLeave(colIdx: number) {
	let cell = props.ui.header?.columns?.value?.at(colIdx)!;
	cell.cellHovered = false;
}

function onRowMouseEnter(rowIdx: number) {
	let row = props.ui.rows?.value?.at(rowIdx)!;
	row.hovered = true;
}

function onRowMouseLeave(rowIdx: number) {
	let row = props.ui.rows?.value?.at(rowIdx)!;
	row.hovered = false;
}
</script>

<template>
	<table class="w-full text-left rtl:text-right overflow-clip" :style="frameStyles">
		<thead v-if="props.ui.header?.columns?.value?.length > 0" class="" :style="headStyles()">
			<tr>
				<th
					class="font-normal"
					v-for="(head, headIdx) in props.ui.header?.columns?.value"
					scope="col"
					:style="headCellStyles(headIdx)"
					@click.stop="onClickHeaderCell(headIdx)"
					@mouseenter="onHeadCellMouseEnter(headIdx)"
					@mouseleave="onHeadCellMouseLeave(headIdx)"
				>
					<ui-generic v-if="head.content" :ui="head.content" />
				</th>
			</tr>
		</thead>

		<tbody class="">
			<tr
				v-for="(row, rowIdx) in props.ui.rows?.value"
				:style="rowStyles(rowIdx)"
				@click="onClickRow(rowIdx)"
				@mouseenter="onRowMouseEnter(rowIdx)"
				@mouseleave="onRowMouseLeave(rowIdx)"
			>
				<td
					:rowspan="cell.rowSpan == 0 ? undefined : cell.rowSpan"
					:colspan="cell.colSpan == 0 ? undefined : cell.colSpan"
					v-for="(cell, colIdx) in row.cells?.value"
					:style="cellStyles(rowIdx, colIdx)"
					@click.stop="onClickCell(rowIdx, colIdx)"
					@mouseenter="onCellMouseEnter(rowIdx, colIdx)"
					@mouseleave="onCellMouseLeave(rowIdx, colIdx)"
				>
					<ui-generic v-if="cell.content" :ui="cell.content" />
				</td>
			</tr>
		</tbody>
	</table>
</template>

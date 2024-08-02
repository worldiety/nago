<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {Table} from "@/shared/protocol/ora/table";
import {frameCSS} from "@/components/shared/frame";
import {colorValue} from "@/components/shared/colors";
import {computed} from "vue";
import {borderCSS} from "@/components/shared/border";
import {cssLengthValue} from "@/components/shared/length";
import {Alignment} from "@/components/shared/alignments";
import {paddingCSS} from "@/components/shared/padding";
import {useServiceAdapter} from "@/composables/serviceAdapter";

const props = defineProps<{
	ui: Table;
}>();

const serviceAdapter = useServiceAdapter();

function commonStyles(): string[] {
	let styles = frameCSS(props.ui.f)

	// background handling
	if (props.ui.bgc) {
		styles.push(`background-color: ${colorValue(props.ui.bgc)}`)
	}

	// border handling
	styles.push(...borderCSS(props.ui.b))


	return styles
}

const frameStyles = computed<string>(() => {
	let styles = commonStyles()


	return styles.join(";")
});


function rowStyles(idx: number): string {
	const styles: string[] = [];
	let row = props.ui.r?.at(idx)!
	if (row.b) {
		styles.push(`background-color: ${colorValue(row.b)}`)
	}

	if (row.h) {
		styles.push(`height: ${cssLengthValue(row.h)}`)
	}

	if (idx > 0 && props.ui.rdc) {
		styles.push("border-collapse: collapse", "border-top-width: 1px", `border-color: ${colorValue(props.ui.rdc)}`)
	}

	return styles.join(";")
}

function headStyles() {
	const styles: string[] = [];
	if (props.ui.rdc) {
		styles.push("border-collapse: collapse", "border-bottom-width: 2px", `border-color: ${colorValue(props.ui.rdc)}`)
	}

	return styles.join(";")
}

function cellStyles(rowIdx: number, colIdx: number): string {
	const styles: string[] = [];
	let cell = props.ui.r.at(rowIdx).c.at(colIdx)!
	if (cell.b) {
		styles.push(`background-color: ${colorValue(cell.b)}`)
	}

	switch (cell.a) {
		case Alignment.Leading:
			styles.push("vertical-align: middle", "text-align: start")
			break
		case Alignment.Trailing:
			styles.push("vertical-align: middle", "text-align: end")
			break
		case Alignment.Center:
			styles.push("vertical-align: middle", "text-align: center")
			break
		case Alignment.TopLeading:
			styles.push("vertical-align: top", "text-align: start")
			break
		case Alignment.BottomLeading:
			styles.push("vertical-align: bottom", "text-align: start")
			break
		case Alignment.TopTrailing:
			styles.push("vertical-align: top", "text-align: end")
			break
		case Alignment.Top:
			styles.push("vertical-align: top", "text-align: center")
			break
		case Alignment.BottomTrailing:
			styles.push("vertical-align: bottom", "text-align: end")
			break
		case Alignment.Bottom:
			styles.push("vertical-align: bottom", "text-align: center")
			break
		default:
			// nothing, just default
			break
	}

	if (!cell.p && props.ui.p) {
		// default cell padding from the entire table
		styles.push(...paddingCSS(props.ui.p))
	} else if (cell.p) {
		// specific cell padding takes precedence
		styles.push(...paddingCSS(cell.p))
	}

	styles.push(...borderCSS(cell.o))


	return styles.join(";")
}

function onClickRow(rowIdx: number) {
	let row = props.ui.r?.at(rowIdx)!
	if (row.a) {
		serviceAdapter.executeFunctions(row.a);
	}
}

function onClickCell(rowIdx: number, colIdx: number) {
	let row = props.ui.r?.at(rowIdx)!
	let cell = row.c.at(colIdx)!
	if (cell.t) {
		serviceAdapter.executeFunctions(cell.t);
	} else if (row.a) {
		serviceAdapter.executeFunctions(row.a);
	}
}

function onClickHeaderCell(colIdx: number) {
	let cell = props.ui.h?.c?.at(colIdx)!
	if (cell.t) {
		serviceAdapter.executeFunctions(cell.t);
	}
}
</script>

<template>
	<div class="relative overflow-x-auto">
		<table class="w-full text-left  rtl:text-right" :style="frameStyles">
			<thead
				v-if="props.ui.h?.c"
				class="font-medium"
				:style="headStyles()"
			>
			<tr>
				<th v-for="(head,headIdx) in props.ui.h.c" scope="col" class="px-6 py-3">
					<ui-generic v-if="head.c" :ui="head.c" @click.stop="onClickHeaderCell(headIdx)"/>
				</th>
			</tr>
			</thead>

			<tbody class="">
			<tr
				v-for="(row,rowIdx) in props.ui.r"
				:style="rowStyles(rowIdx)"
				@click="onClickRow(rowIdx)"
			>
				<td :rowspan="cell.rs" :colspan="cell.cs" v-for="(cell,colIdx) in row.c"
						:style="cellStyles(rowIdx,colIdx)" @click.stop="onClickCell(rowIdx,colIdx)">
					<ui-generic v-if="cell.c" :ui="cell.c"/>
				</td>
			</tr>
			</tbody>
		</table>
	</div>
</template>

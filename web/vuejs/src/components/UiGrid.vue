<script lang="ts" setup>
import {computed} from 'vue';
import UiGridCell from '@/components/UiGridCell.vue';
import {Grid} from "@/shared/protocol/ora/grid";
import {frameCSS} from "@/components/shared/frame";
import {colorValue} from "@/components/shared/colors";
import {borderCSS} from "@/components/shared/border";
import {paddingCSS} from "@/components/shared/padding";
import {fontCSS} from "@/components/shared/font";
import {cssLengthValue} from "@/components/shared/length";

const props = defineProps<{
	ui: Grid;
}>();


const style = computed<string>(() => {
	const styles: string[] = [];

	if (!props.ui.c || props.ui.c === 0) {
		styles.push("grid-auto-columns: auto")
	} else {
		if (!props.ui.cw) {
			styles.push(`grid-template-columns: repeat(${props.ui.c}, minmax(0, 1fr))`)
		} else {
			let tmp = "grid-template-columns: "
			for (let len of props.ui.cw) {
				tmp += cssLengthValue(len)
				tmp += " "
			}

			let restColCount = props.ui.c - props.ui.cw.length
			if (restColCount > 0) {
				tmp += `repeat(${restColCount}, 1fr)`
			}

			styles.push(tmp)
		}
	}

	if (!props.ui.r || props.ui.r === 0) {
		styles.push("grid-auto-rows: auto")
	} else {
		styles.push(`grid-template-rows: repeat(${props.ui.r}, minmax(0, 1fr))`)
	}

	if (props.ui.rg) {
		styles.push(`row-gap: ${props.ui.rg}`)
	}

	if (props.ui.cg) {
		styles.push(`column-gap: ${props.ui.cg}`)
	}

	if (props.ui.bgc) {
		styles.push(`background-color: ${colorValue(props.ui.bgc)}`)
	}


	styles.push(...frameCSS(props.ui.f))
	styles.push(...borderCSS(props.ui.bd))
	styles.push(...paddingCSS(props.ui.p))
	styles.push(...fontCSS(props.ui.fn))

	return styles.join(";");
});


const clazz = computed<string>(() => {
	const styles: string[] = [];
	styles.push("grid overflow-clip")


	return styles.join(" ")
});
</script>

<template>


	<div :class="clazz" :style="style">
		<ui-grid-cell v-for="cell in props.ui.b" :ui="cell"/>
	</div>
</template>

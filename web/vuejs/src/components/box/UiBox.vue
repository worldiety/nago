<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {computed} from 'vue';
import {frameCSS} from "@/components/shared/frame";
import {Box} from "@/shared/protocol/ora/box";
import {Alignment} from "@/shared/protocol/ora/alignment";

import {Alignment as Al} from "@/components/shared/alignments";
import {paddingCSS} from "@/components/shared/padding";
import {Padding} from "@/shared/protocol/ora/padding";
import {cssLengthValue0Px} from "@/components/shared/length";

const props = defineProps<{
	ui: Box;
}>();


const frameStyles = computed<string>(() => {
	let styles = frameCSS(props.ui.frame)
	if (props.ui.bgc) {
		styles.push(`background-color: ${props.ui.bgc}`)
	}

	styles.push(...paddingCSS(props.ui.p))

	return styles.join(";")
});

const clazz = computed<string>(() => {
	let classes = "relative flex";

	return classes
});

function childMargin(a?: Alignment, p?: Padding): string {
	switch (a) {
		case Al.BottomLeading:
			return `margin-left: ${cssLengthValue0Px(p?.l)};margin-bottom: ${cssLengthValue0Px(p?.b)}`
		case Al.TopLeading:
			return `margin-top: ${cssLengthValue0Px(p?.t)};margin-left: ${cssLengthValue0Px(p?.l)}`
		case Al.Leading:
			return `margin-left: ${cssLengthValue0Px(p?.l)}`
		case Al.Top:
			return `margin-top: ${cssLengthValue0Px(p?.t)}`
		case Al.Bottom:
			return `margin-bottom: ${cssLengthValue0Px(p?.b)}`
		case Al.Trailing:
			return `margin-right: ${cssLengthValue0Px(p?.r)}`
		case Al.BottomTrailing:
			return `margin-right: ${cssLengthValue0Px(p?.r)};margin-bottom: ${cssLengthValue0Px(p?.b)}`
		case Al.TopTrailing:
			return `margin-right: ${cssLengthValue0Px(p?.r)};margin-top: ${cssLengthValue0Px(p?.t)}`


	}

	return ""
}

function childClass(a?: Alignment): string {
	// we also use flex for the simple cases, because otherwise we have a gap and even more weired rendering if not enough room
	// note, that flex never calculates the width properly, even with border-box etc.
	// we will use margin instead
	switch (a) {
		case Al.BottomLeading:
			return "absolute flex bottom-0 left-0"
		case Al.TopLeading:
			return "absolute flex top-0 left-0"
		case Al.TopTrailing:
			return "absolute flex top-0 right-0"
		case Al.BottomTrailing:
			return "absolute flex right-0 bottom-0"
		case Al.Top:
			return "absolute w-full flex justify-center top-0"
		case Al.Bottom:
			return "absolute w-full flex justify-center bottom-0"
		case Al.Leading:
			return "absolute h-full flex items-center left-0"
		case Al.Trailing:
			return "absolute h-full flex items-center right-0"
		default:
			return "absolute w-full h-full flex justify-center items-center"
	}

}

</script>

<template v-if="props.ui.children">
	<div :class="clazz" :style="frameStyles">
		<div v-for="ui in props.ui.c" :class="childClass(ui.a)" :style="childMargin(ui.a,props.ui.p)">
			<ui-generic :ui="ui.c"/>
		</div>
	</div>
</template>

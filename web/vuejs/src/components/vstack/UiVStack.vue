<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {computed} from 'vue';
import {frameCSS} from "@/components/shared/frame";
import {Alignment} from "@/components/shared/alignments";
import {VStack} from "@/shared/protocol/ora/vStack";
import {cssLengthValue} from "@/components/shared/length";
import {paddingCSS} from "@/components/shared/padding";
import {colorValue} from "@/components/shared/colors";

const props = defineProps<{
	ui: VStack;
}>();


const frameStyles = computed<string>(() => {
	let styles = frameCSS(props.ui.f)
	if (props.ui.bgc) {
		styles.push(`background-color: ${colorValue(props.ui.bgc)}`)
	}

	if (props.ui.g) {
		styles.push(`row-gap:${cssLengthValue(props.ui.g)}`)
	}

	styles.push(...paddingCSS(props.ui.p).join(";"))

	return styles.join(";")
});

const clazz = computed<string>(() => {
	let classes = "inline-flex flex-col ";
	switch (props.ui.a) {
		case Alignment.Leading:
			classes += " justify-center items-start "
			break
		case Alignment.Trailing:
			classes += " justify-center items-end "
			break
		case Alignment.Center:
			classes += " justify-center items-center "
			break
		case Alignment.TopLeading:
			classes += " justify-start items-start "
			break
		case Alignment.BottomLeading:
			classes += " justify-end items-start "
			break
		case Alignment.TopTrailing:
			classes += " justify-start items-end "
			break
		case Alignment.Top:
			classes += " justify-start items-center "
			break
		case Alignment.BottomTrailing:
			classes += " justify-end items-end "
			break
		case Alignment.Bottom:
			classes += " justify-end items-center "
			break
		default:
			classes += " justify-center items-center "
			break

	}

	return classes
});
</script>

<template v-if="props.ui.children">
	<div :class="clazz" :style="frameStyles">
		<ui-generic v-for="ui in props.ui.c" :ui="ui"/>
	</div>
</template>

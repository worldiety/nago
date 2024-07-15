<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {computed} from 'vue';
import {frameCSS} from "@/components/shared/frame";
import {Alignment} from "@/components/shared/alignments";
import {namedColorClasses, namedColorStyles} from "@/components/shared/namedcolors";
import {VStack} from "@/shared/protocol/ora/vStack";
import {cssLengthValue} from "@/components/shared/length";
import {paddingCSS} from "@/components/shared/padding";

const props = defineProps<{
	ui: VStack;
}>();


const frameStyles = computed<string>(() => {
	let s = frameCSS(props.ui.f).join(";")
	let c = namedColorStyles("background-color", props.ui.bgc)

	let gap = ""
	if (props.ui.g) {
		gap = `row-gap:${cssLengthValue(props.ui.g)}`
	}

	return [s, c, gap, paddingCSS(props.ui.p).join(";")].join(";")
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


	if (props.ui.bgc != undefined && props.ui.bgc !== "") {
		classes += namedColorClasses(props.ui.bgc)
	}


	return classes
});
</script>

<template v-if="props.ui.children">
	<div :class="clazz" :style="frameStyles">
		<ui-generic v-for="ui in props.ui.c" :ui="ui"/>
	</div>
</template>

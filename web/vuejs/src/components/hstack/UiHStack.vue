<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {computed} from 'vue';
import {HStack} from "@/shared/protocol/ora/hStack";
import {createFrameStyles} from "@/components/shared/frame";
import {Alignment} from "@/components/shared/alignments";
import {namedColorClasses, namedColorStyles} from "@/components/shared/namedcolors";

const props = defineProps<{
	ui: HStack;
}>();


const frameStyles = computed<string>(() => {
	let s = createFrameStyles(props.ui.frame)
	let c = namedColorStyles("background-color", props.ui.backgroundColor)

	return [s,c].join(";")
});

const clazz = computed<string>(() => {
	let classes = "inline-flex ";
	switch (props.ui.alignment) {
		case Alignment.Leading:
			classes += " justify-start items-center "
			break
		case Alignment.Trailing:
			classes += " justify-end items-center "
			break
		case Alignment.Center:
			classes += " justify-center items-center "
			break
		case Alignment.TopLeading:
			classes += " justify-start items-start "
			break
		case Alignment.BottomLeading:
			classes += " justify-start items-end "
			break
		case Alignment.TopTrailing:
			classes += " justify-end items-start "
			break
		case Alignment.Top:
			classes += " justify-center items-start "
			break
		case Alignment.BottomTrailing:
			classes += " justify-end items-end "
			break
		case Alignment.Bottom:
			classes += " justify-center items-end "
			break
		default:
			classes += " justify-center items-center "
			break

	}


	if (props.ui.backgroundColor != undefined && props.ui.backgroundColor !== "") {
		classes += namedColorClasses(props.ui.backgroundColor)
	}


	return classes
});
</script>

<template v-if="props.ui.children">
	<div :class="clazz" :style="frameStyles" >
		<ui-generic  v-for="ui in props.ui.children" :ui="ui"/>
	</div>
</template>

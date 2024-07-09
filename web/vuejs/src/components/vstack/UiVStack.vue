<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {computed} from 'vue';
import {HStack} from "@/shared/protocol/ora/hStack";
import {createFrameStyles} from "@/components/shared/frame";
import {Alignment} from "@/components/shared/alignments";
import {namedColorClasses, namedColorStyles} from "@/components/shared/namedcolors";
import {VStack} from "@/shared/protocol/ora/vStack";

const props = defineProps<{
	ui: VStack;
}>();


const frameStyles = computed<string>(() => {
	let s = createFrameStyles(props.ui.frame)
	let c = namedColorStyles("background-color", props.ui.backgroundColor)

	return [s,c].join(";")
});

const clazz = computed<string>(() => {
	let classes = "inline-flex flex-col ";
	switch (props.ui.alignment) {
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

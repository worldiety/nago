<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {computed, ref} from 'vue';
import {frameCSS} from "@/components/shared/frame";
import {Alignment} from "@/components/shared/alignments";
import {VStack} from "@/shared/protocol/ora/vStack";
import {cssLengthValue} from "@/components/shared/length";
import {paddingCSS} from "@/components/shared/padding";
import {colorValue} from "@/components/shared/colors";
import {fontCSS} from "@/components/shared/font";
import {borderCSS} from "@/components/shared/border";
import {useServiceAdapter} from "@/composables/serviceAdapter";

const props = defineProps<{
	ui: VStack;
}>();

const hover = ref(false);
const pressed = ref(false);
const focused = ref(false);
const focusable = ref(false);
const serviceAdapter = useServiceAdapter();

function onClick() {
	if (props.ui.t){
		serviceAdapter.executeFunctions(props.ui.t);
	}
}

// copy-paste me into UiText, UiVStack and UiHStack (or refactor me into some kind of generics-getter-setter-nightmare).
function commonStyles():string[]{
	let styles = frameCSS(props.ui.f)

	// background handling
	if (props.ui.pgc && pressed.value) {
		styles.push(`background-color: ${colorValue(props.ui.pgc)}`)
	} else {
		if (props.ui.hgc) {
			if (hover.value) {
				styles.push(`background-color: ${colorValue(props.ui.hgc)}`)
			} else {
					styles.push(`background-color: ${colorValue(props.ui.bgc)}`)
			}
		}else{
				styles.push(`background-color: ${colorValue(props.ui.bgc)}`)
		}
	}

	if (props.ui.fbc) {
		focusable.value = true;
		if (focused.value && !pressed.value) {
			styles.push(`background-color: ${colorValue(props.ui.fbc)}`)
		}
	}

	// border handling
	if (props.ui.pb && pressed.value){
		styles.push(...borderCSS(props.ui.pb))
	}else{
		if (props.ui.hb){
			if (hover.value){
				styles.push(...borderCSS(props.ui.hb))
			}else{
					styles.push(...borderCSS(props.ui.b))
			}
		}else{
			styles.push(...borderCSS(props.ui.b))
		}
	}

	if (props.ui.fb){
		focusable.value = true;
		if (focused.value && !pressed.value) {
			styles.push(...borderCSS(props.ui.fb))
		}
	}

	// other stuff
	styles.push(...paddingCSS(props.ui.p))
	styles.push(...fontCSS(props.ui.fn))

	if (focused.value){
		styles.push("outline: 2px solid black") // always apply solid and never auto. Auto will create random broken effects on firefox and chrome
	}

	return styles
}

const frameStyles = computed<string>(() => {
	let styles = commonStyles()

	if (props.ui.g) {
		styles.push(`row-gap:${cssLengthValue(props.ui.g)}`)
	}

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

	if (props.ui.t){
		classes+=" cursor-pointer "
	}

	return classes
});
</script>

<template v-if="props.ui.children">
	<div :class="clazz" :style="frameStyles" @mouseover="hover = true" @mouseleave="hover = false"
			 @mousedown="pressed = true" @mouseup="pressed = false" @mouseout="pressed = false" @focusin="focused = true"
			 @focusout="focused = false" :tabindex="focusable?0:-1" @click="onClick">
		<ui-generic v-for="ui in props.ui.c" :ui="ui"/>
	</div>
</template>

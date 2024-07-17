<script lang="ts" setup>
import {computed} from 'vue';
import type {Text} from "@/shared/protocol/ora/text";
import {useServiceAdapter} from '@/composables/serviceAdapter';
import {isNil} from "@/shared/protocol/util";
import {frameCSS} from "@/components/shared/frame";
import {paddingCSS} from "@/components/shared/padding";
import {cssLengthValue} from "@/components/shared/length";
import {colorValue} from "@/components/shared/colors";

const props = defineProps<{
	ui: Text;
}>();

const serviceAdapter = useServiceAdapter();


const styles = computed<string>(() => {
	let styles = frameCSS(props.ui.f)
	if (props.ui.color) {
		styles.push(`color: ${colorValue(props.ui.color)}`)
	}

	if (props.ui.backgroundColor) {
		styles.push(`background-color: ${colorValue(props.ui.backgroundColor)}`)
	}

	styles.push(...paddingCSS(props.ui.p))
	if (props.ui) {
		styles.push(`font-size: ${cssLengthValue(props.ui.s)}`)
	}
	console.log(styles)
	return styles.join(";")
});


function onClick() {
	if (!isNil(props.ui.onClick)) {
		serviceAdapter.executeFunctions(props.ui.onClick);
	}
}

function onMouseEnter() {
	if (!isNil(props.ui.onHoverEnd)) {
		serviceAdapter.executeFunctions(props.ui.onHoverEnd);
	}
}

function onMouseLeave() {
	if (!isNil(props.ui.onHoverEnd)) {
		serviceAdapter.executeFunctions(props.ui.onHoverEnd);
	}
}
</script>

<template>
	<span v-if="!ui.invisible" :style="styles" @click="onClick" @mouseenter="onMouseEnter"
				@mouseleave="onMouseLeave">{{
			props.ui.value
		}}</span>
</template>

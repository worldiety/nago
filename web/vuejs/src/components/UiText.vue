<script lang="ts" setup>
import {computed} from 'vue';
import type {Text} from "@/shared/protocol/ora/text";
import {useServiceAdapter} from '@/composables/serviceAdapter';
import {isNil} from "@/shared/protocol/util";
import {frameCSS} from "@/components/shared/frame";
import {paddingCSS} from "@/components/shared/padding";
import {cssLengthValue} from "@/components/shared/length";
import {colorValue} from "@/components/shared/colors";
import {fontCSS} from "@/components/shared/font";
import {borderCSS} from "@/components/shared/border";

const props = defineProps<{
	ui: Text;
}>();

const serviceAdapter = useServiceAdapter();


const styles = computed<string>(() => {
	let styles = frameCSS(props.ui.f)
	if (props.ui.c) {
		styles.push(`color: ${colorValue(props.ui.c)}`)
	}

	if (props.ui.bgc) {
		styles.push(`background-color: ${colorValue(props.ui.bgc)}`)
	}

	styles.push(...borderCSS(props.ui.b))
	styles.push(...paddingCSS(props.ui.p))
	styles.push(...fontCSS(props.ui.o))

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
	<span v-if="!ui.i" :style="styles" @click="onClick" @mouseenter="onMouseEnter"
				@mouseleave="onMouseLeave">{{
			props.ui.v
		}}</span>
</template>

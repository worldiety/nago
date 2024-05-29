<script lang="ts" setup>
import { textColor2Tailwind, textSize2Tailwind } from '@/shared/tailwindTranslator';
import { computed } from 'vue';
import type {Text} from "@/shared/protocol/ora/text";
import { useServiceAdapter } from '@/composables/serviceAdapter';

const props = defineProps<{
	ui: Text;
}>();

const serviceAdapter = useServiceAdapter();

const clazz = computed<string>(() => {
	let tmp = '';
	if (props.ui.color.v) {
		tmp += textColor2Tailwind(props.ui.color.v);
	} else {
		tmp += 'text-black';
	}

	if (props.ui.size.v) {
		tmp += ' ' + textSize2Tailwind(props.ui.size.v);
	}

	return tmp;
});

function onClick() {
	serviceAdapter.executeFunctions(props.ui.onClick);
}

function onMouseEnter() {
	serviceAdapter.executeFunctions(props.ui.onHoverStart);
}

function onMouseLeave() {
	serviceAdapter.executeFunctions(props.ui.onHoverEnd);
}
</script>

<template>
	<span v-if="ui.visible.v" :class="clazz" @click="onClick" @mouseenter="onMouseEnter" @mouseleave="onMouseLeave">{{
		props.ui.value.v
	}}</span>
</template>

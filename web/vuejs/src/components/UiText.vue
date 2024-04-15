<script lang="ts" setup>
import { textColor2Tailwind, textSize2Tailwind } from '@/shared/tailwindTranslator';
import { computed } from 'vue';
import { useNetworkStore } from '@/stores/networkStore';
import type { LivePage } from '@/shared/model/livePage';
import {Text} from "@/shared/protocol/gen/text";

const props = defineProps<{
	ui: Text;
	page: LivePage;
}>();

const networkStore = useNetworkStore();

const clazz = computed<string>(() => {
	let tmp = '';
	if (props.ui.color.v) {
		tmp += textColor2Tailwind(props.ui.color.v);
	} else {
		tmp += 'text-gray-900';
	}

	if (props.ui.colorDark.v) {
		tmp += ' dark:' + textColor2Tailwind(props.ui.color.v);
	} else {
		tmp += ' dark:text-white';
	}

	if (props.ui.size.v) {
		tmp += ' ' + textSize2Tailwind(props.ui.size.v);
	}

	return tmp;
});

function onClick() {
	networkStore.invokeFunctions(props.ui.onClick);
}

function onMouseEnter() {
	networkStore.invokeFunctions(props.ui.onHoverStart);
}

function onMouseLeave() {
	networkStore.invokeFunctions(props.ui.onHoverEnd);
}
</script>

<template>
	<span :class="clazz" @click="onClick" @mouseenter="onMouseEnter" @mouseleave="onMouseLeave">{{
		props.ui.value.v
	}}</span>
</template>

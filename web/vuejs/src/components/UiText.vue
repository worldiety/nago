<script lang="ts" setup>
import { textColor2Tailwind, textSize2Tailwind } from '@/shared/tailwindTranslator';
import { computed } from 'vue';
import { useNetworkStore } from '@/stores/networkStore';
import type { LiveText } from '@/shared/model/liveText';
import type { LivePage } from '@/shared/model/livePage';

const props = defineProps<{
	ui: LiveText;
	page: LivePage;
}>();

const networkStore = useNetworkStore();

const clazz = computed<string>(() => {
	let tmp = '';
	if (props.ui.color.value) {
		tmp += textColor2Tailwind(props.ui.color.value);
	} else {
		tmp += 'text-gray-900';
	}

	if (props.ui.colorDark.value) {
		tmp += ' dark:' + textColor2Tailwind(props.ui.color.value);
	} else {
		tmp += ' dark:text-white';
	}

	if (props.ui.size.value) {
		tmp += ' ' + textSize2Tailwind(props.ui.size.value);
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
		props.ui.value.value
	}}</span>
</template>

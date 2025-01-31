<script lang="ts" setup>
import { computed } from 'vue';
import { borderCSS } from '@/components/shared/border';
import { colorValue } from '@/components/shared/colors';
import { frameCSS } from '@/components/shared/frame';
import { paddingCSS } from '@/components/shared/padding';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import type { Image } from '@/shared/protocol/ora/image';

const props = defineProps<{
	ui: Image;
}>();

const serviceAdapter = useServiceAdapter();

const styles = computed<string>(() => {
	let styles = borderCSS(props.ui.b);
	styles.push(...frameCSS(props.ui.f));
	styles.push(...paddingCSS(props.ui.p));

	if (!props.ui.s) {
		styles.push('object-fit: cover');
	}

	if (props.ui.c) {
		styles.push(`fill: ${colorValue(props.ui.c)}`);
	}

	if (props.ui.k) {
		styles.push(`stroke: ${colorValue(props.ui.k)}`);
	}

	return styles.join(';');
});

const rewriteSVG = computed<string>(() => {
	if (!props.ui.s && !props.ui.v) {
		return '';
	}

	let data = 'svg cache error';
	if (props.ui.s) {
		data = props.ui.s;
	} else {
		if (props.ui.v) {
			let tmp = serviceAdapter.getBufferFromCache(props.ui.v);
			if (tmp) {
				data = tmp;
			}
		}
	}

	if (props.ui.s && props.ui.v) {
		serviceAdapter.setBufferToCache(props.ui.v, props.ui.s);
	}

	return data.replace('<svg ', `<svg style="${styles.value}" `);
});
</script>

<template>
	<img
		v-if="!ui.iv && !ui.s && !props.ui.v"
		class="h-auto max-w-full"
		:src="props.ui.u"
		:alt="props.ui.al"
		:title="props.ui.al"
		:style="styles"
	/>
	<div :title="props.ui.al" v-if="props.ui.s || props.ui.v" v-html="rewriteSVG"></div>
</template>

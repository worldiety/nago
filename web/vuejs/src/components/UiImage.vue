<script lang="ts" setup>
import { computed } from 'vue';
import { borderCSS } from '@/components/shared/border';
import { colorValue } from '@/components/shared/colors';
import { frameCSS } from '@/components/shared/frame';
import { paddingCSS } from '@/components/shared/padding';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import {Img} from "@/shared/proto/nprotoc_gen";

const props = defineProps<{
	ui: Img;
}>();

const serviceAdapter = useServiceAdapter();

const styles = computed<string>(() => {
	let styles = borderCSS(props.ui.border);
	styles.push(...frameCSS(props.ui.frame));
	styles.push(...paddingCSS(props.ui.padding));

	if (props.ui.sVG.isZero()) {
		// special case for normal images, not for svg
		styles.push('object-fit: cover');
	}

	if (!props.ui.fillColor.isZero()) {
		styles.push(`fill: ${colorValue(props.ui.fillColor.value)}`);
	}

	if (!props.ui.strokeColor.isZero()) {
		styles.push(`stroke: ${colorValue(props.ui.strokeColor.value)}`);
	}

	return styles.join(';');
});

const rewriteSVG = computed<string>(() => {
	if (!props.ui.sVG.isZero() ) {
		return '';
	}

	// todo how to optimize this svg handling which is probably very expensive
	let data = props.ui.sVG.value;

	return data.replace('<svg ', `<svg style="${styles.value}" `);
});
</script>

<template>
	<img
		v-if="!ui.invisible.value && ui.sVG.isZero()"
		class="h-auto max-w-full"
		:src="props.ui.uri.value"
		:alt="props.ui.accessibilityLabel.value"
		:title="props.ui.accessibilityLabel.value"
		:style="styles"
	/>
	<div :title="props.ui.accessibilityLabel.value" v-if="!props.ui.sVG.isZero()" v-html="rewriteSVG"></div>
</template>

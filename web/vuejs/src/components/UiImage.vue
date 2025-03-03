<script lang="ts" setup>
import { computed } from 'vue';
import { borderCSS } from '@/components/shared/border';
import { colorValue } from '@/components/shared/colors';
import { frameCSS } from '@/components/shared/frame';
import { paddingCSS } from '@/components/shared/padding';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { FunctionCallRequested, Img } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Img;
}>();

const serviceAdapter = useServiceAdapter();

const styles = computed<string>(() => {
	let styles = borderCSS(props.ui.border);
	styles.push(...frameCSS(props.ui.frame));
	styles.push(...paddingCSS(props.ui.padding));

	if (!props.ui.sVG) {
		// special case for normal images, not for svg
		styles.push('object-fit: cover');
	}

	if (props.ui.fillColor) {
		styles.push(`fill: ${colorValue(props.ui.fillColor)}`);
	}

	if (props.ui.strokeColor) {
		styles.push(`stroke: ${colorValue(props.ui.strokeColor)}`);
	}

	return styles.join(';');
});

function ngCall(ptr: number) {
	serviceAdapter.sendEvent(new FunctionCallRequested(ptr, nextRID()));
}

function invokePointer(evt: Event) {
	//console.log(evt);
	if (!evt.target) {
		return;
	}

	if (evt.target instanceof SVGElement) {
		if (evt.target.ariaValueNow) {
			ngCall(Number(evt.target.ariaValueNow));
		}
	}
}

const rewriteSVG = computed<string>(() => {
	if (!props.ui.sVG) {
		return '';
	}

	// todo how to optimize this svg handling which is probably very expensive
	let data = props.ui.sVG;

	return data.replace('<svg ', `<svg style="${styles.value}" `);
});
</script>

<template>
	<img
		v-if="!ui.invisible && !ui.sVG"
		class="h-auto max-w-full"
		:src="props.ui.uri"
		:alt="props.ui.accessibilityLabel"
		:title="props.ui.accessibilityLabel"
		:style="styles"
	/>
	<div
		@click.capture="invokePointer"
		:title="props.ui.accessibilityLabel"
		v-if="props.ui.sVG"
		v-html="rewriteSVG"
	></div>
</template>

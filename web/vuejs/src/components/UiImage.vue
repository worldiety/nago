<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed } from 'vue';
import { borderCSS } from '@/components/shared/border';
import { colorValue } from '@/components/shared/colors';
import { frameCSS } from '@/components/shared/frame';
import { paddingCSS } from '@/components/shared/padding';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { FunctionCallRequested, Img, ObjectFitValues } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Img;
}>();

const serviceAdapter = useServiceAdapter();

const styles = computed<string>(() => {
	let styles = borderCSS(props.ui.border);
	styles.push(...frameCSS(props.ui.frame));
	styles.push(...paddingCSS(props.ui.padding));

	switch (props.ui.objectFit) {
		case ObjectFitValues.Fill:
			styles.push('object-fit: fill');
			break;
		case ObjectFitValues.None:
			styles.push('object-fit: none');
			break;
		case ObjectFitValues.Contain:
			styles.push('object-fit: contain');
			break;
		case ObjectFitValues.Cover:
			styles.push('object-fit: cover');
			break;
		default:
			// Auto behavior and unknown states
			if (!props.ui.sVG) {
				// special case for normal images, not for svg
				styles.push('object-fit: cover');
				styles.push('aspect-ratio: 1');
			}
			break;
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

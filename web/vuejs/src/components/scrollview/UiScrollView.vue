<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed, nextTick, watch } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { borderCSS } from '@/components/shared/border';
import { colorValue } from '@/components/shared/colors';
import { frameCSS } from '@/components/shared/frame';
import { paddingCSS } from '@/components/shared/padding';
import { positionCSS } from '@/components/shared/position';
import { ScrollAnimationValues, ScrollView, ScrollViewAxisValues } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: ScrollView;
}>();

const styles = computed<string>(() => {
	let styles = borderCSS(props.ui.border);
	styles.push(...frameCSS(props.ui.frame));
	if (props.ui.backgroundColor) {
		styles.push(`background-color: ${colorValue(props.ui.backgroundColor)}`);
	}

	styles.push(...positionCSS(props.ui.position));
	styles.push(...borderCSS(props.ui.border));
	styles.push(...paddingCSS(props.ui.padding));

	return styles.join(';');
});

const classes = computed<string>(() => {
	const css: string[] = [];

	// note, that we defined its style in scrollbars.css
	switch (props.ui.axis) {
		case ScrollViewAxisValues.ScrollViewAxisHorizontal:
			css.push('overflow-x-auto', 'overflow-y-hidden');
			break;
		case ScrollViewAxisValues.ScrollViewAxisBoth:
			css.push('overflow-x-auto', 'overflow-y-auto');
			break;
		default:
			css.push('overflow-y-auto', 'overflow-x-hidden');
			break;
	}

	return css.join(' ');
});

const innerStyles = computed<string>(() => {
	const css: string[] = []; //borderCSS(props.ui.border);

	switch (props.ui.axis) {
		case ScrollViewAxisValues.ScrollViewAxisHorizontal:
			css.push('min-width: max-content');
			break;
		default:
			css.push('height: max-content');
			break;
	}

	return css.join(';');
});

watch(props, (newValue, oldValue) => {
	if (props.ui.scrollIntoView) {
		let id = props.ui.scrollIntoView;
		nextTick(() => {
			const child = document.getElementById(id);
			switch (props.ui.scrollAnimation) {
				case ScrollAnimationValues.Instant:
					child?.scrollIntoView({});
					break;
				default:
					child?.scrollIntoView({ behavior: 'smooth', block: 'end' });
			}
		});
	}
});

// note that we need the max-content hack, otherwise we get layout bugs at least for horizontal areas
</script>

<template v-if="props.ui.iv">
	<!-- UiScrollView -->
	<div :class="classes" :style="styles">
		<div :style="innerStyles">
			<UiGeneric v-if="ui.content" :ui="ui.content" />
		</div>
	</div>
</template>

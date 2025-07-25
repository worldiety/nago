<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed, ref } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { borderCSS } from '@/components/shared/border';
import { colorValue } from '@/components/shared/colors';
import { fontCSS } from '@/components/shared/font';
import { frameCSS } from '@/components/shared/frame';
import { cssLengthValue } from '@/components/shared/length';
import { paddingCSS } from '@/components/shared/padding';
import { positionCSS } from '@/components/shared/position';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { AlignmentValues, FunctionCallRequested, StylePresetValues, VStack } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: VStack;
}>();

const hover = ref(false);
const pressed = ref(false);
const focused = ref(false);
const focusable = ref(false);
const serviceAdapter = useServiceAdapter();

function onClick(event: Event) {
	if (props.ui.action) {
		event.stopPropagation();
		serviceAdapter.sendEvent(new FunctionCallRequested(props.ui.action, nextRID()));
	}
}

function onKeydown(event: KeyboardEvent) {
	if (props.ui.action) {
		event.stopPropagation();
		if (event.code === 'Enter' || event.code === 'Space') {
			serviceAdapter.sendEvent(new FunctionCallRequested(props.ui.action, nextRID()));
		}
	}
}

// copy-paste me into UiText, UiVStack and UiHStack (or refactor me into some kind of generics-getter-setter-nightmare).
function commonStyles(): string[] {
	let styles = frameCSS(props.ui.frame);
	styles.push(...positionCSS(props.ui.position));

	// background handling
	if (props.ui.pressedBackgroundColor && pressed.value) {
		styles.push(`background-color: ${colorValue(props.ui.pressedBackgroundColor)}`);
	} else {
		if (props.ui.hoveredBackgroundColor) {
			if (hover.value) {
				styles.push(`background-color: ${colorValue(props.ui.hoveredBackgroundColor)}`);
			} else {
				styles.push(`background-color: ${colorValue(props.ui.backgroundColor)}`);
			}
		} else {
			styles.push(`background-color: ${colorValue(props.ui.backgroundColor)}`);
		}
	}

	if (props.ui.action) {
		focusable.value = true;
	}

	if (props.ui.focusedBackgroundColor) {
		focusable.value = true;
		if (focused.value && !pressed.value) {
			styles.push(`background-color: ${colorValue(props.ui.focusedBackgroundColor)}`);
		}
	}

	// border handling
	if (props.ui.pressedBorder && pressed.value) {
		styles.push(...borderCSS(props.ui.pressedBorder));
	} else {
		if (props.ui.hoveredBorder) {
			if (hover.value) {
				styles.push(...borderCSS(props.ui.hoveredBorder));
			} else {
				styles.push(...borderCSS(props.ui.border));
			}
		} else {
			styles.push(...borderCSS(props.ui.border));
		}
	}

	if (props.ui.focusedBorder) {
		focusable.value = true;
		if (focused.value && !pressed.value) {
			styles.push(...borderCSS(props.ui.focusedBorder));
		}
	}

	// other stuff
	styles.push(...paddingCSS(props.ui.padding));
	styles.push(...fontCSS(props.ui.font));

	if (focusable.value && focused.value) {
		styles.push('outline: 2px solid black'); // always apply solid and never auto. Auto will create random broken effects on firefox and chrome
	}

	if (props.ui.textColor) {
		styles.push(`color: ${colorValue(props.ui.textColor)}`);
	}

	return styles;
}

const frameStyles = computed<string>(() => {
	let styles = commonStyles();

	if (props.ui.gap) {
		styles.push(`row-gap:${cssLengthValue(props.ui.gap)}`);
	}

	return styles.join(';');
});

const clazz = computed<string>(() => {
	let classes = ['inline-flex', 'flex-col'];
	if (!props.ui.noClip) {
		classes.push('overflow-clip');
	} else {
		classes.push('overflow-visible');
	}

	switch (props.ui.alignment) {
		case AlignmentValues.Stretch:
			classes.push('items-stretch');
			break;
		case AlignmentValues.Leading:
			classes.push('justify-center', 'items-start');
			break;
		case AlignmentValues.Trailing:
			classes.push('justify-center', 'items-end');
			break;
		case AlignmentValues.Center:
			classes.push('justify-center', 'items-center');
			break;
		case AlignmentValues.TopLeading:
			classes.push('justify-start', 'items-start');
			break;
		case AlignmentValues.BottomLeading:
			classes.push('justify-end', 'items-start');
			break;
		case AlignmentValues.TopTrailing:
			classes.push('justify-start', 'items-end');
			break;
		case AlignmentValues.Top:
			classes.push('justify-start', 'items-center');
			break;
		case AlignmentValues.BottomTrailing:
			classes.push('justify-end', 'items-end');
			break;
		case AlignmentValues.Bottom:
			classes.push('justify-end', 'items-center');
			break;
		default:
			classes.push('justify-center', 'items-center');
			break;
	}

	if (props.ui.action) {
		classes.push('cursor-pointer');
	}

	switch (props.ui.stylePreset) {
		case StylePresetValues.StyleButtonPrimary:
			classes.push('button-primary');
			break;
		case StylePresetValues.StyleButtonSecondary:
			classes.push('button-secondary');
			break;
		case StylePresetValues.StyleButtonTertiary:
			classes.push('button-tertiary');
			break;
	}

	return classes.join(' ');
});
</script>

<template>
	<!-- vstack-->
	<div
		:id="ui.id"
		v-if="
			(props.ui.stylePreset === StylePresetValues.StyleNone || props.ui.stylePreset === undefined) &&
			!props.ui.invisible
		"
		:class="clazz"
		:style="frameStyles"
		@mouseover="hover = true"
		:title="props.ui.accessibilityLabel"
		@mouseleave="hover = false"
		@mousedown="pressed = true"
		@mouseup="pressed = false"
		@mouseout="pressed = false"
		@focusin="focused = true"
		@focusout="focused = false"
		:tabindex="focusable ? 0 : -1"
		@click="onClick"
		@keydown="onKeydown"
	>
		<ui-generic v-for="ui in props.ui.children?.value" :ui="ui" />
	</div>

	<button
		:id="ui.id"
		v-else-if="
			props.ui.stylePreset !== StylePresetValues.StyleNone &&
			props.ui.stylePreset !== undefined &&
			(props.ui.invisible === undefined || !props.ui.invisible)
		"
		:class="clazz"
		:style="frameStyles"
		@click="onClick"
		:title="props.ui.accessibilityLabel"
		:disabled="props.ui.disabled"
	>
		<ui-generic v-for="ui in props.ui.children?.value" :ui="ui" />
	</button>
</template>

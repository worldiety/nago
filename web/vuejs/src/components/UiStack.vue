<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed, onUnmounted, watch } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { backgroundCSS } from '@/components/shared/background';
import { borderCSS } from '@/components/shared/border';
import { colorValue } from '@/components/shared/colors';
import { fontCSS } from '@/components/shared/font';
import { frameCSS } from '@/components/shared/frame';
import { cssLengthValue } from '@/components/shared/length';
import { paddingCSS } from '@/components/shared/padding';
import { positionCSS } from '@/components/shared/position';
import { transformationCSS } from '@/components/shared/transformation';
import { randomStr } from '@/components/shared/util';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { CssStyles } from '@/shared/cssStyles';
import { HStack, VStack } from '@/shared/proto/nprotoc_gen';
import {
	AlignmentValues,
	AnimationValues,
	FunctionCallRequested,
	Img,
	StylePresetValues,
} from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: HStack | VStack;
}>();

const id = props.ui.id || randomStr(16);
const cssStyles = new CssStyles(id);
const serviceAdapter = useServiceAdapter();

const focusable = computed<boolean>(
	() => !!props.ui.action || !!props.ui.focusedBorder || !!props.ui.focusedBackgroundColor
);

const classes = computed<string>(() => {
	const classes = ['inline-flex'];
	if (props.ui instanceof VStack) classes.push('flex-col');
	if (!props.ui.noClip) classes.push('overflow-clip');
	else classes.push('overflow-visible');
	if (props.ui.action) classes.push('cursor-pointer');
	if (props.ui instanceof HStack && props.ui.wrap) classes.push('flex-wrap');
	if (activeStyles.value.length) classes.push('custom-active');
	if (focusStyles.value.length) classes.push('custom-focus');
	if (hoverStyles.value.length) classes.push('custom-hover');
	classes.push(...getAlignmentClasses());
	classes.push(...getPresetClasses());

	const animationClass = getAnimationClass();
	if (animationClass) classes.push(animationClass);

	return classes.join(' ');
});

const activeStyles = computed<string[]>(() => {
	const styles: string[] = [];
	if (props.ui.pressedBorder) styles.push(...borderCSS(props.ui.pressedBorder));
	if (props.ui.pressedBackgroundColor)
		styles.push(`background-color: ${colorValue(props.ui.pressedBackgroundColor)}`);
	return styles;
});

const defaultStyles = computed<string[]>(() => {
	const styles = frameCSS(props.ui.frame);
	styles.push(...borderCSS(props.ui.border));
	styles.push(...positionCSS(props.ui.position));
	styles.push(...transformationCSS(props.ui.transformation));
	styles.push(...backgroundCSS(props.ui.background));
	styles.push(`background-color: ${colorValue(props.ui.backgroundColor)}`);
	styles.push(...paddingCSS(props.ui.padding));
	styles.push(...fontCSS(props.ui.font));

	if (props.ui.opacity) styles.push(`opacity: ${100 - props.ui.opacity}%`);
	if (props.ui.gap) styles.push(`column-gap:${cssLengthValue(props.ui.gap)}`);
	if ((!(props.ui instanceof HStack) || props.ui.wrap) && props.ui.gap)
		styles.push(`row-gap:${cssLengthValue(props.ui.gap)}`);

	return styles;
});

const focusStyles = computed<string[]>(() => {
	const styles: string[] = [];
	if (props.ui.focusedBorder) styles.push(...borderCSS(props.ui.focusedBorder));
	if (props.ui.focusedBackgroundColor)
		styles.push(`background-color: ${colorValue(props.ui.focusedBackgroundColor)}`);
	return styles;
});

const hoverStyles = computed<string[]>(() => {
	const styles: string[] = [];
	if (props.ui.hoveredBorder) styles.push(...borderCSS(props.ui.hoveredBorder));
	if (props.ui.hoveredBackgroundColor)
		styles.push(`background-color: ${colorValue(props.ui.hoveredBackgroundColor)}`);
	return styles;
});

function getAlignmentClasses(): string[] {
	if (props.ui instanceof HStack) {
		switch (props.ui.alignment) {
			case AlignmentValues.Stretch:
				return ['items-stretch'];
			case AlignmentValues.Leading:
				return ['justify-start', 'items-center'];
			case AlignmentValues.Trailing:
				return ['justify-end', 'items-center'];
			case AlignmentValues.Center:
				return ['justify-center', 'items-center'];
			case AlignmentValues.TopLeading:
				return ['justify-start', 'items-start'];
			case AlignmentValues.BottomLeading:
				return ['justify-start', 'items-end'];
			case AlignmentValues.TopTrailing:
				return ['justify-end', 'items-start'];
			case AlignmentValues.Top:
				return ['justify-center', 'items-start'];
			case AlignmentValues.BottomTrailing:
				return ['justify-end', 'items-end'];
			case AlignmentValues.Bottom:
				return ['justify-center', 'items-end'];
			default:
				return ['justify-center', 'items-center'];
		}
	} else {
		switch (props.ui.alignment) {
			case AlignmentValues.Stretch:
				return ['items-stretch'];
			case AlignmentValues.Leading:
				return ['justify-center', 'items-start'];
			case AlignmentValues.Trailing:
				return ['justify-center', 'items-end'];
			case AlignmentValues.Center:
				return ['justify-center', 'items-center'];
			case AlignmentValues.TopLeading:
				return ['justify-start', 'items-start'];
			case AlignmentValues.BottomLeading:
				return ['justify-end', 'items-start'];
			case AlignmentValues.TopTrailing:
				return ['justify-start', 'items-end'];
			case AlignmentValues.Top:
				return ['justify-start', 'items-center'];
			case AlignmentValues.BottomTrailing:
				return ['justify-end', 'items-end'];
			case AlignmentValues.Bottom:
				return ['justify-end', 'items-center'];
			default:
				return ['justify-center', 'items-center'];
		}
	}
}

function getAnimationClass(): string | undefined {
	switch (props.ui.animation) {
		case AnimationValues.AnimateBounce:
			return 'animate-bounce';
		case AnimationValues.AnimatePing:
			return 'animate-ping';
		case AnimationValues.AnimatePulse:
			return 'animate-pulse';
		case AnimationValues.AnimateSpin:
			return 'animate-spin';
		case AnimationValues.AnimateTransition:
			return 'transition-all';
	}
}

function getPresetClasses(): string[] {
	const presetClasses: string[] = [];

	switch (props.ui.stylePreset) {
		case StylePresetValues.StyleButtonPrimary:
			presetClasses.push('button-primary');
			break;
		case StylePresetValues.StyleButtonSecondary:
			presetClasses.push('button-secondary');
			break;
		case StylePresetValues.StyleButtonTertiary:
			presetClasses.push('button-tertiary');
			break;
	}

	if (props.ui.stylePreset) {
		if (props.ui.children?.value.length == 1 && props.ui.children.value[0] instanceof Img) {
			presetClasses.push('!p-0', '!w-10');
		}
	}

	return presetClasses;
}

function onClick() {
	if (!props.ui.action) return;
	serviceAdapter.sendEvent(new FunctionCallRequested(props.ui.action, nextRID()));
}

function onKeydownEnterOrSpace() {
	if (!props.ui.action) return;
	serviceAdapter.sendEvent(new FunctionCallRequested(props.ui.action, nextRID()));
}

function loadStyles() {
	cssStyles.setStyles(defaultStyles.value, hoverStyles.value, focusStyles.value, activeStyles.value);
}

function init() {
	loadStyles();
	watch(props.ui, loadStyles, { deep: true });
}

init();
onUnmounted(() => cssStyles.remove());
</script>

<template>
	<!-- hstack -->
	<div
		v-if="
			!(props.ui instanceof HStack && props.ui.url) &&
			(props.ui.stylePreset === StylePresetValues.StyleNone || props.ui.stylePreset === undefined) &&
			!props.ui.invisible
		"
		:id="id"
		:class="classes"
		:title="props.ui.accessibilityLabel"
		:tabindex="focusable ? 0 : -1"
		@click.stop="onClick"
		@keydown.enter.stop="onKeydownEnterOrSpace"
		@keydown.space.stop="onKeydownEnterOrSpace"
	>
		<ui-generic v-for="ui in props.ui.children?.value" :ui="ui" />
	</div>

	<button
		v-else-if="
			!(props.ui instanceof HStack && props.ui.url) &&
			props.ui.stylePreset !== StylePresetValues.StyleNone &&
			props.ui.stylePreset !== undefined &&
			(props.ui.invisible === undefined || !props.ui.invisible)
		"
		:id="id"
		:tabindex="focusable ? 0 : -1"
		:disabled="props.ui.disabled"
		:class="classes"
		:title="props.ui.accessibilityLabel"
		@click.stop="onClick"
	>
		<ui-generic v-for="ui in props.ui.children?.value" :ui="ui" />
	</button>

	<a
		v-else-if="props.ui instanceof HStack && props.ui.url"
		:id="id"
		:class="classes"
		:href="props.ui.url"
		:target="props.ui.target"
		:title="props.ui.accessibilityLabel"
	>
		<ui-generic v-for="ui in props.ui.children?.value" :ui="ui" />
	</a>
</template>

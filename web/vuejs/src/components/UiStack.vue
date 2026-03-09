<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed } from 'vue';
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
import { CssClasses } from '@/shared/cssClasses';
import type { Stack } from '@/shared/proto/nprotoc_gen';
import {
	AlignmentValues,
	AnimationValues,
	FunctionCallRequested,
	Img,
	OrientationValues,
	StylePresetValues,
} from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Stack;
}>();

const id = props.ui.id || randomStr(16);
const serviceAdapter = useServiceAdapter();

const focusable = computed<boolean>(
	() => !!props.ui.action || !!props.ui.focusedBorder || !!props.ui.backgroundColorStates?.focus
);

const classes = computed<string>(() => {
	const classes = ['inline-flex'];
	if (props.ui.orientation === OrientationValues.Vertical) classes.push('flex-col');
	if (!props.ui.noClip) classes.push('overflow-clip');
	else classes.push('overflow-visible');
	if (props.ui.action) classes.push('cursor-pointer');
	if (props.ui.orientation === OrientationValues.Horizontal && props.ui.wrap) classes.push('flex-wrap');
	classes.push(defaultClass.value);
	if (activeClass.value) {
		classes.push(activeClass.value);
		classes.push('custom-active');
	}
	if (focusClass.value) {
		classes.push(focusClass.value);
		classes.push('custom-focus');
	}
	if (hoverClass.value) {
		classes.push(hoverClass.value);
		classes.push('custom-hover');
	}
	classes.push(...getAlignmentClasses());
	classes.push(...getPresetClasses());

	const animationClass = getAnimationClass();
	if (animationClass) classes.push(animationClass);

	return classes.join(' ');
});

const activeClass = computed<string | undefined>(() => {
	const styles: string[] = [];
	if (props.ui.pressedBorder) styles.push(...borderCSS(props.ui.pressedBorder));
	if (props.ui.backgroundColorStates?.pressed)
		styles.push(`background-color: ${colorValue(props.ui.backgroundColorStates.pressed)}`);

	if (!styles.length) return;
	return CssClasses.getOrCreate(styles, 'active');
});

const defaultClass = computed<string>(() => {
	const styles = frameCSS(props.ui.frame);
	styles.push(...borderCSS(props.ui.border));
	styles.push(...positionCSS(props.ui.position));
	styles.push(...transformationCSS(props.ui.transformation));
	styles.push(...backgroundCSS(props.ui.background));
	styles.push(`background-color: ${colorValue(props.ui.backgroundColor)}`);
	styles.push(...paddingCSS(props.ui.padding));
	styles.push(...fontCSS(props.ui.font));

	if (props.ui.textColor) styles.push(`color: ${colorValue(props.ui.textColor)}`);
	if (props.ui.opacity) styles.push(`opacity: ${100 - props.ui.opacity}%`);
	if (props.ui.gap) styles.push(`column-gap:${cssLengthValue(props.ui.gap)}`);
	if ((props.ui.orientation !== OrientationValues.Horizontal || props.ui.wrap) && props.ui.gap)
		styles.push(`row-gap:${cssLengthValue(props.ui.gap)}`);

	return CssClasses.getOrCreate(styles);
});

const focusClass = computed<string | undefined>(() => {
	const styles: string[] = [];
	if (props.ui.focusedBorder) styles.push(...borderCSS(props.ui.focusedBorder));
	if (props.ui.backgroundColorStates?.focus)
		styles.push(`background-color: ${colorValue(props.ui.backgroundColorStates.focus)}`);

	if (!styles.length) return;
	return CssClasses.getOrCreate(styles, 'focus');
});

const hoverClass = computed<string | undefined>(() => {
	const styles: string[] = [];
	if (props.ui.hoveredBorder) styles.push(...borderCSS(props.ui.hoveredBorder));
	if (props.ui.backgroundColorStates?.hover)
		styles.push(`background-color: ${colorValue(props.ui.backgroundColorStates.hover)}`);

	if (!styles.length) return;
	return CssClasses.getOrCreate(styles, 'hover');
});

function getAlignmentClasses(): string[] {
	if (props.ui.orientation === OrientationValues.Horizontal) {
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

function onClick(event: MouseEvent) {
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
</script>

<template>
	<div
		v-if="
			!(props.ui.orientation === OrientationValues.Horizontal && props.ui.url) &&
			(props.ui.stylePreset === StylePresetValues.StyleNone || props.ui.stylePreset === undefined) &&
			!props.ui.invisible
		"
		:id="id"
		:class="classes"
		:title="props.ui.accessibilityLabel"
		:tabindex="focusable ? 0 : -1"
		@click="onClick"
		@keydown="onKeydown"
	>
		<ui-generic v-for="ui in props.ui.children?.value" :ui="ui" />
	</div>

	<button
		v-else-if="
			!(props.ui.orientation === OrientationValues.Horizontal && props.ui.url) &&
			props.ui.stylePreset !== StylePresetValues.StyleNone &&
			props.ui.stylePreset !== undefined &&
			(props.ui.invisible === undefined || !props.ui.invisible)
		"
		:id="id"
		:tabindex="focusable ? 0 : -1"
		:disabled="props.ui.disabled"
		:class="classes"
		:title="props.ui.accessibilityLabel"
		@click="onClick"
	>
		<ui-generic v-for="ui in props.ui.children?.value" :ui="ui" />
	</button>

	<a
		v-else-if="props.ui.orientation === OrientationValues.Horizontal && props.ui.url"
		:id="id"
		:class="classes"
		:href="props.ui.url"
		:target="props.ui.target"
		:title="props.ui.accessibilityLabel"
	>
		<ui-generic v-for="ui in props.ui.children?.value" :ui="ui" />
	</a>
</template>

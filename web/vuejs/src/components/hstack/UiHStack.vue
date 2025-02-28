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
import { AlignmentValues, FunctionCallRequested, HStack, Img, StylePresetValues } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: HStack;
}>();

const hover = ref(false);
const pressed = ref(false);
const focused = ref(false);
const focusable = ref(false);
const focusVisible = ref(false);
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

function checkFocusVisible(event: Event) {
	const element = event.target as HTMLElement;
	focusVisible.value = element.matches(':focus-visible');
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

	if (focusVisible.value) {
		styles.push('outline: 2px solid black'); // always apply solid and never auto. Auto will create random broken effects on firefox and chrome
	}

	return styles;
}

const frameStyles = computed<string>(() => {
	let styles = commonStyles();

	if (props.ui.gap) {
		styles.push(`column-gap:${cssLengthValue(props.ui.gap)}`);
	}

	return styles.join(';');
});

const clazz = computed<string>(() => {
	let classes = ['inline-flex'];
	switch (props.ui.alignment) {
		case AlignmentValues.Stretch:
			classes.push('items-stretch');
			break;
		case AlignmentValues.Leading:
			classes.push('justify-start', 'items-center');
			break;
		case AlignmentValues.Trailing:
			classes.push('justify-end', 'items-center');
			break;
		case AlignmentValues.Center:
			classes.push('justify-center', 'items-center');
			break;
		case AlignmentValues.TopLeading:
			classes.push('justify-start', 'items-start');
			break;
		case AlignmentValues.BottomLeading:
			classes.push('justify-start', 'items-end');
			break;
		case AlignmentValues.TopTrailing:
			classes.push('justify-end', 'items-start');
			break;
		case AlignmentValues.Top:
			classes.push('justify-center', 'items-start');
			break;
		case AlignmentValues.BottomTrailing:
			classes.push('justify-end', 'items-end');
			break;
		case AlignmentValues.Bottom:
			classes.push('justify-center', 'items-end');
			break;
		default:
			classes.push('justify-center', 'items-center');
			break;
	}

	if (props.ui.action) {
		classes.push('cursor-pointer');
	}

	if (props.ui.wrap) {
		classes.push('flex-wrap');
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

	// preset special round icon mode in buttons
	if (props.ui.stylePreset) {
		if (props.ui.children?.value.length == 1 && props.ui.children.value[0] instanceof Img) {
			classes.push('!p-0', '!w-10');
		}
	}

	return classes.join(' ');
});
</script>

<template>
	<!-- hstack -->
	<div
		:id="ui.id"
		v-if="
			(props.ui.stylePreset === StylePresetValues.StyleNone || props.ui.stylePreset === undefined) &&
			!props.ui.invisible
		"
		:class="clazz"
		:style="frameStyles"
		@mouseover="hover = true"
		@mouseleave="hover = false"
		@mousedown="pressed = true"
		@mouseup="pressed = false"
		@mouseout="pressed = false"
		@focusin="focused = true"
		:title="props.ui.accessibilityLabel"
		@focusout="
			focused = false;
			focusVisible = false;
		"
		:tabindex="focusable ? 0 : -1"
		@click="onClick"
		@keydown="onKeydown"
		@focus="checkFocusVisible"
	>
		<ui-generic v-for="ui in props.ui.children?.value" :ui="ui" />
	</div>

	<button
		:id="ui.id"
		:disabled="props.ui.disabled"
		v-else-if="
			props.ui.stylePreset !== StylePresetValues.StyleNone &&
			props.ui.stylePreset !== undefined &&
			(props.ui.invisible === undefined || !props.ui.invisible)
		"
		:class="clazz"
		:style="frameStyles"
		@click="onClick"
		:title="props.ui.accessibilityLabel"
	>
		<ui-generic v-for="ui in props.ui.children?.value" :ui="ui" />
	</button>
</template>

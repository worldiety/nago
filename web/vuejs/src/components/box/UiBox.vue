<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { borderCSS } from '@/components/shared/border';
import { frameCSS } from '@/components/shared/frame';
import { cssLengthValue0Px } from '@/components/shared/length';
import { paddingCSS } from '@/components/shared/padding';
import { Alignment, AlignmentValues, Box, Padding } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Box;
}>();

const frameStyles = computed<string>(() => {
	let styles = frameCSS(props.ui.frame);
	if (props.ui.backgroundColor) {
		styles.push(`background-color: ${props.ui.backgroundColor}`);
	}

	styles.push(...borderCSS(props.ui.border));
	styles.push(...paddingCSS(props.ui.padding));

	return styles.join(';');
});

const clazz = computed<string>(() => {
	const classes = ['relative', 'flex'];
	if (props.ui.disableOutsidePointerEvents) {
		classes.push('pointer-events-none');
	}
	return classes.join(' ');
});

function childMargin(a?: Alignment, p?: Padding): string {
	switch (a) {
		case AlignmentValues.BottomLeading:
			return `margin-left: ${cssLengthValue0Px(p?.left)};margin-bottom: ${cssLengthValue0Px(p?.bottom)}`;
		case AlignmentValues.TopLeading:
			return `margin-top: ${cssLengthValue0Px(p?.top)};margin-left: ${cssLengthValue0Px(p?.left)}`;
		case AlignmentValues.Leading:
			return `margin-left: ${cssLengthValue0Px(p?.left)}`;
		case AlignmentValues.Top:
			return `margin-top: ${cssLengthValue0Px(p?.top)}`;
		case AlignmentValues.Bottom:
			return `margin-bottom: ${cssLengthValue0Px(p?.bottom)}`;
		case AlignmentValues.Trailing:
			return `margin-right: ${cssLengthValue0Px(p?.right)}`;
		case AlignmentValues.BottomTrailing:
			return `margin-right: ${cssLengthValue0Px(p?.right)};margin-bottom: ${cssLengthValue0Px(p?.bottom)}`;
		case AlignmentValues.TopTrailing:
			return `margin-right: ${cssLengthValue0Px(p?.right)};margin-top: ${cssLengthValue0Px(p?.top)}`;
	}

	return '';
}

function childClass(a?: Alignment): string {
	// we also use flex for the simple cases, because otherwise we have a gap and even more weired rendering if not enough room
	// note, that flex never calculates the width properly, even with border-box etc.
	// we will use margin instead
	switch (a) {
		case AlignmentValues.BottomLeading:
			return 'absolute flex bottom-0 left-0';
		case AlignmentValues.TopLeading:
			return 'absolute flex top-0 left-0';
		case AlignmentValues.TopTrailing:
			return 'absolute flex top-0 right-0';
		case AlignmentValues.BottomTrailing:
			return 'absolute flex right-0 bottom-0';
		case AlignmentValues.Top:
			return 'absolute w-full flex justify-center top-0';
		case AlignmentValues.Bottom:
			return 'absolute w-full flex justify-center bottom-0';
		case AlignmentValues.Leading:
			return 'absolute h-full flex items-center left-0';
		case AlignmentValues.Trailing:
			return 'absolute h-full flex items-center right-0';
		default:
			return 'absolute w-full h-full flex justify-center items-center';
	}
}
</script>

<template v-if="props.ui.children">
	<!-- box -->
	<div :class="clazz" :style="frameStyles">
		<div
			v-for="ui in props.ui.children?.value"
			:class="childClass(ui.alignment)"
			:style="childMargin(ui.alignment, props.ui.padding)"
		>
			<ui-generic v-if="ui.component" :ui="ui.component" class="pointer-events-auto" />
		</div>
	</div>
</template>

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
import { colorValue } from '@/components/shared/colors';
import { frameCSS } from '@/components/shared/frame';
import { paddingCSS } from '@/components/shared/padding';
import { positionCSS } from '@/components/shared/position';
import { HoverGroup } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: HoverGroup;
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

// note that we need the max-content hack, otherwise we get layout bugs at least for horizontal areas
</script>

<template v-if="props.ui.iv">
	<!-- UiHoverGroup -->
	<div class="group relative" :style="styles">
		<UiGeneric
			v-if="ui.content"
			:ui="ui.content"
			class="absolute transition-opacity duration-300 opacity-100 group-hover:opacity-0"
		/>
		<UiGeneric
			v-if="ui.contentHover"
			:ui="ui.contentHover"
			class="absolute transition-opacity duration-300 opacity-0 group-hover:opacity-100"
		/>
	</div>
</template>

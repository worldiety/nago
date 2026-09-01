<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template>
	<fieldset :name="ui.title" :style="frameStyles">
		<legend v-if="ui.title" class="font-headline-small">
			{{ ui.title }}
		</legend>
		<ui-generic v-for="child in ui.children?.value" :ui="child" />
	</fieldset>
</template>
<script setup lang="ts">
import { computed } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { frameCSS } from '@/components/shared/frame';
import { Fieldset } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Fieldset;
}>();

const frameStyles = computed<string>(() => {
	let styles = frameCSS(props.ui.frame);

	return styles.join(';');
});
</script>
<style scoped>
fieldset {
	@apply bg-M2 rounded-xl p-5 -mt-5;

	legend {
		@apply inline-block pb-2 mb-10 border-b border-M5 translate-y-full;
	}
}
</style>

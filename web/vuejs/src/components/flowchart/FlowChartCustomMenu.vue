<!--
 Copyright (c) 2026 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template>
	<div ref="container" class="custom-node style-none">
		<div class="opacity-0 pointer-events-none h-px">.</div> <!-- This is needed for some reason... -->
		<div>
			<div v-if="zoom" :style="`transform: scale(${1 / zoom});`">
				<div ref="menu" :style="floatingStyles">
					<UiGeneric :ui="content" />
				</div>
			</div>
		</div>
	</div>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import {
	Component,
} from '@/shared/proto/nprotoc_gen';
import { autoUpdate, offset, shift, useFloating } from '@floating-ui/vue';

interface Props {
	content: Component;
	zoom?: number;
}

defineProps<Props>();

const container = ref();
const menu = ref();

const { floatingStyles } = useFloating(container, menu, {
	placement: 'bottom-start',
	strategy: 'fixed',
	whileElementsMounted: autoUpdate,
	middleware: [offset(8), shift({ crossAxis: true })],
});
</script>

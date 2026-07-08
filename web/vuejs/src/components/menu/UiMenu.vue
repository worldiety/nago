<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed, ref } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { frameCSS } from '@/components/shared/frame';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { autoUpdate, flip, offset, shift, useFloating } from '@floating-ui/vue';
import { Menu, MenuButton, MenuItem, MenuItems } from '@headlessui/vue';
import type { Menu as ProtoMenu, MenuItem as ProtoMenuItem } from '@/shared/proto/nprotoc_gen';
import { FunctionCallRequested } from '@/shared/proto/nprotoc_gen';
import { pxLengthValue } from '@/components/shared/length';

const props = defineProps<{
	ui: ProtoMenu;
}>();

const serviceAdapter = useServiceAdapter();

const trigger = ref();
const menu = ref();

const { floatingStyles } = useFloating(trigger, menu, {
	placement: 'bottom-start',
	strategy: 'fixed',
	whileElementsMounted: autoUpdate,
	middleware: [flip(), shift({ crossAxis: true }), offset(pxLengthValue(props.ui.offset))],
});

function itemClick(item: ProtoMenuItem) {
	if (item.action) {
		serviceAdapter.sendEvent(new FunctionCallRequested(item.action, nextRID()));
		return;
	}
}

const styles = computed<string>(() => {
	const styles = frameCSS(props.ui.frame);
	return styles.join(';');
});
</script>

<template>
	<Menu as="div" class="relative inline-block text-left" :style="styles">
		<div>
			<MenuButton ref="trigger" class="inline-flex w-full justify-center">
				<ui-generic v-if="props.ui.anchor" :ui="props.ui.anchor" />
			</MenuButton>
		</div>

		<MenuItems
			ref="menu"
			class="z-40 min-w-56 max-h-screen divide-y divide-M3 rounded-md bg-M1 shadow-lg ring-1 ring-black/5 focus:outline-none border border-M3 overflow-y-auto"
			:style="floatingStyles"
		>
			<div v-for="section in props.ui.groups?.value" class="px-1 py-1">
				<template v-if="section.customContent">
					<UiGeneric :ui="section.customContent" />
				</template>
				<template v-else>
					<MenuItem v-for="itemUi in section.items?.value" v-slot="{ active }">
						<button
							:class="[
								active ? 'bg-I0 bg-opacity-25' : '',
								'group flex w-full items-center rounded-md px-2 py-2 text-sm',
							]"
							@click="itemClick(itemUi)"
						>
							<ui-generic v-if="itemUi.content" :ui="itemUi.content" />
						</button>
					</MenuItem>
				</template>
			</div>
		</MenuItems>
	</Menu>
</template>

<template>
	<div
		class="flex flex-col justify-between items-center cursor-pointer h-full"
		tabindex="0"
		@mousedown="active = true"
		@click="handleClick"
		@keydown.enter="handleClick"
		@keydown.down.prevent="focusFirstLinkedSubMenuEntry('down')"
		@keydown.right.prevent="focusFirstLinkedSubMenuEntry('right')"
		@mouseenter="expandMenuEntry"
		@mouseleave="handleMouseLeave"
		@mouseup="active = false"
		@focus="expandMenuEntry"
	>
		<div
			class="flex justify-center items-center grow shrink rounded-full py-2 w-full"
			:class="{'bg-disabled-background bg-opacity-25': ui.expanded.v, 'bg-opacity-35': active}"
		>
			<div class="relative w-4">
				<div v-if="ui.expanded.v" class="*:h-full" v-html="ui.iconActive.v"></div>
				<div v-else class="*:h-full" v-html="ui.icon.v"></div>
				<!-- Optional red badge -->
				<div v-if="ui.badge.v" class="absolute -top-1.5 -right-1.5 flex justify-center items-center h-3.5 px-1 rounded-full bg-error">
					<p class="text-xs text-white">{{ ui.badge.v }}</p>
				</div>
			</div>
		</div>
		<p class="text-sm text-center font-medium select-none">{{ ui.title.v }}</p>
	</div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import type { MenuEntry } from '@/shared/protocol/ora/menuEntry';

const emit = defineEmits<{
	(e: 'focusFirstLinkedSubMenuEntry'): void;
}>();

const props = defineProps<{
	ui: MenuEntry;
	menuEntryIndex: number;
	mode: 'navigationBar'|'sidebar';
}>();

const serviceAdapter = useServiceAdapter();
const active = ref<boolean>(false);

const hasSubMenuEntries = computed((): boolean => {
	return props.ui.menu.v && props.ui.menu.v.length > 0;
});

function handleClick(): void {
	if (props.ui.action.v && !hasSubMenuEntries.value) {
		// If the menu entry has an action and no sub menu entries, execute the action
		serviceAdapter.executeFunctions(props.ui.action);
	} else {
		// Else expand the menu entry
		expandMenuEntry();
	}
}

function handleMouseLeave(): void {
	active.value = false;
	if (!hasSubMenuEntries.value) {
		// Collapse the menu entry if it has no sub menu entries
		serviceAdapter.setProperties({
			...props.ui.expanded,
			v: false,
		});
	}
}

function expandMenuEntry(): void {
	serviceAdapter.setPropertiesAndCallFunctions([
		{
			...props.ui.expanded,
			v: true,
		}
	], [props.ui.onFocus])
}

function focusFirstLinkedSubMenuEntry(keyPressed: 'down'|'right'): void {
	if (!props.ui.menu.v || props.ui.menu.v.length === 0) {
		return;
	}
	if (props.mode === 'navigationBar' && keyPressed === 'down' || props.mode === 'sidebar' && keyPressed === 'right') {
		emit('focusFirstLinkedSubMenuEntry');
	}
}
</script>

<template>
	<div
		class="flex flex-col justify-center items-center cursor-pointer"
		:class="{'menu-entry-linked': ui.action.v && !hasSubMenuEntries}"
		tabindex="0"
		@mousedown="active = true"
		@click="handleClick"
		@keydown.enter="handleClick"
		@keydown.down.prevent="focusFirstLinkedSubMenuEntry('down')"
		@keydown.right.prevent="focusFirstLinkedSubMenuEntry('right')"
		@mouseenter="expandMenuEntry"
		@mouseleave="active = false"
		@mouseup="active = false"
		@focus="expandMenuEntry"
	>
		<div
			class="flex justify-center items-center rounded-full py-2 w-16"
			:class="{'bg-disabled-background bg-opacity-25': ui.expanded.v, 'bg-opacity-35': active}"
		>
			<div v-if="ui.expanded.v" class="h-4 *:h-full" v-html="ui.iconActive.v"></div>
			<div v-else class="h-4 *:h-full" v-html="ui.icon.v"></div>
		</div>
		<p class="menu-entry-title text-sm text-center font-medium select-none">{{ ui.title.v }}</p>
	</div>
</template>

<script setup lang="ts">
import type { MenuEntry } from '@/shared/protocol/gen/menuEntry';
import { computed, ref } from 'vue';
import { useServiceAdapter } from '@/composables/serviceAdapter';

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
		serviceAdapter.executeFunctions(props.ui.action);
	} else {
		expandMenuEntry();
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

<style scoped>
.menu-entry-linked:hover .menu-entry-title,
.menu-entry-linked:focus-visible .menu-entry-title {
	@apply underline;
}
</style>

<template>
	<div
		class="flex flex-col justify-center items-center cursor-pointer"
		:class="{'menu-entry-linked': ui.action.v}"
		tabindex="0"
		@mousedown="active = true"
		@click="handleClick"
		@keydown.enter="handleClick"
		@keydown.down.prevent="focusFirstLinkedSubMenuEntry"
		@mouseenter="expandMenuEntry"
		@mouseleave="handleMouseLeave"
		@mouseup="active = false"
		@focus="expandMenuEntry"
	>
		<div
			class="flex justify-center items-center rounded-full py-2 w-16"
			:class="{'bg-disabled-background bg-opacity-25': expanded, 'bg-opacity-35': active}"
		>
			<div v-if="expanded" class="h-4 *:h-full" v-html="ui.iconActive.v"></div>
			<div v-else class="h-4 *:h-full" v-html="ui.icon.v"></div>
		</div>
		<p class="menu-entry-title text-sm text-center font-medium select-none">{{ ui.title.v }}</p>
	</div>
</template>

<script setup lang="ts">
import type { MenuEntry } from '@/shared/protocol/gen/menuEntry';
import { ref } from 'vue';
import { useServiceAdapter } from '@/composables/serviceAdapter';

const emit = defineEmits<{
	(e: 'expandMenuEntry', menuEntry: MenuEntry, menuEntryIndex: number): void;
	(e: 'collapseMenuEntry'): void;
	(e: 'focusFirstLinkedSubMenuEntry'): void;
}>();

const props = defineProps<{
	ui: MenuEntry;
	menuEntryIndex: number;
	expanded: boolean;
}>();

const serviceAdapter = useServiceAdapter();
const active = ref<boolean>(false);

function handleClick(): void {
	if (props.ui.action.v) {
		serviceAdapter.executeFunctions(props.ui.action);
	} else {
		expandMenuEntry();
	}
}

function expandMenuEntry(): void {
	emit('expandMenuEntry', props.ui, props.menuEntryIndex);
}

function handleMouseLeave(): void {
	active.value = false;
	if (!props.ui.menu.v || props.ui.menu.v.length === 0) {
		emit('collapseMenuEntry');
	}
}

function focusFirstLinkedSubMenuEntry(): void {
	if (props.ui.menu.v?.length > 0) {
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

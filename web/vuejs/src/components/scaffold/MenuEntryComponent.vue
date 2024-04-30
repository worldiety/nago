<template>
	<div
		class="flex flex-col justify-center items-center cursor-pointer"
		tabindex="0"
		@mousedown="active = true"
		@mouseenter="setHovered"
		@mouseleave="resetState"
		@mouseup="active = false"
	>
		<div
			class="flex justify-center items-center rounded-full py-2 w-16"
			:class="{'bg-disabled-background bg-opacity-25': hovered, 'bg-opacity-35': active}"
		>
			<div v-if="hovered" class="h-4 *:h-full" v-html="ui.iconActive.v"></div>
			<div v-else class="h-4 *:h-full" v-html="ui.icon.v"></div>
		</div>
		<p class="text-sm font-medium select-none">{{ ui.title.v }}</p>
	</div>
</template>

<script setup lang="ts">
import type { MenuEntry } from '@/shared/protocol/gen/menuEntry';
import { ref } from 'vue';

const emit = defineEmits<{
	(e: 'menuEntryHovered', menuEntry: MenuEntry): void;
}>();

const props = defineProps<{
	ui: MenuEntry;
}>();

const hovered = ref<boolean>(false);
const active = ref<boolean>(false);

function setHovered(): void {
	hovered.value = true;
	emit('menuEntryHovered', props.ui);
}

function resetState(): void {
	hovered.value = false;
	active.value = false;
}
</script>

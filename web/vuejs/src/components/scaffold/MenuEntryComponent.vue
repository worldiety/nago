<template>
	<div
		class="flex flex-col justify-center items-center cursor-pointer"
		tabindex="0"
		@mousedown="active = true"
		@click="expandMenuEntry"
		@mouseenter="expandMenuEntry"
		@mouseleave="active = false"
		@mouseup="active = false"
	>
		<div
			class="flex justify-center items-center rounded-full py-2 w-16"
			:class="{'bg-disabled-background bg-opacity-25': expanded, 'bg-opacity-35': active}"
		>
			<div v-if="expanded" class="h-4 *:h-full" v-html="ui.iconActive.v"></div>
			<div v-else class="h-4 *:h-full" v-html="ui.icon.v"></div>
		</div>
		<p class="text-sm font-medium select-none">{{ ui.title.v }}</p>
	</div>
</template>

<script setup lang="ts">
import type { MenuEntry } from '@/shared/protocol/gen/menuEntry';
import { ref } from 'vue';

const emit = defineEmits<{
	(e: 'expandMenuEntry', menuEntry: MenuEntry, menuEntryIndex: number): void;
}>();

const props = defineProps<{
	ui: MenuEntry;
	menuEntryIndex: number;
	expanded: boolean;
}>();

const active = ref<boolean>(false);

function expandMenuEntry(): void {
	emit('expandMenuEntry', props.ui, props.menuEntryIndex);
}
</script>

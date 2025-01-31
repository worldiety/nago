<template>
	<div
		class="flex flex-col justify-between items-center cursor-pointer h-full w-full"
		tabindex="0"
		@mousedown="interacted = true"
		@click="handleClick"
		@keydown.enter="handleClick"
		@keydown.down.prevent="focusFirstLinkedSubMenuEntry('down')"
		@keydown.right.prevent="focusFirstLinkedSubMenuEntry('right')"
		@mouseenter="
			emit('expand', ui);
			hover = true;
		"
		@mouseleave="
			handleMouseLeave;
			hover = false;
		"
		@mouseup="interacted = false"
		@focus="emit('expand', ui)"
	>
		<div
			v-if="ui.i"
			class="flex justify-center items-center grow shrink rounded-full py-2 w-full"
			:class="{
				'h-10': !ui.t,
				'mix-blend-multiply bg-M7': !ui.t && hover,
				'bg-M7 bg-opacity-25': ui.x,
				'bg-opacity-35': interacted,
				'bg-M7 bg-opacity-35': active,
			}"
		>
			<div class="relative">
				<div class="*:h-full" v-if="ui.x && ui.v">
					<ui-generic :ui="props.ui.v" />
				</div>
				<div v-else-if="ui.t" class="*:h-full">
					<ui-generic :ui="props.ui.i" />
				</div>
				<div v-else class="h-10">
					<ui-generic :ui="props.ui.i" />
				</div>

				<!-- Optional red badge -->
				<div
					v-if="ui.b"
					class="absolute -top-1.5 -right-1.5 flex justify-center items-center h-3.5 px-1 rounded-full bg-A0"
				>
					<p class="text-xs text-white">{{ ui.b }}</p>
				</div>
			</div>
		</div>
		<p
			class="text-sm text-center font-medium select-none hyphens-auto w-full"
			:class="{ 'font-semibold': linksToCurrentPage }"
		>
			{{ ui.t }}
		</p>
	</div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import type { MenuEntry } from '@/shared/protocol/ora/menuEntry';
import type { SVG } from '@/shared/protocol/ora/sVG';
import { ScaffoldMenuEntry } from '@/shared/protocol/ora/scaffoldMenuEntry';

const emit = defineEmits<{
	(e: 'focusFirstLinkedSubMenuEntry'): void;
	(e: 'expand', menuEntry: ScaffoldMenuEntry): void;
}>();

const props = defineProps<{
	ui: ScaffoldMenuEntry;
	menuEntryIndex: number;
	mode: 'navigationBar' | 'sidebar';
}>();

const serviceAdapter = useServiceAdapter();
const interacted = ref<boolean>(false);
const hover = ref<boolean>(false);

const linksToCurrentPage = computed((): boolean => {
	if (props.ui.f == '.' && (window.location.pathname == '' || window.location.pathname == '/')) {
		return true;
	}

	return `/${props.ui.f}` === window.location.pathname;
});

const active = computed((): boolean => {
	return interacted.value || linksToCurrentPage.value;
});

const hasSubMenuEntries = computed((): boolean => {
	return !!(props.ui.m && props.ui.m.length > 0);
});

function handleClick(): void {
	if (props.ui.a && !hasSubMenuEntries.value) {
		// If the menu entry has an action and no sub menu entries, execute the action
		serviceAdapter.executeFunctions(props.ui.a);
	} else {
		// Else expand the menu entry
		emit('expand', props.ui);
	}
}

function handleMouseLeave(): void {
	interacted.value = false;
	if (!hasSubMenuEntries.value) {
		// Collapse the menu entry if it has no sub menu entries
		/*serviceAdapter.setProperties({
			...props.ui.x,
			v: false,
		});*/
		props.ui.x = false;
	}
}

function focusFirstLinkedSubMenuEntry(keyPressed: 'down' | 'right'): void {
	if (!props.ui.m || props.ui.m.length === 0) {
		return;
	}
	if (
		(props.mode === 'navigationBar' && keyPressed === 'down') ||
		(props.mode === 'sidebar' && keyPressed === 'right')
	) {
		emit('focusFirstLinkedSubMenuEntry');
	}
}
</script>

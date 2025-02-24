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
			v-if="ui.icon"
			class="flex justify-center items-center grow shrink rounded-full py-2 w-full"
			:class="{
				'h-10': ui.title.isZero(),
				'mix-blend-multiply bg-M7': ui.isZero() && hover,
				'bg-M7 bg-opacity-25': ui.expanded.value,
				'bg-opacity-35': interacted,
				'bg-M7 bg-opacity-35': active,
			}"
		>
			<div class="relative">
				<div class="*:h-full" v-if="ui.expanded.value && ui.iconActive">
					<ui-generic :ui="props.ui.iconActive!" />
				</div>
				<div v-else-if="!ui.title.isZero() && props.ui.icon" class="*:h-full">
					<ui-generic :ui="props.ui.icon" />
				</div>
				<div v-else-if="props.ui.icon" class="h-10">
					<ui-generic :ui="props.ui.icon" />
				</div>

				<!-- Optional red badge -->
				<div
					v-if="!ui.badge.isZero()"
					class="absolute -top-1.5 -right-1.5 flex justify-center items-center h-3.5 px-1 rounded-full bg-A0"
				>
					<p class="text-xs text-white">{{ ui.badge.value }}</p>
				</div>
			</div>
		</div>
		<p
			class="text-sm text-center font-medium select-none hyphens-auto w-full"
			:class="{ 'font-semibold': linksToCurrentPage }"
		>
			{{ ui.title.value }}
		</p>
	</div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import {FunctionCallRequested, ScaffoldMenuEntry} from "@/shared/proto/nprotoc_gen";
import {nextRID} from "@/eventhandling";

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
	if (props.ui.rootView.value == '.' && (window.location.pathname == '' || window.location.pathname == '/')) {
		return true;
	}

	return `/${props.ui.rootView.value}` === window.location.pathname;
});

const active = computed((): boolean => {
	return interacted.value || linksToCurrentPage.value;
});

const hasSubMenuEntries = computed((): boolean => {
	return (props.ui.menu.value && props.ui.menu.value.length > 0);
});

function handleClick(): void {
	if (!props.ui.action.isZero() && !hasSubMenuEntries.value) {
		// If the menu entry has an action and no sub menu entries, execute the action
		serviceAdapter.sendEvent(new FunctionCallRequested(
			props.ui.action,
			nextRID()
		));
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
		props.ui.expanded.value = false;
	}
}

function focusFirstLinkedSubMenuEntry(keyPressed: 'down' | 'right'): void {
	if (!props.ui.menu.value || props.ui.menu.value.length === 0) {
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

<template>
	<div
		class="flex flex-col justify-between items-center cursor-pointer h-full w-full"
		tabindex="0"
		@mousedown="interacted = true"
		@click="handleClick"
		@keydown.enter="handleClick"
		@keydown.down.prevent="focusFirstLinkedSubMenuEntry('down')"
		@keydown.right.prevent="focusFirstLinkedSubMenuEntry('right')"
		@mouseenter="handleMouseEnter"
		@mouseleave="handleMouseLeave"
		@mouseup="interacted = false"
		@focus="emit('expand', ui)"
	>
		<!-- icon -->
		<div v-if="ui.icon" class="flex justify-center items-center grow shrink rounded-full py-2" :class="iconClasses">
			<div class="relative">
				<div v-if="ui.expanded && ui.iconActive" class="*:h-full">
					<ui-generic :ui="props.ui.iconActive!" />
				</div>
				<div v-else-if="ui.title && props.ui.icon" class="*:h-full">
					<ui-generic :ui="props.ui.icon" />
				</div>
				<div v-else-if="props.ui.icon" class="h-10">
					<ui-generic :ui="props.ui.icon" />
				</div>

				<!-- Optional red badge -->
				<div
					v-if="ui.badge"
					class="absolute -top-1.5 -right-1.5 flex justify-center items-center h-3.5 px-1 rounded-full bg-A0"
				>
					<p class="text-xs text-white">{{ ui.badge }}</p>
				</div>
			</div>
		</div>

		<!-- title -->
		<p
			class="text-sm text-center font-medium select-none whitespace-nowrap w-full"
			:class="{ 'font-semibold': linksToCurrentPage }"
		>
			{{ ui.title }}
		</p>
	</div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import type { ScaffoldMenuEntry } from '@/shared/proto/nprotoc_gen';
import { FunctionCallRequested } from '@/shared/proto/nprotoc_gen';

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
	if (props.ui.rootView == '.' && (window.location.pathname == '' || window.location.pathname == '/')) {
		return true;
	}

	return `/${props.ui.rootView}` === window.location.pathname;
});

const active = computed((): boolean => {
	return interacted.value || linksToCurrentPage.value;
});

const hasSubMenuEntries = computed((): boolean => {
	if (!props.ui.menu) {
		return false;
	}

	return props.ui.menu.value && props.ui.menu.value.length > 0;
});

const iconClasses = computed((): string => {
	const iconClasses: string[] = [];
	if (props.ui.title === undefined) {
		iconClasses.push('size-12');
	} else {
		iconClasses.push('h-10 w-16');
	}
	if (props.ui.isZero() && hover.value) {
		iconClasses.push('mix-blend-multiply', 'bg-M7');
	}
	if (props.ui.expanded) {
		iconClasses.push('bg-M7', 'bg-opacity-25');
	}
	if (interacted.value) {
		iconClasses.push('bg-opacity-35');
	}
	if (active.value) {
		iconClasses.push('bg-M7', 'bg-opacity-35');
	}
	return iconClasses.join(' ');
});

function handleClick(): void {
	if (props.ui.action && !hasSubMenuEntries.value) {
		// If the menu entry has an action and no sub menu entries, execute the action
		serviceAdapter.sendEvent(new FunctionCallRequested(props.ui.action, nextRID()));
	} else {
		// Else expand the menu entry
		emit('expand', props.ui);
	}
}

function handleMouseEnter(): void {
	emit('expand', props.ui);
	hover.value = true;
}

function handleMouseLeave(): void {
	interacted.value = false;
	hover.value = false;
	if (!hasSubMenuEntries.value) {
		// Collapse the menu entry if it has no sub menu entries
		/*serviceAdapter.setProperties({
			...props.ui.x,
			v: false,
		});*/
		props.ui.expanded = false;
	}
}

function focusFirstLinkedSubMenuEntry(keyPressed: 'down' | 'right'): void {
	if (!props.ui.menu?.value || props.ui.menu.value.length === 0) {
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

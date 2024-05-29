<template>
	<div
		ref="dialogContainer"
		class="relative flex justify-center items-center pointer-events-auto bg-black bg-opacity-60 h-full"
		@keydown.tab.exact="moveFocusForward"
		@keydown.shift.tab="moveFocusBackwards"
	>
		<div class="flex flex-col text-black rounded-xl shadow-md max-h-screen" :class="dialogClasses" @click.stop>
			<!-- Dialog header -->
			<div class="flex justify-start items-center gap-x-2 bg-[#F9F9F9] rounded-t-xl px-6 py-3">
				<div v-if="ui.icon.v" v-html="ui.icon.v" class="w-6 *:h-full"></div>
				<p class="font-bold">{{ ui.title.v }}</p>
			</div>

			<!-- Dialog body -->
			<div class="bg-white pt-3.5 px-6 pb-6 overflow-y-auto" :class="{'rounded-b-xl': !ui.footer.v}">
				<UiGeneric :ui="ui.body.v" />
			</div>

			<div v-if="ui.footer.v" class="bg-white rounded-b-xl px-6 pb-6">
				<hr class="border-[#E2E2E2] pb-6" />
				<!-- Dialog footer -->
				<UiGeneric :ui="ui.footer.v" />
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import type { Dialog } from '@/shared/protocol/ora/dialog';
import UiGeneric from '@/components/UiGeneric.vue';
import { computed, onMounted, ref, watch } from 'vue';
import { ElementSize } from '@/shared/protocol/ora/elementSize';

const props = defineProps<{
	ui: Dialog;
	isActiveDialog: boolean;
}>();

const dialogContainer = ref<HTMLElement|undefined>();
let firstFocusableElement: HTMLElement|undefined;
let lastFocusableElement: HTMLElement|undefined;

onMounted(() => {
	if (props.isActiveDialog) {
		captureFocusInDialog();
	}
});

watch(() => props.isActiveDialog, (newValue) => {
	if (newValue) {
		captureFocusInDialog();
	}
});

const dialogClasses = computed((): string => {
	const dialogClasses: string[] = ['w-full'];
	switch (props.ui.size.v) {
		case ElementSize.SIZE_AUTO:
			dialogClasses.push('sm:w-auto');
			break;
		case ElementSize.SIZE_SMALL:
			dialogClasses.push('sm:w-[25rem]');
			break;
		case ElementSize.SIZE_MEDIUM:
			dialogClasses.push('sm:w-[35rem]');
			break;
		case ElementSize.SIZE_LARGE:
			dialogClasses.push('sm:w-[45rem]');
			break;
		default:
			dialogClasses.push('sm:w-auto');
			break;
	}
	return dialogClasses.join(' ');
});

function captureFocusInDialog(): void {
	const focusableElements = dialogContainer.value?.querySelectorAll('[tabindex="0"], button:not([tabindex="-1"])') ?? [];
	const firstFocusable = focusableElements[0];
	const lastFocusable = focusableElements[focusableElements.length - 1];
	if (firstFocusable) {
		firstFocusableElement = firstFocusable as HTMLElement;
		firstFocusableElement.focus();
	}
	if (lastFocusable) {
		lastFocusableElement = lastFocusable as HTMLElement;
	}
}

function moveFocusForward(e: KeyboardEvent): void {
	if (e.shiftKey) {
		return;
	}
	if (document.activeElement as HTMLElement === lastFocusableElement) {
		e.preventDefault();
		firstFocusableElement?.focus();
		return;
	}
}

function moveFocusBackwards(e: KeyboardEvent): void {
	if (document.activeElement as HTMLElement === firstFocusableElement) {
		e.preventDefault();
		lastFocusableElement?.focus();
		return;
	}
}
</script>

<template>
	<div
		v-if="ui.visible.v"
		class="relative flex justify-center items-center pointer-events-auto bg-black bg-opacity-60 h-full"
		@keydown.tab.exact="moveFocusForward"
		@keydown.shift.tab="moveFocusBackwards"
	>
		<!-- Upper focus boundary -->
		<div ref="firstFocusableElement" tabindex="0" class="size-0 focus:outline-none focus:ring-0"></div>

		<div class="text-black dark:text-white rounded-xl shadow-md overflow-y-auto max-h-screen" :class="dialogClasses" @click.stop>
			<!-- Dialog header -->
			<div class="flex justify-start items-center gap-x-2 bg-[#F9F9F9] dark:bg-black rounded-t-xl px-6 py-3">
				<div v-if="ui.icon.v" v-html="ui.icon.v" class="w-6 *:h-full"></div>
				<p class="font-bold">{{ ui.title.v }}</p>
			</div>

			<!-- Dialog body -->
			<div class="bg-white dark:bg-[#2B2B2B] pt-3.5 px-6 pb-6" :class="{'rounded-b-xl': !ui.footer.v}">
				<UiGeneric :ui="ui.body.v" />
			</div>

			<div v-if="ui.footer.v" class="bg-white dark:bg-[#2B2B2B] rounded-b-xl px-6 pb-6">
				<hr class="border-[#E2E2E2] dark:border-[#848484] pb-6" />
				<!-- Dialog footer -->
				<UiGeneric :ui="ui.footer.v" />
			</div>

			<!-- Lower focus boundary -->
			<div ref="lastFocusableElement" tabindex="0" class="size-0 focus:outline-none focus:ring-0"></div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import type { Dialog } from '@/shared/protocol/ora/dialog';
import UiGeneric from '@/components/UiGeneric.vue';
import { computed, nextTick, ref, watch } from 'vue';
import { ElementSize } from '@/shared/protocol/ora/elementSize';

const props = defineProps<{
	ui: Dialog;
}>();

const firstFocusableElement = ref<HTMLElement|undefined>();
const lastFocusableElement = ref<HTMLElement|undefined>();

watch(() => props.ui.visible.v, (newValue) => {
	if (newValue) {
		nextTick(() => {
			lastFocusableElement.value?.focus();
		});
	}
})

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

function moveFocusForward(e: KeyboardEvent): void {
	if (e.shiftKey) {
		return;
	}
	if (document.activeElement as HTMLElement === lastFocusableElement.value) {
		e.preventDefault();
		firstFocusableElement.value?.focus();
		return;
	}
}

function moveFocusBackwards(e: KeyboardEvent): void {
	if (document.activeElement as HTMLElement === firstFocusableElement.value) {
		e.preventDefault();
		lastFocusableElement.value?.focus();
		return;
	}
}
</script>

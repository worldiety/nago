<template>
	<Teleport to="#ora-modals">
		<div
			ref="dialogContainer"
			class="pointer-events-auto"
			@keydown.tab.exact="moveFocusForward"
			@keydown.shift.tab="moveFocusBackwards"
		>

			<UiGeneric v-if="ui.b" :ui="ui.b" class="h-screen w-screen" @click.stop/>

		</div>
	</Teleport>
</template>

<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {onMounted, ref, watch} from 'vue';
import {Modal} from "@/shared/protocol/ora/modal";

const props = defineProps<{
	ui: Modal;
	//isActiveDialog: boolean|undefined;
}>();

const dialogContainer = ref<HTMLElement | undefined>();
let firstFocusableElement: HTMLElement | undefined;
let lastFocusableElement: HTMLElement | undefined;

onMounted(() => {
	//if (props.isActiveDialog) {
		captureFocusInDialog();
	//}
});

watch(() => props.ui, (newValue) => {
	if (newValue) {
		captureFocusInDialog();
	}
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

<template>
	<div v-if="!ui.content"></div>
	<Teleport v-else-if="ui.modalType.value == ModalTypeValues.ModalTypeOverlay" to="#ora-overlay">
		<Transition>
			<div v-show="ui.content" class="pointer-events-auto fixed" :style="styles">
				<UiGeneric :ui="ui.content" class="" />
			</div>
		</Transition>
	</Teleport>

	<Teleport to="#ora-modals" v-else>
		<div
			ref="dialogContainer"
			class="pointer-events-auto fixed"
			@keydown.tab.exact="moveFocusForward"
			@keydown.shift.tab="moveFocusBackwards"
		>
			<UiGeneric v-if="ui.content" :ui="ui.content" class="h-screen w-screen" @click.stop />
		</div>
	</Teleport>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { cssLengthValue } from '@/components/shared/length';
import { Modal, ModalTypeValues } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Modal;
	//isActiveDialog: boolean|undefined;
}>();

const dialogContainer = ref<HTMLElement | undefined>();
let firstFocusableElement: HTMLElement | undefined;
let lastFocusableElement: HTMLElement | undefined;

onMounted(() => {
	//if (props.isActiveDialog) {
	if (props.ui.modalType.value !== ModalTypeValues.ModalTypeOverlay) {
		captureFocusInDialog();
	}

	//}
});

// TODO the following code causes focus-lost event in input elements and seems not be appropriate anymore - this is a port from a dialog
// watch(() => props.ui, (newValue) => {
// 	if (newValue) {
// 		captureFocusInDialog();
// 	}
// });

const styles = computed<string>(() => {
	const styles: string[] = [];
	if (!props.ui.right.isZero()) {
		styles.push(`right: ${cssLengthValue(props.ui.right.value)}`);
	}

	if (!props.ui.top.isZero()) {
		styles.push(`top: ${cssLengthValue(props.ui.top.value)}`);
	}

	if (!props.ui.left.isZero()) {
		styles.push(`left: ${cssLengthValue(props.ui.left.value)}`);
	}

	if (!props.ui.bottom.isZero()) {
		styles.push(`bottom: ${cssLengthValue(props.ui.bottom.value)}`);
	}

	return styles.join(';');
});

function captureFocusInDialog(): void {
	const focusableElements =
		dialogContainer.value?.querySelectorAll('[tabindex="0"], button:not([tabindex="-1"])') ?? [];
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
	if ((document.activeElement as HTMLElement) === lastFocusableElement) {
		e.preventDefault();
		firstFocusableElement?.focus();
		return;
	}
}

function moveFocusBackwards(e: KeyboardEvent): void {
	if ((document.activeElement as HTMLElement) === firstFocusableElement) {
		e.preventDefault();
		lastFocusableElement?.focus();
		return;
	}
}
</script>
<style>
.v-enter-active,
.v-leave-active {
	transition: opacity 0.5s ease;
}

.v-enter-from,
.v-leave-to {
	opacity: 0;
}
</style>

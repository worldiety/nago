<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template>
	<div v-if="!ui.content"></div>
	<Teleport v-else-if="ui.modalType == ModalTypeValues.ModalTypeOverlay" to="#ora-overlay">
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
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { cssLengthValue } from '@/components/shared/length';
import { onModalClose, onModalOpen } from '@/components/shared/modalManager';
import { Modal, ModalTypeValues } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Modal;
	//isActiveDialog: boolean|undefined;
}>();

const dialogContainer = ref<HTMLElement | undefined>();
let firstFocusableElement: HTMLElement | undefined;
let lastFocusableElement: HTMLElement | undefined;

onMounted(() => {
	if (!props.ui.allowBackgroundScrolling) {
		onModalOpen();
	}
	//if (props.isActiveDialog) {
	if (props.ui.modalType !== ModalTypeValues.ModalTypeOverlay) {
		captureFocusInDialog();
	}

	//}
});

onBeforeUnmount(() => {
	if (!props.ui.allowBackgroundScrolling) {
		onModalClose();
	}
});

// TODO the following code causes focus-lost event in input elements and seems not be appropriate anymore - this is a port from a dialog
// watch(() => props.ui, (newValue) => {
// 	if (newValue) {
// 		captureFocusInDialog();
// 	}
// });

const styles = computed<string>(() => {
	const styles: string[] = [];
	if (props.ui.right) {
		styles.push(`right: ${cssLengthValue(props.ui.right)}`);
	}

	if (props.ui.top) {
		styles.push(`top: ${cssLengthValue(props.ui.top)}`);
	}

	if (props.ui.left) {
		styles.push(`left: ${cssLengthValue(props.ui.left)}`);
	}

	if (props.ui.bottom) {
		styles.push(`bottom: ${cssLengthValue(props.ui.bottom)}`);
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

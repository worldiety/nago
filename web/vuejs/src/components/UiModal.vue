<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template>
	<div v-if="!ui.content"></div>
	<Teleport v-else-if="ui.modalType == ModalTypeValues.ModalTypeNotifications" to="#ora-notifications">
		<Transition>
			<div v-show="ui.content" class="pointer-events-none fixed max-w-full" :style="styles">
				<UiGeneric :ui="ui.content" class="" />
			</div>
		</Transition>
	</Teleport>

	<Teleport v-else-if="ui.modalType == ModalTypeValues.ModalTypeOverlay" to="#ora-overlay">
		<Transition>
			<div v-show="ui.content" class="pointer-events-auto fixed" :style="styles" @click.stop="sendDismissEvent">
				<UiGeneric :ui="ui.content" class="" />
			</div>
		</Transition>
	</Teleport>

	<Teleport v-else to="#ora-modals">
		<div
			ref="dialogContainer"
			class="pointer-events-auto fixed"
			@pointerdown="pressedOnBackdrop = $event.target === $event.currentTarget"
			@pointerup="releasedOnBackdrop = $event.target === $event.currentTarget"
			@click.stop="onBackdropClick"
		>
			<UiGeneric v-if="ui.content" :ui="ui.content" class="h-screen w-screen" @click.stop />
		</div>
	</Teleport>
</template>

<script lang="ts" setup>
import { computed, onBeforeUnmount, onMounted, onUnmounted, ref } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { cssLengthValue } from '@/components/shared/length';
import { onModalClose, onModalOpen } from '@/components/shared/modalManager';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import type { Modal } from '@/shared/proto/nprotoc_gen';
import { FunctionCallRequested } from '@/shared/proto/nprotoc_gen';
import { ModalTypeValues } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Modal;
	//isActiveDialog: boolean|undefined;
}>();

const serviceAdapter = useServiceAdapter();
const dialogContainer = ref<HTMLElement | undefined>();

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

function sendDismissEvent() {
	if (!props.ui.onDismissRequest) return;

	serviceAdapter.sendEvent(new FunctionCallRequested(props.ui.onDismissRequest, nextRID()));
}

// Chrome and Safari send a click to the common ancestor, if the button is pressed inside the dialog, e.g. to
// select text, and released outside of it, or the other way round. That ancestor is the backdrop, so only a click
// which starts and ends on the backdrop dismisses.
let pressedOnBackdrop = false;
let releasedOnBackdrop = false;

function onBackdropClick(e: MouseEvent) {
	const onBackdrop = pressedOnBackdrop && releasedOnBackdrop && e.target === e.currentTarget;
	pressedOnBackdrop = false;
	releasedOnBackdrop = false;
	if (onBackdrop) sendDismissEvent();
}

const focusableSelector = [
	'a[href]',
	'area[href]',
	'button',
	'input:not([type="hidden"])',
	'select',
	'textarea',
	'iframe',
	'[contenteditable]:not([contenteditable="false"])',
	'[tabindex]',
].join(',');

// returns all tabbable elements of the dialog in document order. It is evaluated on each use, because the
// dialog content may change while it is open.
function getFocusableElements(): HTMLElement[] {
	if (!dialogContainer.value) return [];

	const dialog = dialogContainer.value.querySelector('dialog');
	if (!dialog) return [];

	// find the highest dialog section and assume it contains the dialogs main content
	let highestDialogSection: HTMLElement | undefined;
	for (let child of dialog.children) {
		if (!highestDialogSection || child.clientHeight > highestDialogSection.clientHeight) {
			highestDialogSection = child as HTMLElement;
		}
	}

	if (!highestDialogSection) return [];

	const candidates = highestDialogSection.querySelectorAll<HTMLElement>(focusableSelector) ?? [];
	return Array.from(candidates).filter(
		(el) =>
			el.tabIndex >= 0 &&
			!(el as HTMLButtonElement).disabled &&
			!el.closest('[inert]') &&
			el.getClientRects().length > 0
	);
}

function captureFocusInDialog(): void {
	getFocusableElements()[0]?.focus({ focusVisible: true } as any); // focus options help to make the focus event visible
}

function onKeydown(e: KeyboardEvent) {
	if (e.code === 'Escape') sendDismissEvent();
}

onMounted(() => {
	addEventListener('keydown', onKeydown);

	if (!props.ui.allowBackgroundScrolling) {
		onModalOpen();
	}

	if (props.ui.modalType !== ModalTypeValues.ModalTypeOverlay) {
		setTimeout(captureFocusInDialog, 50); // wait for dialog content to be rendered
	}
});

onBeforeUnmount(() => {
	if (!props.ui.allowBackgroundScrolling) {
		onModalClose();
	}
});

onUnmounted(() => {
	removeEventListener('keydown', onKeydown);
});
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

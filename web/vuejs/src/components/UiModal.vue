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
import { computed, onBeforeUnmount, onMounted, onUnmounted, ref, watch } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { cssLengthValue } from '@/components/shared/length';
import { ModalManager } from '@/components/shared/modalManager';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import type { Modal } from '@/shared/proto/nprotoc_gen';
import { FunctionCallRequested } from '@/shared/proto/nprotoc_gen';
import { ModalTypeValues } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Modal;
}>();

const serviceAdapter = useServiceAdapter();
const dialogContainer = ref<HTMLElement | undefined>();

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

// returns all tabbable elements of the dialogs main content in document order. It is evaluated on each use, because
// the dialog content may change while it is open.
function getInitialFocusCandidates(): HTMLElement[] {
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

	return ModalManager.getTabbableElements(highestDialogSection);
}

function captureFocusInDialog(): void {
	getInitialFocusCandidates()[0]?.focus({ focusVisible: true } as any); // focus options help to make the focus event visible
}

function onKeydown(e: KeyboardEvent) {
	if (e.code === 'Escape') sendDismissEvent();
}

// the dialog container only exists, if this is a dialog with content. While it exists, the focus is trapped in it.
watch(dialogContainer, (newContainer, oldContainer) => {
	if (oldContainer) ModalManager.popDialog(oldContainer);

	if (newContainer) {
		ModalManager.pushDialog(newContainer);
		setTimeout(captureFocusInDialog, 50); // wait for dialog content to be rendered
	}
});

let scrollLocked = false;

function setScrollLock(lock: boolean) {
	if (lock === scrollLocked) return;

	scrollLocked = lock;
	if (lock) {
		ModalManager.onModalOpen();
	} else {
		ModalManager.onModalClose();
	}
}

watch(
	() => !!props.ui.content && !props.ui.allowBackgroundScrolling,
	(lock) => setScrollLock(lock)
);

onMounted(() => {
	addEventListener('keydown', onKeydown);
	setScrollLock(!!props.ui.content && !props.ui.allowBackgroundScrolling);
});

onBeforeUnmount(() => {
	setScrollLock(false);
	if (dialogContainer.value) ModalManager.popDialog(dialogContainer.value);
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

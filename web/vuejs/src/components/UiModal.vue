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

// returns the dialog element. It is evaluated on each use, because the dialog content may change while it is open.
function getDialog(): HTMLElement | undefined {
	return dialogContainer.value?.querySelector('dialog') ?? undefined;
}

// Moves the focus to the start of the dialog (usually its title) without focusing an interactive element. The dialog
// is made programmatically focusable (but not tabbable), thus the next tab key press or the next screen reader
// navigation step starts at the top of the dialog.
// The dialog is only focusable temporarily: once the focus leaves it, the tabindex is removed again. Otherwise,
// VoiceOver treats the dialog as a focusable group and gets stuck at its boundary, when navigating backwards into
// it from outside.
function captureFocusInDialog(): void {
	const dialog = getDialog();
	if (!dialog) return;

	const addedTabIndex = !dialog.hasAttribute('tabindex');
	if (addedTabIndex) dialog.setAttribute('tabindex', '-1');
	dialog.setAttribute('data-modal-focus-start', '');

	const release = () => {
		dialog.removeEventListener('blur', release);
		dialog.removeAttribute('data-modal-focus-start');
		if (addedTabIndex) dialog.removeAttribute('tabindex');
	};
	dialog.addEventListener('blur', release);

	dialog.focus({ preventScroll: true });
	if (document.activeElement !== dialog) release();
}

function onKeydown(e: KeyboardEvent) {
	if (e.code === 'Escape') sendDismissEvent();
}

// the dialog container only exists, if this is a dialog with content. While it exists, the focus is trapped in it.
watch(dialogContainer, (newContainer, oldContainer) => {
	if (oldContainer) ModalManager.popDialog(oldContainer);

	if (newContainer) {
		ModalManager.pushDialog(newContainer);
		setTimeout(captureFocusInDialog, 100); // wait for dialog content to be rendered
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

/* the dialog is only focused as starting point for the tab order, it is not interactive */
[data-modal-focus-start]:focus {
	outline: none;
}
</style>

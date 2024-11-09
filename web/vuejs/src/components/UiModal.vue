<template>

	<div v-if="!ui.b"></div>
	<Teleport v-else-if="ui.t==1" to="#ora-overlay">
		<Transition>
			<div v-show="ui.b" class="pointer-events-auto fixed " :style="styles">

				<UiGeneric :ui="ui.b" class=""/>

			</div>
		</Transition>
	</Teleport>

	<Teleport to="#ora-modals" v-else>
		<div
			ref="dialogContainer"
			class="pointer-events-auto fixed "
			@keydown.tab.exact="moveFocusForward"
			@keydown.shift.tab="moveFocusBackwards"
		>

			<UiGeneric v-if="ui.b" :ui="ui.b" class="h-screen w-screen" @click.stop/>

		</div>
	</Teleport>

</template>

<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import {computed, onMounted, ref} from 'vue';
import {Modal} from "@/shared/protocol/ora/modal";
import {cssLengthValue} from "@/components/shared/length";


const props = defineProps<{
	ui: Modal;
	//isActiveDialog: boolean|undefined;
}>();

const dialogContainer = ref<HTMLElement | undefined>();
let firstFocusableElement: HTMLElement | undefined;
let lastFocusableElement: HTMLElement | undefined;

onMounted(() => {
	//if (props.isActiveDialog) {
	if (props.ui.t !== 1) {
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
	if (props.ui.r) {
		styles.push(`right: ${cssLengthValue(props.ui.r)}`)
	}

	if (props.ui.u) {
		styles.push(`top: ${cssLengthValue(props.ui.u)}`)
	}

	if (props.ui.l) {
		styles.push(`left: ${cssLengthValue(props.ui.l)}`)
	}

	if (props.ui.bt) {
		styles.push(`bottom: ${cssLengthValue(props.ui.bt)}`)
	}

	return styles.join(";")
})

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

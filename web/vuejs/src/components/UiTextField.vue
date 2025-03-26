<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<script lang="ts" setup>
import { computed, ref, watch } from 'vue';
import CloseIcon from '@/assets/svg/close.svg';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import { frameCSS } from '@/components/shared/frame';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import {
	FunctionCallRequested,
	KeyboardTypeValues,
	TextField,
	TextFieldStyleValues,
	UpdateStateValueRequested,
} from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: TextField;
}>();

const serviceAdapter = useServiceAdapter();
const inputValue = ref<string>(props.ui.value ? props.ui.value : '');

//console.log("uitextfield", props.ui.inputValue.value, "=" + props.ui.value.value)

/**
 * Validates the input value and submits it, if it is valid.
 * The '-' sign and the empty string are treated as 0.
 * If the input value is invalid, the value gets reset to the last known valid value.
 */
watch(inputValue, (newValue, oldValue) => {
	if (newValue == oldValue) {
		return;
	}

	if (props.ui.keyboardOptions?.keyboardType == KeyboardTypeValues.KeyboardInteger) {
		if (newValue === '' || newValue === '-') {
			inputValue.value = '0';
		} else if (!newValue.match(/^-?[0-9]+$/)) {
			inputValue.value = oldValue;
		}

		return;
	}

	if (props.ui.keyboardOptions?.keyboardType == KeyboardTypeValues.KeyboardFloat) {
		if (newValue === '' || newValue === '-') {
			inputValue.value = '0';
		} else if (!newValue.match(/^[+-]?(\d+(\.\d*)?|\.\d+)$/)) {
			inputValue.value = oldValue;
		}

		return;
	}
});

watch(
	() => props.ui.value,
	(newValue) => {
		//console.log("textfield triggered props.ui.value",inputValue.value,newValue)
		if (newValue) {
			inputValue.value = newValue;
		} else {
			inputValue.value = '';
		}
	}
);

watch(
	() => props.ui,
	(newValue) => {
		//console.log("textfield triggered props.ui","p="+props.ui.p,"old="+inputValue.value," new="+newValue.v,newValue)
		if (newValue.value) {
			inputValue.value = newValue.value;
		} else {
			inputValue.value = '';
		}
	}
);

function handleKeydownEnter(event: Event) {
	if (props.ui.keydownEnter) {
		event.stopPropagation();
		serviceAdapter.sendEvent(new FunctionCallRequested(props.ui.keydownEnter, nextRID()));
	}
}

function submitInputValue(force: boolean): void {
	if (inputValue.value == props.ui.value) {
		return;
	}

	// Note, that the sendEvent may have a huge latency, causing ghost updates for the user input.
	// Thus, immediately increase the request id, so that everybody knows, that any older responses are outdated.
	nextRID();

	if (force || props.ui.disableDebounce) {
		serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), inputValue.value));

		return;
	}

	debouncedInput();
}

function clearInputValue(): void {
	inputValue.value = '';
	submitInputValue(true);
}

function deserializeGoDuration(durationInNanoseconds: number): number {
	return durationInNanoseconds / 1e6;
}

let timer: number = 0;

function debouncedInput() {
	let debounceTime = 500; // ms
	if (props.ui.debounceTime && props.ui.debounceTime > 0) {
		debounceTime = deserializeGoDuration(props.ui.debounceTime);
	}

	clearTimeout(timer);
	timer = window.setTimeout(() => {
		if (inputValue.value == props.ui.value) {
			return;
		}

		serviceAdapter.sendEvent(new UpdateStateValueRequested(props.ui.inputValue, 0, nextRID(), inputValue.value));
	}, debounceTime);
}

const frameStyles = computed<string>(() => {
	return frameCSS(props.ui.frame).join(';');
});

const id = computed<string>(() => {
	if (props.ui.id) {
		return props.ui.id;
	}

	return 'tf-' + props.ui.inputValue;
});

const inputMode = computed<string>(() => {
	switch (props.ui.keyboardOptions?.keyboardType) {
		case KeyboardTypeValues.KeyboardInteger:
			return 'numeric';
		case KeyboardTypeValues.KeyboardFloat:
			return 'decimal';
		case KeyboardTypeValues.KeyboardEMail:
			return 'email';
		case KeyboardTypeValues.KeyboardPhone:
			return 'tel';
		case KeyboardTypeValues.KeyboardURL:
			return 'url';
		case KeyboardTypeValues.KeyboardSearch:
			return 'search';
	}

	return 'text';
});

// TODO check :id="idPrefix + props.ui.id.toString()"

// TODO this is not properly modelled: the padding trick below does not work with arbitrary content (prefix, postfix). Use focus-within and a border around flex flex-row, so that we don't need that padding stuff
// TODO implement TextFieldBasic (b) render mode
</script>

<template>
	<div v-if="!ui.invisible" :style="frameStyles">
		<InputWrapper
			:simple="props.ui.style == TextFieldStyleValues.TextFieldReduced"
			:label="props.ui.label"
			:error="props.ui.errorText"
			:help="props.ui.supportingText"
			:disabled="props.ui.disabled"
		>
			<div class="relative">
				<input
					v-if="!props.ui.lines"
					@keydown.enter="handleKeydownEnter"
					:id="id"
					v-model="inputValue"
					class="input-field !pr-10"
					:disabled="props.ui.disabled"
					type="text"
					:inputmode="inputMode"
					@focusout="submitInputValue(true)"
					@input="submitInputValue(false)"
				/>
				<textarea
					v-if="props.ui.lines"
					:id="id"
					v-model="inputValue"
					class="input-field !pr-10"
					:disabled="props.ui.disabled"
					type="text"
					:rows="props.ui.lines"
					@focusout="submitInputValue(true)"
					@input="submitInputValue(false)"
				/>

				<div
					v-if="inputValue && !props.ui.disabled"
					class="absolute top-0 bottom-0 right-4 flex items-center h-full"
				>
					<CloseIcon class="w-4" tabindex="-1" @click="clearInputValue" @keydown.enter="clearInputValue" />
				</div>
			</div>
		</InputWrapper>
	</div>
</template>

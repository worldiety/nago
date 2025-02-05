<script lang="ts" setup>
import {computed, ref, watch} from 'vue';
import CloseIcon from '@/assets/svg/close.svg';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import {frameCSS} from '@/components/shared/frame';
import {useServiceAdapter} from '@/composables/serviceAdapter';
import {
	KeyboardTypeValues,
	Ptr,
	Str,
	TextField,
	TextFieldStyleValues,
	UpdateStateValueRequested
} from "@/shared/proto/nprotoc_gen";
import {nextRID} from "@/eventhandling";

const props = defineProps<{
	ui: TextField;
}>();

const serviceAdapter = useServiceAdapter();
const inputValue = ref<string>(props.ui.value.value);

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

	if (props.ui.keyboardOptions.keyboardType.value == KeyboardTypeValues.KeyboardInteger) {
		if (newValue === '' || newValue === '-') {
			inputValue.value = '0';
		} else if (!newValue.match(/^-?[0-9]+$/)) {
			inputValue.value = oldValue;
		}

		return;
	}

	if (props.ui.keyboardOptions.keyboardType.value == KeyboardTypeValues.KeyboardFloat) {
		if (newValue === '' || newValue === '-') {
			inputValue.value = '0';
		} else if (!newValue.match(/^[+-]?(\d+(\.\d*)?|\.\d+)$/)) {
			inputValue.value = oldValue;
		}

		return;
	}
});

watch(
	() => props.ui.value.value,
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
		if (newValue.value.value) {
			inputValue.value = newValue.value.value;
		} else {
			inputValue.value = '';
		}
	}
);

function submitInputValue(force: boolean): void {
	if (inputValue.value == props.ui.value.value) {
		return;
	}

	if (force || (props.ui.disableDebounce.value && !props.ui.inputValue.isZero())) {
		serviceAdapter.sendEvent(new UpdateStateValueRequested(
			props.ui.inputValue,
			new Ptr(),
			nextRID(),
			new Str(inputValue.value),
		))

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
	if (props.ui.debounceTime.value > 0) {
		debounceTime = deserializeGoDuration(props.ui.debounceTime.value);
	}

	clearTimeout(timer);
	timer = window.setTimeout(() => {
		if (inputValue.value == props.ui.value.value) {
			return;
		}

		serviceAdapter.sendEvent(new UpdateStateValueRequested(
			props.ui.inputValue,
			new Ptr(),
			nextRID(),
			new Str(inputValue.value),
		))
	}, debounceTime);
}

const frameStyles = computed<string>(() => {
	return frameCSS(props.ui.frame).join(';');
});

const id = computed<string>(() => {
	return 'tf-' + props.ui.inputValue.value;
});

const inputMode = computed<string>(() => {
	switch (props.ui.keyboardOptions.keyboardType.value) {
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
	<div v-if="!ui.invisible.value" :style="frameStyles">
		<InputWrapper
			:simple="props.ui.style.value==TextFieldStyleValues.TextFieldReduced"
			:label="props.ui.label.value"
			:error="props.ui.errorText.value"
			:help="props.ui.supportingText.value"
			:disabled="props.ui.disabled.value"
		>
			<div class="relative">
				<input
					v-if="!props.ui.lines.value"
					:id="id"
					v-model="inputValue"
					class="input-field !pr-10"
					:disabled="props.ui.disabled.value"
					type="text"
					:inputmode="inputMode"
					@focusout="submitInputValue(true)"
					@input="submitInputValue(false)"
				/>
				<textarea
					v-if="props.ui.lines.value"
					:id="id"
					v-model="inputValue"
					class="input-field !pr-10"
					:disabled="props.ui.disabled.value"
					type="text"
					:rows="props.ui.lines.value"
					@focusout="submitInputValue(true)"
					@input="submitInputValue(false)"
				/>

				<div v-if="inputValue && !props.ui.disabled.value"
						 class="absolute top-0 bottom-0 right-4 flex items-center h-full">
					<CloseIcon class="w-4" tabindex="-1" @click="clearInputValue" @keydown.enter="clearInputValue"/>
				</div>
			</div>
		</InputWrapper>
	</div>
</template>

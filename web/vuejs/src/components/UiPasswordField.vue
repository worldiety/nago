<script lang="ts" setup>
import {computed, ref, watch} from 'vue';
import HideIcon from '@/assets/svg/hide.svg';
import RevealIcon from '@/assets/svg/reveal.svg';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import {frameCSS} from '@/components/shared/frame';
import {useServiceAdapter} from '@/composables/serviceAdapter';
import {
	PasswordField,
	Ptr,
	Str,
	StylePresetValues,
	TextFieldStyle, TextFieldStyleValues,
	UpdateStateValueRequested
} from "@/shared/proto/nprotoc_gen";
import {nextRID} from "@/eventhandling";

const props = defineProps<{
	ui: PasswordField;
}>();

const serviceAdapter = useServiceAdapter();
const passwordInput = ref<HTMLElement | undefined>();
const inputValue = ref<string>(props.ui.value.value);
const idPrefix = 'password-field-';

watch(
	() => props.ui.value.value,
	(newValue) => {
		if (newValue) {
			inputValue.value = newValue;
		} else {
			inputValue.value = '';
		}
		//console.log("textfield triggered props.ui.value")
	}
);

watch(
	() => props.ui,
	(newValue) => {
		//console.log("textfield triggered props.ui")
		inputValue.value = newValue.value.value;
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
		if ( inputValue.value == props.ui.value.value) {
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

function toggleRevealed(): void {
	props.ui.revealed.value = !props.ui.revealed.value;
	passwordInput.value?.focus();
}
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
			<div class="relative hover:text-primary focus-within:text-primary">
				<input
					:id="idPrefix"
					:autocomplete="props.ui.disableAutocomplete.value ? 'off' : 'on'"
					ref="passwordInput"
					v-model="inputValue"
					class="input-field !pr-10"
					:disabled="props.ui.disabled.value"
					:type="props.ui.revealed.value ? 'text' : 'password'"
					@focusout="submitInputValue(true)"
					@input="submitInputValue(false)"
				/>
				<div class="absolute top-0 bottom-0 right-4 flex items-center text-black h-full">
					<div :tabindex="props.ui.disabled.value ? '-1' : '0'" @click="toggleRevealed" @keydown.enter="toggleRevealed">
						<RevealIcon v-if="!props.ui.disabled.value" class="w-6"/>
						<HideIcon v-else class="w-6"/>
					</div>
				</div>
			</div>
		</InputWrapper>
	</div>
</template>

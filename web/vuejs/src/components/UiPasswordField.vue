<script lang="ts" setup>
import { computed, ref, watch } from 'vue';
import HideIcon from '@/assets/svg/hide.svg';
import RevealIcon from '@/assets/svg/reveal.svg';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import { frameCSS } from '@/components/shared/frame';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import {
	FunctionCallRequested,
	PasswordField,
	TextFieldStyleValues,
	UpdateStateValueRequested,
} from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: PasswordField;
}>();

const serviceAdapter = useServiceAdapter();
const passwordInput = ref<HTMLElement | undefined>();
const inputValue = ref<string>(props.ui.value ? props.ui.value : '');

watch(
	() => props.ui.value,
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
		inputValue.value = newValue.value ? newValue.value : '';
	}
);

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

function handleKeydownEnter(event: Event) {
	if (props.ui.keydownEnter) {
		event.stopPropagation();
		serviceAdapter.sendEvent(new FunctionCallRequested(props.ui.keydownEnter, nextRID()));
	}
}

function toggleRevealed(): void {
	props.ui.revealed = !props.ui.revealed;
	passwordInput.value?.focus();
}
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
			<div class="relative hover:text-primary focus-within:text-primary">
				<input
					:id="ui.id"
					@keydown.enter="handleKeydownEnter"
					:autocomplete="props.ui.disableAutocomplete ? 'off' : 'on'"
					ref="passwordInput"
					v-model="inputValue"
					class="input-field !pr-10"
					:disabled="props.ui.disabled"
					:type="props.ui.revealed ? 'text' : 'password'"
					@focusout="submitInputValue(true)"
					@input="submitInputValue(false)"
				/>
				<div class="absolute top-0 bottom-0 right-4 flex items-center text-black h-full">
					<div
						:tabindex="props.ui.disabled ? '-1' : '0'"
						@click="toggleRevealed"
						@keydown.enter="toggleRevealed"
					>
						<RevealIcon v-if="!props.ui.disabled" class="w-6" />
						<HideIcon v-else class="w-6" />
					</div>
				</div>
			</div>
		</InputWrapper>
	</div>
</template>

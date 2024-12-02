<script lang="ts" setup>
import {computed, ref, watch} from 'vue';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import RevealIcon from '@/assets/svg/reveal.svg';
import HideIcon from '@/assets/svg/hide.svg';
import type {PasswordField} from '@/shared/protocol/ora/passwordField';
import {useServiceAdapter} from '@/composables/serviceAdapter';
import {frameCSS} from "@/components/shared/frame";

const props = defineProps<{
	ui: PasswordField;
}>();

const serviceAdapter = useServiceAdapter();
const passwordInput = ref<HTMLElement | undefined>();
const inputValue = ref<string>(props.ui.v ? props.ui.v : "");
const idPrefix = 'password-field-';


watch(() => props.ui.v, (newValue) => {
	if (newValue) {
		inputValue.value = newValue
	}else{
		inputValue.value=""
	}
	//console.log("textfield triggered props.ui.value")
});

watch(() => props.ui, (newValue) => {
	//console.log("textfield triggered props.ui")
	if (newValue.v) {
		inputValue.value = newValue.v;
	}else{
		inputValue.value=""
	}
});

function submitInputValue(force: boolean): void {
	if (props.ui.v && inputValue.value == props.ui.v) {
		return
	}

	if (force || props.ui.i && props.ui.p) {
		serviceAdapter.setProperties({
			p: props.ui.p,
			v: inputValue.value,
		});
		return
	}


	debouncedInput()
}

function deserializeGoDuration(durationInNanoseconds: number): number {
	return durationInNanoseconds / 1e6;
}

let timer: number = 0;

function debouncedInput() {
	let debounceTime = 500 // ms
	if (props.ui.dt && props.ui.dt > 0) {
		debounceTime = deserializeGoDuration(props.ui.dt)
	}

	clearTimeout(timer)
	timer = window.setTimeout(() => {
		if (props.ui.v && inputValue.value == props.ui.v) {
			return
		}

		serviceAdapter.setProperties({
			p: props.ui.p,
			v: inputValue.value,
		});
	}, debounceTime)
}

const frameStyles = computed<string>(() => {
	return frameCSS(props.ui.f).join(";")
});

function toggleRevealed(): void {
	props.ui.rv = !props.ui.rv
	passwordInput.value?.focus();
}
</script>

<template>
	<div v-if="!ui.iv" :style="frameStyles">
		<InputWrapper
			:simple="props.ui.t && props.ui.t=='r'"
			:label="props.ui.l"
			:error="props.ui.e"
			:hint="props.ui.s"
			:disabled="props.ui.d"
		>
			<div class="relative hover:text-primary focus-within:text-primary">
				<input
					:id="idPrefix"
					ref="passwordInput"
					v-model="inputValue"
					class="input-field !pr-10"
					:disabled="props.ui.d"
					:type="props.ui.rv ? 'text' : 'password'"
					@focusout="submitInputValue(true)"
					@input="submitInputValue(false)"
				/>
				<div class="absolute top-0 bottom-0 right-4 flex items-center text-black h-full">
					<div :tabindex="props.ui.d ? '-1' : '0'" @click="toggleRevealed" @keydown.enter="toggleRevealed">
						<RevealIcon v-if="!props.ui.d" class="w-6"/>
						<HideIcon v-else class="w-6"/>
					</div>
				</div>
			</div>
		</InputWrapper>
	</div>
</template>

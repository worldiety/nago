<script lang="ts" setup>
import { ref, watch } from 'vue';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import CloseIcon from '@/assets/svg/close.svg';
import type {TextField} from "@/shared/protocol/ora/textField";
import { useServiceAdapter } from '@/composables/serviceAdapter';

const props = defineProps<{
	ui: TextField;
}>();

const serviceAdapter = useServiceAdapter();
const inputValue = ref<string>(props.ui.value.v);
const idPrefix = 'text-field-';
let timeout: number|null = null;

watch(() => props.ui.value.v, (newValue) => {
	inputValue.value = newValue;
});

function inputValueChanged(): void {
	// Use debouncing here to prevent excessive updates
	if (timeout !== null) {
		return;
	}
	timeout = window.setTimeout(() => {
		submitInputValue();
		timeout = null;
	}, 500);
}

function submitInputValue(): void {
	serviceAdapter.setPropertiesAndCallFunctions([{
		...props.ui.value,
		v: inputValue.value,
	}], [props.ui.onTextChanged]);
}

function clearInputValue(): void {
	inputValue.value = '';
	submitInputValue();
}
</script>

<template>
	<div>
		<InputWrapper
			:simple="props.ui.simple.v"
			:label="props.ui.label.v"
			:error="props.ui.error.v"
			:hint="props.ui.hint.v"
			:help="props.ui.help.v"
			:disabled="props.ui.disabled.v"
		>
			<div class="relative">
				<input
					:id="idPrefix + props.ui.id.toString()"
					v-model="inputValue"
					class="input-field"
					:class="{'!pr-10': inputValue}"
					:placeholder="props.ui.placeholder.v"
					:disabled="props.ui.disabled.v"
					type="text"
					@input="inputValueChanged"
				/>
				<div v-if="inputValue" class="absolute top-0 bottom-0 right-4 flex items-center h-full">
					<CloseIcon class="w-4" tabindex="0" @click="clearInputValue" @keydown.enter="clearInputValue" />
				</div>
			</div>
		</InputWrapper>
	</div>
</template>

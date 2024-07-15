<script lang="ts" setup>
import {computed, ref, watch} from 'vue';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import CloseIcon from '@/assets/svg/close.svg';
import type {TextField} from "@/shared/protocol/ora/textField";
import {useServiceAdapter} from '@/composables/serviceAdapter';
import {isNil} from "@/shared/protocol/util";
import {frameCSS} from "@/components/shared/frame";

const props = defineProps<{
	ui: TextField;
}>();

const serviceAdapter = useServiceAdapter();
const inputValue = ref<string>(props.ui.value.v);
const idPrefix = 'text-field-';

watch(() => props.ui.value, (newValue) => {
	console.log("textfield triggered props.ui.value")
});

watch(() => props.ui, (newValue) => {
	console.log("textfield triggered props.ui")
	inputValue.value = newValue.value.v;
});

watch(() => props.ui.value.v, (newValue) => {
	console.log("textfield triggered props.ui.v")
	// TODO this is sometimes broken!!!! or is this a logical race, where v inputvalue is never updated and thus keeps empty?
	// see also https://vuejs.org/guide/essentials/watchers#deep-watchers
	inputValue.value = newValue;
});

function submitInputValue(): void {
	debouncedInput()

	if (isNil(props.ui.onTextChanged) && !isNil(props.ui.onDebouncedTextChanged)) {
		// this is a special case, to optimize re-render behavior for things like quick search:
		// we delay the property update until the debounce callback triggers, so we cannot cause any dirty roundtrips
		// TODO this has the known side effect, that you cannot read out the changed property AND have a debounced text changed event.
		return
	}

	serviceAdapter.setPropertiesAndCallFunctions([{
		...props.ui.value,
		v: inputValue.value,
	}], [props.ui.onTextChanged]);
}

function clearInputValue(): void {
	inputValue.value = '';
	submitInputValue();
}


function deserializeGoDuration(durationInNanoseconds: number): number {
	return durationInNanoseconds / 1e6;
}

let timer: number = 0;

function debouncedInput() {
	if (isNil(props.ui.onDebouncedTextChanged)) {
		return;
	}

	// TODO this will always cause an { type: "ErrorOccurred", r: 15, message: "cannot call function: no such pointer found: 7325" } but I don't see why

	clearTimeout(timer)
	timer = window.setTimeout(() => {
		serviceAdapter.setPropertiesAndCallFunctions([{
			...props.ui.value,
			v: inputValue.value,
		}], [props.ui.onDebouncedTextChanged]);
	}, deserializeGoDuration(props.ui.debounceTime))
}

const frameStyles = computed<string>(() => {
	return frameCSS(props.ui.frame).join(";")
});

// TODO check :id="idPrefix + props.ui.id.toString()"
</script>

<template>
	<div v-if="!ui.visible" :style="frameStyles">
		<InputWrapper
			:simple="props.ui.simple"
			:label="props.ui.label"
			:error="props.ui.error"
			:hint="props.ui.hint"
			:help="props.ui.help"
			:disabled="props.ui.disabled"
		>
			<div class="relative">
				<input
					:id="idPrefix"
					v-model="inputValue"
					class="input-field"
					:class="{'!pr-10': inputValue}"
					:placeholder="props.ui.placeholder"
					:disabled="props.ui.disabled"
					type="text"
					@input="submitInputValue"
				/>
				<div v-if="inputValue" class="absolute top-0 bottom-0 right-4 flex items-center h-full">
					<CloseIcon class="w-4" tabindex="0" @click="clearInputValue" @keydown.enter="clearInputValue"/>
				</div>
			</div>
		</InputWrapper>
	</div>
</template>

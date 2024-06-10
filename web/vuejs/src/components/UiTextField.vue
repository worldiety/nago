<script lang="ts" setup>
import {ref, watch} from 'vue';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import CloseIcon from '@/assets/svg/close.svg';
import type {TextField} from "@/shared/protocol/ora/textField";
import {useServiceAdapter} from '@/composables/serviceAdapter';
import {isNil} from "@/shared/protocol/util";

const props = defineProps<{
	ui: TextField;
}>();

const serviceAdapter = useServiceAdapter();
const inputValue = ref<string>(props.ui.value.v);
const idPrefix = 'text-field-';

watch(() => props.ui.value.v, (newValue) => {
	inputValue.value = newValue;
});

function submitInputValue(): void {
	debouncedInput()

	if (isNil(props.ui.onTextChanged.p) && !isNil(props.ui.onDebouncedTextChanged.p)) {
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
	if (isNil(props.ui.onDebouncedTextChanged.p)) {
		return;
	}

	// TODO this will always cause an { type: "ErrorOccurred", r: 15, message: "cannot call function: no such pointer found: 7325" } but I don't see why

	clearTimeout(timer)
	timer = window.setTimeout(() => {
		serviceAdapter.setPropertiesAndCallFunctions([{
			...props.ui.value,
			v: inputValue.value,
		}], [props.ui.onDebouncedTextChanged]);
	}, deserializeGoDuration(props.ui.debounceTime.v))
}

</script>

<template>
	<div v-if="ui.visible.v">
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
					@input="submitInputValue"
				/>
				<div v-if="inputValue" class="absolute top-0 bottom-0 right-4 flex items-center h-full">
					<CloseIcon class="w-4" tabindex="0" @click="clearInputValue" @keydown.enter="clearInputValue"/>
				</div>
			</div>
		</InputWrapper>
	</div>
</template>

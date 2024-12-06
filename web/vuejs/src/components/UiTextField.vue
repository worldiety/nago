<script lang="ts" setup>
import {computed, ref, watch} from 'vue';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import CloseIcon from '@/assets/svg/close.svg';
import type {TextField} from "@/shared/protocol/ora/textField";
import {useServiceAdapter} from '@/composables/serviceAdapter';
import {frameCSS} from "@/components/shared/frame";

const props = defineProps<{
	ui: TextField;
}>();

const serviceAdapter = useServiceAdapter();
const inputValue = ref<string>(props.ui.v ? props.ui.v : "");

//console.log("uitextfield", props.ui.p, "=" + props.ui.v)

/**
 * Validates the input value and submits it, if it is valid.
 * The '-' sign and the empty string are treated as 0.
 * If the input value is invalid, the value gets reset to the last known valid value.
 */
watch(inputValue, (newValue, oldValue) => {
	if (newValue == oldValue) {
		return
	}

	if (props.ui.o?.k == "i") {
		if (newValue === '' || newValue === '-') {
			inputValue.value = '0';
		} else if (!newValue.match(/^-?[0-9]+$/)) {
			inputValue.value = oldValue;
		}

		return
	}

	if (props.ui.o?.k == "f") {
		if (newValue === '' || newValue === '-') {
			inputValue.value = '0';
		} else if (!newValue.match(/^[+-]?(\d+(\.\d*)?|\.\d+)$/)) {
			inputValue.value = oldValue;
		}

		return
	}

});

watch(() => props.ui.v, (newValue) => {
	//console.log("textfield triggered props.ui.value",inputValue.value,newValue)
	if (newValue) {
		inputValue.value = newValue
	} else {
		inputValue.value = ""
	}

});

watch(() => props.ui, (newValue) => {
	//console.log("textfield triggered props.ui","p="+props.ui.p,"old="+inputValue.value," new="+newValue.v,newValue)
	if (newValue.v) {
		inputValue.value = newValue.v;
	} else {
		inputValue.value = ""
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

function clearInputValue(): void {
	inputValue.value = '';
	submitInputValue(true);
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

const id = computed<string>(() => {
	return "tf-" + props.ui.p
});

const inputMode = computed<string>(() => {
	switch (props.ui.o?.k) {
		case "i":
			return "numeric"
		case "f":
			return "decimal"
		case "m":
			return "email"
		case "p":
			return "tel"
		case "u":
			return "url"
		case "s":
			return "search"
	}

	return "text"
})

// TODO check :id="idPrefix + props.ui.id.toString()"

// TODO this is not properly modelled: the padding trick below does not work with arbitrary content (prefix, postfix). Use focus-within and a border around flex flex-row, so that we don't need that padding stuff
// TODO implement TextFieldBasic (b) render mode
</script>

<template>
	<div v-if="!ui.iv" :style="frameStyles">
		<InputWrapper
			:simple="props.ui.t && props.ui.t=='r'"
			:label="props.ui.l"
			:error="props.ui.e"
			:help="props.ui.s"
			:disabled="props.ui.d"
		>
			<div class="relative">
				<input
					v-if="!props.ui.li"
					:id="id"
					v-model="inputValue"
					class="input-field !pr-10"
					:disabled="props.ui.d"
					type="text"
					:inputmode="inputMode"
					@focusout="submitInputValue(true)"
					@input="submitInputValue(false)"
				/>
				<textarea
					v-if="props.ui.li"
					:id="id"
					v-model="inputValue"
					class="input-field !pr-10"
					:disabled="props.ui.d"
					type="text"
					:rows="props.ui.li"
					@focusout="submitInputValue(true)"
					@input="submitInputValue(false)"
				/>

				<div v-if="inputValue && !props.ui.d" class="absolute top-0 bottom-0 right-4 flex items-center h-full">
					<CloseIcon class="w-4" tabindex="-1" @click="clearInputValue" @keydown.enter="clearInputValue"/>
				</div>

			</div>
		</InputWrapper>
	</div>
</template>

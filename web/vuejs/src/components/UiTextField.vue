<script lang="ts" setup>
import { ref, watch } from 'vue';
import { useNetworkStore } from '@/stores/networkStore';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import CloseIcon from '@/assets/svg/close.svg';
import {TextField} from "@/shared/protocol/gen/textField";

const props = defineProps<{
	ui: TextField;
}>();

const networkStore = useNetworkStore();
const inputValue = ref<string>(props.ui.value.v);
const idPrefix = 'text-field-';

watch(inputValue, (newValue) => {
	networkStore.invokeFunctionsAndSetProperties([{
		...props.ui.value,
		v: newValue,
	}], [props.ui.onTextChanged]);
});
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
				/>
				<div v-if="inputValue" class="absolute top-0 bottom-0 right-4 flex items-center h-full">
					<CloseIcon class="w-4" tabindex="0" @click="inputValue = ''" @keydown.enter="inputValue = ''" />
				</div>
			</div>
		</InputWrapper>
	</div>
</template>

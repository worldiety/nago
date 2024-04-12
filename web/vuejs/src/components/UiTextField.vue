<script lang="ts" setup>
import { ref, watch } from 'vue';
import { useNetworkStore } from '@/stores/networkStore';
import type { LiveTextField } from '@/shared/model/liveTextField';
import type { LivePage } from '@/shared/model/livePage';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import CloseIcon from '@/assets/svg/close.svg';

const props = defineProps<{
	ui: LiveTextField;
	page: LivePage;
}>();

const networkStore = useNetworkStore();
const inputValue = ref<string>(props.ui.value.value);
const idPrefix = 'text-field-';

watch(inputValue, (newValue) => {
	networkStore.invokeFunctionsAndSetProperties([{
		...props.ui.value,
		value: newValue,
	}], [props.ui.onTextChanged]);
});
</script>

<template>
	<div>
		<InputWrapper
			:simple="props.ui.simple.value"
			:label="props.ui.label.value"
			:error="props.ui.error.value"
			:hint="props.ui.hint.value"
			:help="props.ui.help.value"
			:disabled="props.ui.disabled.value"
		>
			<div class="relative">
				<input
					:id="idPrefix + props.ui.id.toString()"
					v-model="inputValue"
					class="input-field"
					:class="{'!pr-10': inputValue}"
					:placeholder="props.ui.placeholder.value"
					:disabled="props.ui.disabled.value"
					type="text"
				/>
				<div v-if="inputValue" class="absolute top-0 bottom-0 right-4 flex items-center h-full">
					<CloseIcon class="w-4" tabindex="0" @click="inputValue = ''" @keydown.enter="inputValue = ''" />
				</div>
			</div>
		</InputWrapper>
	</div>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue';
import { useNetworkStore } from '@/stores/networkStore';
import type { LiveTextField } from '@/shared/model/liveTextField';
import type { LivePage } from '@/shared/model/livePage';
import InputWrapper from '@/components/shared/InputWrapper.vue';

const props = defineProps<{
	ui: LiveTextField;
	page: LivePage;
}>();

const networkStore = useNetworkStore();
const inputValue = ref<string>(props.ui.value.value);
const idPrefix = 'text-field-';

watch(inputValue, (newValue) => {
	networkStore.invokeFuncAndSetProp({
		...props.ui.value,
		value: newValue,
	}, props.ui.onTextChanged);
});
</script>

<template>
	<div>
		<label :for="idPrefix + props.ui.id.toString()" class="mb-2 block text-sm">
			{{ props.ui.label.value }}
		</label>
		<InputWrapper>
			<input
				:id="idPrefix + props.ui.id.toString()"
				v-model="inputValue"
				:placeholder="props.ui.placeholder.value"
				:disabled="props.ui.disabled.value"
				type="text"
			/>
		</InputWrapper>
		<p v-if="props.ui.error.value !== ''">{{ props.ui.error.value }}</p>
		<p v-else>{{ props.ui.hint.value }}</p>
	</div>
</template>

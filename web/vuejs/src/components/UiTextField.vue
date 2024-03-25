<script lang="ts" setup>
import { ref, watch } from 'vue';
import { useNetworkStore } from '@/stores/networkStore';
import type { LiveTextField } from '@/shared/model/liveTextField';
import type { LivePage } from '@/shared/model/livePage';

const props = defineProps<{
	ui: LiveTextField;
	page: LivePage;
}>();

const networkStore = useNetworkStore();
const inputValue = ref<string>(props.ui.value.value);

watch(inputValue, (newValue) => {
	networkStore.invokeFuncAndSetProp({
		...props.ui.value,
		value: newValue,
	}, props.ui.onTextChanged);
});

function isErr(): boolean {
	return props.ui.error.value != '';
}
</script>

<template>
	<div>
		<label :for="props.ui.id.toString()" class="mb-2 block text-sm">{{
			props.ui.label.value
		}}</label>
		<input
			v-model="inputValue"
			:disabled="props.ui.disabled.value"
			type="text"
			:id="props.ui.id.toString()"
		/>
		<p v-if="isErr()" class="mt-2 text-sm text-red-600 dark:text-red-500">{{ props.ui.error.value }}</p>
		<p v-if="!isErr()" class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ props.ui.hint.value }}</p>
	</div>
</template>

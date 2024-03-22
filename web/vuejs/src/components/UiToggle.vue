<script lang="ts" setup>
import { useNetworkStore } from '@/stores/networkStore';
import type { LiveToggle } from '@/shared/model/liveToggle';
import type { LivePage } from '@/shared/model/livePage';
import { ref } from 'vue';

const props = defineProps<{
	ui: LiveToggle;
	page: LivePage;
}>();

const networkStore = useNetworkStore();
const checked = ref<boolean>(props.ui.checked.value);

function onClick() {
	networkStore.invokeFuncAndSetProp(props.ui.checked, props.ui.onCheckedChanged);
}
</script>

<template>
	<label class="relative inline-flex cursor-pointer items-center">
		<input
			v-model="checked"
			type="checkbox"
			value=""
			class="peer sr-only"
			:checked="props.ui.checked.value"
			:disabled="props.ui.disabled.value"
			@change="onClick"
		/>
		<span
			v-if="ui.disabled.value"
			class="peer h-6 w-11 rounded-full bg-gray-200 after:absolute after:start-[2px] after:top-0.5 after:h-5 after:w-5 after:rounded-full after:border after:border-gray-300 after:bg-white after:transition-all after:content-[''] peer-checked:bg-gray-400 peer-checked:after:translate-x-full peer-checked:after:border-white rtl:peer-checked:after:-translate-x-full dark:border-gray-600 dark:bg-gray-700"
		></span>
		<span
			v-else
			class="peer h-6 w-11 rounded-full border border-black after:absolute after:start-[6px] after:top-1 after:h-4 after:w-4 after:rounded-full after:border after:border-gray-300 after:bg-white after:transition-all after:content-[''] peer-checked:after:translate-x-[105%] peer-checked:after:border-ora-orange peer-checked:after:bg-ora-orange peer-focus:outline-none peer-focus:outline-2 peer-focus:outline-offset-2 peer-focus:outline-black peer-focus:ring-white peer-focus:ring-2 rtl:peer-checked:after:-translate-x-full dark:border-white"
		></span>
		<span class="ms-3 text-sm font-medium text-gray-900 dark:text-gray-300">{{ props.ui.label.value }}</span>
	</label>
</template>

<script setup lang="ts">
import type { LiveDatepicker } from '@/shared/model/liveDatepicker';
import { ref } from 'vue';
import Calendar from '@/assets/svg/calendar.svg';
import { useNetworkStore } from '@/stores/networkStore';

const props = defineProps<{
	ui: LiveDatepicker;
}>();

const networkStore = useNetworkStore();
const date = ref('DD.MM.YYYY');
</script>

<template>
	<div>
		<span v-if="props.ui.label.value" class="block mb-2 text-sm font-medium">{{ props.ui.label.value }}</span>
		<div class="relative">
			<!-- Input field -->
			<div class="relative z-0">
				<input
					v-model="date"
					type="text"
					readonly
					:disabled="props.ui.disabled.value"
					class="input-field w-full pr-8"
					@click="networkStore.invokeFunc(props.ui.onToggleExpanded)"
				>
				<div class="absolute top-0 bottom-0 right-2 flex items-center pointer-events-none h-full">
					<Calendar class="w-4" :class="props.ui.disabled.value ? 'text-disabled-text' : 'text-black'" />
				</div>
			</div>

			<!-- Datepicker -->
			<div v-if="props.ui.expanded.value" class="absolute top-8 right-8 bg-white rounded-md shadow-lg h-64 w-64 z-10">

			</div>
		</div>
		<!-- Error message has precedence over hints -->
		<p v-if="props.ui.error.value" class="mt-2 text-sm text-red-600 dark:text-red-500">{{ props.ui.error.value }}</p>
		<p v-else-if="props.ui.hint.value" class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ props.ui.hint.value }}</p>
	</div>
</template>

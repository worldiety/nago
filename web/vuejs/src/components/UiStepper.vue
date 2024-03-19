<script lang="ts" setup>
import type { LiveStepper } from '@/shared/model/liveStepper';
import type { LivePage } from '@/shared/model/livePage';

const props = defineProps<{
	ui: LiveStepper;
	page: LivePage;
}>();
</script>

<template>
	<ol class="items-center space-y-4 rtl:space-x-reverse sm:flex sm:space-x-8 sm:space-y-0">
		<li
			v-for="(step, index) in props.ui.steps.value"
			:class="
				index == props.ui.selectedIndex.value
					? 'text-blue-600 dark:border-blue-500 dark:text-blue-500'
					: 'text-gray-500 dark:text-gray-400'
			"
			class="flex items-center space-x-2.5 rtl:space-x-reverse"
		>
			<span
				:class="index == props.ui.selectedIndex.value ? 'border-blue-600' : 'border-gray-500'"
				class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border"
			>
				{{ step.number.value }}
			</span>
			<span>
				<h3 class="font-medium leading-tight">{{ step.caption.value }}</h3>
				<p class="text-sm">{{ step.details.value }}</p>
			</span>
		</li>
	</ol>
</template>

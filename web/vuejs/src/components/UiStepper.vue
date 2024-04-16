<script lang="ts" setup>
import type { LivePage } from '@/shared/model/livePage';
import {Stepper} from "@/shared/protocol/gen/stepper";

const props = defineProps<{
	ui: Stepper;
	page: LivePage;
}>();
</script>

<template>
	<ol class="items-center space-y-4 rtl:space-x-reverse sm:flex sm:space-x-8 sm:space-y-0">
		<li
			v-for="(step, index) in props.ui.steps.v"
			:class="
				index == props.ui.selectedIndex.v
					? 'text-blue-600 dark:border-blue-500 dark:text-blue-500'
					: 'text-gray-500 dark:text-gray-400'
			"
			class="flex items-center space-x-2.5 rtl:space-x-reverse"
		>
			<span
				:class="index == props.ui.selectedIndex.v ? 'border-blue-600' : 'border-gray-500'"
				class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border"
			>
				{{ step.number.v }}
			</span>
			<span>
				<h3 class="font-medium leading-tight">{{ step.caption.v }}</h3>
				<p class="text-sm">{{ step.details.v }}</p>
			</span>
		</li>
	</ol>
</template>

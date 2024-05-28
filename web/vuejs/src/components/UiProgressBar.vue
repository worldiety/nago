<template>
	<div v-if="ui.max.v >= 0 && ui.value.v >= 0" class="flex flex-col justify-between items-start gap-y-1">
		<p v-if="ui.showPercentage.v" class="text-sm text-disabled-text">{{ percentage }}</p>
		<progress :max="ui.max.v" :value="ui.value.v"></progress>
	</div>
	<progress v-else></progress>
</template>

<script setup lang="ts">
import type { ProgressBar } from '@/shared/protocol/ora/progressBar';
import { computed } from 'vue';
import { localizeNumber } from '@/shared/localization';

const props = defineProps<{
	ui: ProgressBar;
}>();

const percentage = computed((): string => {
	if (props.ui.max.v === 0) {
		return '0%';
	}
	const percentage = Math.min(props.ui.value.v / props.ui.max.v * 100, 100);
	return `${localizeNumber(percentage, { maximumFractionDigits: 0 })}%`;
});
</script>

<script setup lang="ts">
import type { LiveDatepicker } from '@/shared/model/liveDatepicker';
import { computed, ref, watch } from 'vue';
import Calendar from '@/assets/svg/calendar.svg';
import ArrowDown from '@/assets/svg/arrowDown.svg';
import Close from '@/assets/svg/close.svg';
import { useNetworkStore } from '@/stores/networkStore';
import monthNames from '@/shared/monthNames';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import DatepickerOverlay from '@/components/datepicker/DatepickerOverlay.vue';

const props = defineProps<{
	ui: LiveDatepicker;
}>();

const networkStore = useNetworkStore();

const dateFormatted = computed((): string => {
	const date = new Date();
	date.setFullYear(props.ui.selectedYear.value, props.ui.selectedMonth.value - 1, props.ui.selectedDay.value);
	return date.toLocaleDateString();
});

function datepickerClicked(forceClose: boolean): void {
	if (!props.ui.disabled.value && (forceClose || !props.ui.expanded.value)) {
		networkStore.invokeFunc(props.ui.onClicked);
	}
}

function selectDay(day: number, month: number, year: number): void {
	networkStore.invokeSetProp({
		...props.ui.selectedDay,
		value: day,
	});
	networkStore.invokeSetProp({
		...props.ui.selectedMonth,
		value: month,
	});
	networkStore.invokeSetProp({
		...props.ui.selectedYear,
		value: year,
	});
	networkStore.invokeFunc(props.ui.onSelectionChanged);
}
</script>

<template>
	<div>
		<div class="relative">
			<!-- Input field -->
			<InputWrapper
				:label="props.ui.label.value"
				:error="props.ui.error.value"
				:hint="props.ui.hint.value"
				:disabled="props.ui.disabled.value"
			>
				<div
					class="input-field relative z-0"
					tabindex="0"
					@click="datepickerClicked(false)"
					@keydown.enter="datepickerClicked(true)">
					<p>{{ dateFormatted }}</p>
					<div class="absolute top-0 bottom-0 right-4 flex items-center pointer-events-none h-full">
						<Calendar class="w-4" />
					</div>
				</div>
			</InputWrapper>

			<DatepickerOverlay
				:expanded="props.ui.expanded.value"
				:label="props.ui.label.value"
				:selected-day="props.ui.selectedDay.value"
				:selected-month="props.ui.selectedMonth.value"
				:selected-year="props.ui.selectedYear.value"
				@close="datepickerClicked(true)"
				@select="selectDay"
			/>
		</div>
	</div>
</template>

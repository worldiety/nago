<script setup lang="ts">
import type { LiveDatepicker } from '@/shared/model/liveDatepicker';
import { computed } from 'vue';
import Calendar from '@/assets/svg/calendar.svg';
import { useNetworkStore } from '@/stores/networkStore';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import DatepickerOverlay from '@/components/datepicker/DatepickerOverlay.vue';
import type { PropertyBool } from '@/shared/model/propertyBool';
import { useI18n } from 'vue-i18n';

const props = defineProps<{
	ui: LiveDatepicker;
}>();

const { t } = useI18n();
const networkStore = useNetworkStore();

const dateFormatted = computed((): string => {
	if (!props.ui.startDateSelected.value || (props.ui.rangeMode.value && !props.ui.endDateSelected.value)) {
		return t('datepicker.select');
	}

	const startDate = new Date();
	startDate.setFullYear(props.ui.selectedStartYear.value, props.ui.selectedStartMonth.value - 1, props.ui.selectedStartDay.value);
	if (!props.ui.rangeMode.value || !props.ui.endDateSelected.value) {
		return startDate.toLocaleDateString();
	}
	const endDate = new Date();
	endDate.setFullYear(props.ui.selectedEndYear.value, props.ui.selectedEndMonth.value - 1, props.ui.selectedEndDay.value);
	return `${startDate.toLocaleDateString()} - ${endDate.toLocaleDateString()}`;
});

function showDatepicker(): void {
	if (!props.ui.disabled.value && !props.ui.expanded.value) {
		networkStore.invokeFunctions(props.ui.onClicked);
	}
}

function closeDatepicker(): void {
	if (props.ui.expanded.value) {
		networkStore.invokeFunctions(props.ui.onSelectionChanged);
	}
}

function selectDate(day: number, monthIndex: number, year: number): void {
	const selectedDate = new Date(year, monthIndex, day, 0, 0, 0, 0);
	if (!props.ui.rangeMode.value || !props.ui.startDateSelected.value) {
		selectStartDate(selectedDate);
		return;
	}
	const currentStartDate: Date = new Date(
		props.ui.selectedStartYear.value,
		props.ui.selectedStartMonth.value - 1,
		props.ui.selectedStartDay.value,
		0,
		0,
		0,
		0,
	);
	if (selectedDate.getTime() > currentStartDate.getTime()) {
		// If the selected date is after the current start date, set it as the end date
		selectEndDate(selectedDate);
	} else if (selectedDate.getTime() < currentStartDate.getTime()) {
		// If the selected date is before the current start date, set is as the start date
		selectStartDate(selectedDate);
		if (!props.ui.endDateSelected.value) {
			// If the no end date is selected yet, set the current start date as the end date
			selectEndDate(currentStartDate);
		}
	} else {
		if (!props.ui.endDateSelected.value) {
			// If the selected date is equal to the current start date and no end date has been selected yet, set the selected
			// date as the start and end date
			selectStartDate(selectedDate);
			selectEndDate(selectedDate);
		} else {
			// If the selected date is equal to the current start date and an end date has been selected yet, set the current
			// end date as the start date
			const currentEndDate: Date = new Date(
				props.ui.selectedEndYear.value,
				props.ui.selectedEndMonth.value - 1,
				props.ui.selectedEndDay.value,
				0,
				0,
				0,
				0,
			);
			selectStartDate(currentEndDate);
		}
	}
}

function selectStartDate(selectedDate: Date): void {
	networkStore.invokeSetProperties({
		...props.ui.selectedStartYear,
		value: selectedDate.getFullYear(),
	});
	networkStore.invokeSetProperties({
		...props.ui.selectedStartMonth,
		value: selectedDate.getMonth() + 1,
	});
	networkStore.invokeSetProperties({
		...props.ui.selectedStartDay,
		value: selectedDate.getDate(),
	});
	const startDateSelected: PropertyBool = {
		...props.ui.startDateSelected,
		value: true,
	};
	networkStore.invokeSetProperties(startDateSelected);
}

function selectEndDate(selectedDate: Date): void {
	networkStore.invokeSetProperties({
		...props.ui.selectedEndYear,
		value: selectedDate.getFullYear(),
	});
	networkStore.invokeSetProperties({
		...props.ui.selectedEndMonth,
		value: selectedDate.getMonth() + 1,
	});
	networkStore.invokeSetProperties({
		...props.ui.selectedEndDay,
		value: selectedDate.getDate(),
	});
	const endDateSelected: PropertyBool = {
		...props.ui.endDateSelected,
		value: true,
	};
	networkStore.invokeSetProperties(endDateSelected);
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
					@click="showDatepicker()"
					@keydown.enter="showDatepicker()">
					<p>{{ dateFormatted }}</p>
					<div class="absolute top-0 bottom-0 right-4 flex items-center pointer-events-none h-full">
						<Calendar class="w-4" />
					</div>
				</div>
			</InputWrapper>

			<DatepickerOverlay
				:expanded="props.ui.expanded.value"
				:range-mode="props.ui.rangeMode.value"
				:label="props.ui.label.value"
				:start-date-selected="props.ui.startDateSelected.value"
				:selected-start-day="props.ui.selectedStartDay.value"
				:selected-start-month="props.ui.selectedStartMonth.value"
				:selected-start-year="props.ui.selectedStartYear.value"
				:end-date-selected="props.ui.endDateSelected.value"
				:selected-end-day="props.ui.selectedEndDay.value"
				:selected-end-month="props.ui.selectedEndMonth.value"
				:selected-end-year="props.ui.selectedEndYear.value"
				@close="closeDatepicker()"
				@select="selectDate"
			/>
		</div>
	</div>
</template>

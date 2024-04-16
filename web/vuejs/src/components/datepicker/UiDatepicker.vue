<script setup lang="ts">
import { computed } from 'vue';
import Calendar from '@/assets/svg/calendar.svg';
import { useNetworkStore } from '@/stores/networkStore';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import DatepickerOverlay from '@/components/datepicker/DatepickerOverlay.vue';
import type { PropertyBool } from '@/shared/model/propertyBool';
import { useI18n } from 'vue-i18n';
import {DatePicker} from "@/shared/protocol/gen/datePicker";

const props = defineProps<{
	ui:DatePicker;
}>();

const { t } = useI18n();
const networkStore = useNetworkStore();

const dateFormatted = computed((): string => {
	if (!props.ui.startDateSelected.v || (props.ui.rangeMode.v && !props.ui.endDateSelected.v)) {
		return t('datepicker.select');
	}

	const startDate = new Date();
	startDate.setFullYear(props.ui.selectedStartYear.v, props.ui.selectedStartMonth.v - 1, props.ui.selectedStartDay.v);
	if (!props.ui.rangeMode.v || !props.ui.endDateSelected.v) {
		return startDate.toLocaleDateString();
	}
	const endDate = new Date();
	endDate.setFullYear(props.ui.selectedEndYear.v, props.ui.selectedEndMonth.v - 1, props.ui.selectedEndDay.v);
	return `${startDate.toLocaleDateString()} - ${endDate.toLocaleDateString()}`;
});

function showDatepicker(): void {
	if (!props.ui.disabled.v && !props.ui.expanded.v) {
		networkStore.invokeFunctions(props.ui.onClicked);
	}
}

function closeDatepicker(): void {
	if (props.ui.expanded.v) {
		networkStore.invokeFunctions(props.ui.onSelectionChanged);
	}
}

function selectDate(day: number, monthIndex: number, year: number): void {
	const selectedDate = new Date(year, monthIndex, day, 0, 0, 0, 0);
	if (!props.ui.rangeMode.v || !props.ui.startDateSelected.v) {
		selectStartDate(selectedDate);
		return;
	}
	const currentStartDate: Date = new Date(
		props.ui.selectedStartYear.v,
		props.ui.selectedStartMonth.v - 1,
		props.ui.selectedStartDay.v,
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
		if (!props.ui.endDateSelected.v) {
			// If the no end date is selected yet, set the current start date as the end date
			selectEndDate(currentStartDate);
		}
	} else {
		if (!props.ui.endDateSelected.v) {
			// If the selected date is equal to the current start date and no end date has been selected yet, set the selected
			// date as the start and end date
			selectStartDate(selectedDate);
			selectEndDate(selectedDate);
		} else {
			// If the selected date is equal to the current start date and an end date has been selected yet, set the current
			// end date as the start date
			const currentEndDate: Date = new Date(
				props.ui.selectedEndYear.v,
				props.ui.selectedEndMonth.v - 1,
				props.ui.selectedEndDay.v,
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
	networkStore.invokeSetProperties(
		{
			...props.ui.selectedStartYear,
			v: selectedDate.getFullYear(),
		},
		{
			...props.ui.selectedStartMonth,
			v: selectedDate.getMonth() + 1,
		},
		{
			...props.ui.selectedStartDay,
			v: selectedDate.getDate(),
		},
		{
			...props.ui.startDateSelected,
			v: true,
		},
	);
}

function selectEndDate(selectedDate: Date): void {
	networkStore.invokeSetProperties(
		{
			...props.ui.selectedEndYear,
			v: selectedDate.getFullYear(),
		},
		{
			...props.ui.selectedEndMonth,
			v: selectedDate.getMonth() + 1,
		},
		{
			...props.ui.selectedEndDay,
			v: selectedDate.getDate(),
		},
		{
			...props.ui.endDateSelected,
			v: true,
		},
	);
}
</script>

<template>
	<div>
		<div class="relative">
			<!-- Input field -->
			<InputWrapper
				:label="props.ui.label.v"
				:error="props.ui.error.v"
				:hint="props.ui.hint.v"
				:disabled="props.ui.disabled.v"
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
				:expanded="props.ui.expanded.v"
				:range-mode="props.ui.rangeMode.v"
				:label="props.ui.label.v"
				:start-date-selected="props.ui.startDateSelected.v"
				:selected-start-day="props.ui.selectedStartDay.v"
				:selected-start-month="props.ui.selectedStartMonth.v"
				:selected-start-year="props.ui.selectedStartYear.v"
				:end-date-selected="props.ui.endDateSelected.v"
				:selected-end-day="props.ui.selectedEndDay.v"
				:selected-end-month="props.ui.selectedEndMonth.v"
				:selected-end-year="props.ui.selectedEndYear.v"
				@close="closeDatepicker()"
				@select="selectDate"
			/>
		</div>
	</div>
</template>

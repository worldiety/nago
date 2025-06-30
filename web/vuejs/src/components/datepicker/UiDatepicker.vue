<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template>
	<div v-if="!ui.invisible" class="relative" :style="frameStyles">
		<!-- Input field -->
		<DatepickerInput
			:value="ui.value"
			:label="ui.label"
			:error-text="ui.errorText"
			:supporting-text="ui.supportingText"
			:disabled="ui.disabled"
			:datepicker-style="ui.style"
			:datepicker-expanded="expanded"
			:selected-start-year="selectedStartYear"
			:selected-start-month="selectedStartMonth"
			:selected-start-day="selectedStartDay"
			:selected-end-year="selectedEndYear"
			:selected-end-month="selectedEndMonth"
			:selected-end-day="selectedEndDay"
			:range-mode="ui.style === DatePickerStyleValues.DatePickerDateRange"
			:input-value="ui.inputValue"
			:end-input-value="ui.endInputValue"
			@show-datepicker="showDatepicker"
		/>

		<DatepickerOverlay
			:datepicker-expanded="expanded"
			:range-mode="ui.style === DatePickerStyleValues.DatePickerDateRange"
			:label="ui.label"
			:selected-start-day="selectedStartDay"
			:selected-start-month="selectedStartMonth"
			:selected-start-year="selectedStartYear"
			:selected-end-day="selectedEndDay"
			:selected-end-month="selectedEndMonth"
			:selected-end-year="selectedEndYear"
			:range-selection-state="rangeSelectionState"
			@close="closeDatepicker"
			@select="selectDate"
			@submit-selection="submitSelection"
			@clear-selection="initialize"
		/>
	</div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import DatepickerInput from '@/components/datepicker/DatepickerInput.vue';
import DatepickerOverlay from '@/components/datepicker/DatepickerOverlay.vue';
import { RangeSelectionState } from '@/components/datepicker/rangeSelectionState';
import { frameCSS } from '@/components/shared/frame';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import {
	DateData,
	DatePicker,
	DatePickerStyleValues,
	UpdateStateValueRequested,
	UpdateStateValues2Requested,
} from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: DatePicker;
}>();

const serviceAdapter = useServiceAdapter();
const expanded = ref<boolean>(false);
const selectedStartDay = ref<number>(0);
const selectedStartMonth = ref<number>(0);
const selectedStartYear = ref<number>(0);
const selectedEndDay = ref<number>(0);
const selectedEndMonth = ref<number>(0);
const selectedEndYear = ref<number>(0);
const rangeSelectionState = ref<RangeSelectionState>(RangeSelectionState.SELECT_START);

onMounted(initialize);

const frameStyles = computed<string>(() => {
	return frameCSS(props.ui.frame).join(';');
});

watch(() => props.ui, initialize);

function initialize(): void {
	if (props.ui.value === undefined) {
		props.ui.value = new DateData();
	}
	selectedStartDay.value = props.ui.value.day ? props.ui.value.day : 0;
	selectedStartMonth.value = props.ui.value.month ? props.ui.value.month : 0;
	selectedStartYear.value = props.ui.value.year ? props.ui.value.year : 0;

	if (props.ui.endValue === undefined) {
		props.ui.endValue = new DateData();
	}
	selectedEndDay.value = props.ui.endValue.day ? props.ui.endValue.day : 0;
	selectedEndMonth.value = props.ui.endValue.month ? props.ui.endValue.month : 0;
	selectedEndYear.value = props.ui.endValue.year ? props.ui.endValue.year : 0;

	rangeSelectionState.value = RangeSelectionState.SELECT_START;
}

function showDatepicker(): void {
	if (!props.ui.disabled && !expanded.value) {
		rangeSelectionState.value = RangeSelectionState.SELECT_START;
		expanded.value = true;
	}
}

function closeDatepicker(): void {
	expanded.value = false;
	initialize();
}

function selectDate(day: number, monthIndex: number, year: number): void {
	const selectedDate = new Date(year, monthIndex, day, 0, 0, 0, 0);

	if (props.ui.style === DatePickerStyleValues.DatePickerSingleDate) {
		selectStartDate(selectedDate);
		return;
	}

	switch (rangeSelectionState.value) {
		case RangeSelectionState.SELECT_START:
			selectStartDate(selectedDate);
			break;
		case RangeSelectionState.SELECT_END:
			selectEndDate(selectedDate);
			break;
		case RangeSelectionState.COMPLETE:
			selectStartDate(selectedDate);
			break;
	}
}

function selectStartDate(selectedDate: Date): void {
	selectedStartDay.value = selectedDate.getDate();
	selectedStartMonth.value = selectedDate.getMonth() + 1;
	selectedStartYear.value = selectedDate.getFullYear();

	if (props.ui.style !== DatePickerStyleValues.DatePickerDateRange) {
		serviceAdapter.sendEvent(
			new UpdateStateValueRequested(
				props.ui.inputValue,
				0,
				nextRID(),
				JSON.stringify({
					d: selectedStartDay.value,
					m: selectedStartMonth.value,
					y: selectedStartYear.value,
				})
			)
		);
	} else {
		rangeSelectionState.value = RangeSelectionState.SELECT_END;
	}
}

function selectEndDate(selectedDate: Date): void {
	selectedDate.setHours(0, 0, 0, 0);
	const selectedStartDate = new Date(
		selectedStartYear.value,
		selectedStartMonth.value - 1,
		selectedStartDay.value,
		0,
		0,
		0,
		0
	);
	if (selectedDate < selectedStartDate) {
		// selected date is before currently selected start date so we switch start and end dates here
		selectedStartDay.value = selectedDate.getDate();
		selectedStartMonth.value = selectedDate.getMonth() + 1;
		selectedStartYear.value = selectedDate.getFullYear();

		selectedEndDay.value = selectedStartDate.getDate();
		selectedEndMonth.value = selectedStartDate.getMonth() + 1;
		selectedEndYear.value = selectedStartDate.getFullYear();
	} else {
		// selected date equals or is after currently selected start date so we just have to set it as the end date
		selectedEndDay.value = selectedDate.getDate();
		selectedEndMonth.value = selectedDate.getMonth() + 1;
		selectedEndYear.value = selectedDate.getFullYear();
	}
	rangeSelectionState.value = RangeSelectionState.COMPLETE;
}

function submitSelection(): void {
	expanded.value = false;

	switch (props.ui.style) {
		case DatePickerStyleValues.DatePickerSingleDate: {
			serviceAdapter.sendEvent(
				new UpdateStateValueRequested(
					props.ui.inputValue,
					0,
					nextRID(),
					JSON.stringify({
						d: selectedStartDay.value,
						m: selectedStartMonth.value,
						y: selectedStartYear.value,
					})
				)
			);

			return;
		}
		case DatePickerStyleValues.DatePickerDateRange: {
			serviceAdapter.sendEvent(
				new UpdateStateValues2Requested(
					props.ui.inputValue,
					JSON.stringify({
						d: selectedStartDay.value,
						m: selectedStartMonth.value,
						y: selectedStartYear.value,
					}),
					props.ui.endInputValue,
					JSON.stringify({
						d: selectedEndDay.value,
						m: selectedEndMonth.value,
						y: selectedEndYear.value,
					}),
					0,
					nextRID()
				)
			);

			return;
		}
		default:
			throw 'unknown date picker style';
	}
}
</script>

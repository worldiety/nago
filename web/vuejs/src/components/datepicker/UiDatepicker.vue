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
			:selected-start-date="selectedStartDate"
			:selected-end-date="selectedEndDate"
			:range-mode="ui.style === DatePickerStyleValues.DatePickerDateRange"
			:input-value="ui.inputValue"
			:end-input-value="ui.endInputValue"
			@show-datepicker="showDatepicker"
		/>

		<DatepickerOverlay
			v-if="expanded"
			:range-mode="ui.style === DatePickerStyleValues.DatePickerDateRange"
			:double-mode="ui.doubleMode"
			:label="ui.label"
			:selected-start-date="selectedStartDate"
			:selected-end-date="selectedEndDate"
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
import type { DatePicker } from '@/shared/proto/nprotoc_gen';
import {
	DateData,
	DatePickerStyleValues,
	UpdateStateValueRequested,
	UpdateStateValues2Requested,
} from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: DatePicker;
}>();

const serviceAdapter = useServiceAdapter();
const expanded = ref<boolean>(false);
const selectedStartDate = ref<Date>();
const selectedEndDate = ref<Date>();
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

	const hasValue =
		props.ui.value.year !== undefined && props.ui.value.month !== undefined && props.ui.value.day !== undefined;
	selectedStartDate.value = hasValue
		? new Date(props.ui.value.year!, props.ui.value.month! - 1, props.ui.value.day, 0, 0, 0, 0)
		: undefined;

	if (props.ui.endValue === undefined) {
		props.ui.endValue = new DateData();
	}

	const hasEndValue =
		props.ui.endValue.year !== undefined &&
		props.ui.endValue.month !== undefined &&
		props.ui.endValue.day !== undefined;
	selectedEndDate.value = hasEndValue
		? new Date(props.ui.endValue.year!, props.ui.endValue.month! - 1, props.ui.endValue.day, 0, 0, 0, 0)
		: undefined;

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
	selectedStartDate.value = new Date(selectedDate);
	selectedStartDate.value.setHours(0, 0, 0, 0);

	if (props.ui.style !== DatePickerStyleValues.DatePickerDateRange) {
		serviceAdapter.sendEvent(
			new UpdateStateValueRequested(
				props.ui.inputValue,
				0,
				nextRID(),
				JSON.stringify({
					d: selectedStartDate.value.getDate(),
					m: selectedStartDate.value.getMonth() + 1,
					y: selectedStartDate.value.getFullYear(),
				})
			)
		);
	} else {
		rangeSelectionState.value = RangeSelectionState.SELECT_END;
	}
}

function selectEndDate(selectedDate: Date): void {
	if (!selectedStartDate.value) return;

	selectedDate.setHours(0, 0, 0, 0);
	if (selectedDate < selectedStartDate.value) {
		// selected date is before currently selected start date so we switch start and end dates here
		selectedEndDate.value = new Date(selectedStartDate.value);
		selectedStartDate.value = new Date(selectedDate);
	} else {
		// selected date equals or is after currently selected start date so we just have to set it as the end date
		selectedEndDate.value = new Date(selectedDate);
	}
	rangeSelectionState.value = RangeSelectionState.COMPLETE;
}

function submitSelection(): void {
	expanded.value = false;
	if (
		!selectedStartDate.value ||
		(props.ui.style === DatePickerStyleValues.DatePickerDateRange && !selectedEndDate.value)
	)
		return;

	switch (props.ui.style) {
		case DatePickerStyleValues.DatePickerSingleDate: {
			serviceAdapter.sendEvent(
				new UpdateStateValueRequested(
					props.ui.inputValue,
					0,
					nextRID(),
					JSON.stringify({
						d: selectedStartDate.value.getDate(),
						m: selectedStartDate.value.getMonth() + 1,
						y: selectedStartDate.value.getFullYear(),
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
						d: selectedStartDate.value.getDate(),
						m: selectedStartDate.value.getMonth() + 1,
						y: selectedStartDate.value.getFullYear(),
					}),
					props.ui.endInputValue,
					JSON.stringify({
						d: selectedEndDate.value!.getDate(),
						m: selectedEndDate.value!.getMonth() + 1,
						y: selectedEndDate.value!.getFullYear(),
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

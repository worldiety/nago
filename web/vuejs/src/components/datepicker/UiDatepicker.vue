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
			:id="id"
			:value="value"
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
			:min-year="MIN_YEAR"
			:max-year="MAX_YEAR"
			:optional="ui.optional"
			@show-datepicker="showDatepicker"
		/>

		<DatepickerOverlay
			v-if="expanded"
			:range-mode="ui.style === DatePickerStyleValues.DatePickerDateRange"
			:double-mode="isDoubleMode"
			:label="ui.label"
			:selected-start-date="selectedStartDate"
			:selected-end-date="selectedEndDate"
			:range-selection-state="rangeSelectionState"
			:min-year="MIN_YEAR"
			@close="closeDatepicker"
			@select="selectDate"
			@submit-selection="submitSelection"
			@clear-selection="initialize"
		/>
	</div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
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
import { randomStr } from '@/components/shared/util';

const props = defineProps<{
	ui: DatePicker;
}>();

const MIN_YEAR = 1583;
const MAX_YEAR = 9999;

const id = `datepicker-${randomStr(16)}`;
const serviceAdapter = useServiceAdapter();
const viewportWidth = ref<number>(document.body.clientWidth);

const value = ref(props.ui.value);
const endValue = ref(props.ui.endValue);
const expanded = ref<boolean>(false);
const selectedStartDate = ref<Date>();
const selectedEndDate = ref<Date>();
const rangeSelectionState = ref<RangeSelectionState>(RangeSelectionState.SELECT_START);

onMounted(initialize);

const frameStyles = computed<string>(() => {
	return frameCSS(props.ui.frame).join(';');
});

const isDoubleMode = computed<boolean>(() => {
	return !!props.ui.doubleMode && viewportWidth.value > 600;
})

function initialize(): void {
	if (value.value === undefined) {
		value.value = new DateData();
	}

	const hasValue = value.value.year !== undefined && value.value.month !== undefined && value.value.day !== undefined;
	selectedStartDate.value = hasValue
		? new Date(value.value.year!, value.value.month! - 1, value.value.day, 0, 0, 0, 0)
		: undefined;

	if (endValue.value === undefined) {
		endValue.value = new DateData();
	}

	const hasEndValue =
		endValue.value.year !== undefined && endValue.value.month !== undefined && endValue.value.day !== undefined;
	selectedEndDate.value = hasEndValue
		? new Date(endValue.value.year!, endValue.value.month! - 1, endValue.value.day, 0, 0, 0, 0)
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

function onViewportResize() {
	viewportWidth.value = document.body.clientWidth;
}

watch(
	() => props.ui.value,
	() => {
		value.value = props.ui.value;
		initialize();
	}
);

watch(
	() => props.ui.endValue,
	() => {
		endValue.value = props.ui.endValue;
		initialize();
	}
);

onMounted(() => {
	addEventListener('resize', onViewportResize);
});

onBeforeUnmount(() => {
	removeEventListener('resize', onViewportResize);
});
</script>

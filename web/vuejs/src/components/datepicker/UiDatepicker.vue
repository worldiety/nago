<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template>
	<div v-if="!ui.invisible" class="relative" :style="frameStyles">
		<!-- Input field -->
		<InputWrapper
			:label="props.ui.label"
			:error="props.ui.errorText"
			:hint="props.ui.supportingText"
			:disabled="props.ui.disabled"
		>
			<div
				class="input-field relative z-0 !pr-10"
				tabindex="0"
				@click="showDatepicker"
				@keydown.enter="showDatepicker"
			>
				<p :class="{ 'text-placeholder-text': !dateFormatted }">
					{{ dateFormatted ?? $t('datepicker.select') }}
				</p>
				<div class="absolute top-0 bottom-0 right-4 flex items-center pointer-events-none h-full">
					<Calendar class="w-4" />
				</div>
			</div>
		</InputWrapper>

		<DatepickerOverlay
			:expanded="expanded"
			:range-mode="props.ui.style == DatePickerStyleValues.DatePickerDateRange"
			:label="props.ui.label"
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
		/>
	</div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import Calendar from '@/assets/svg/calendar.svg';
import DatepickerOverlay from '@/components/datepicker/DatepickerOverlay.vue';
import { RangeSelectionState } from '@/components/datepicker/rangeSelectionState';
import InputWrapper from '@/components/shared/InputWrapper.vue';
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

const dateFormatted = computed((): string | null => {
	if (!props.ui.value) {
		return null;
	}

	if (!props.ui.value.year) {
		return null;
	}

	const startDate = new Date();
	startDate.setFullYear(selectedStartYear.value, selectedStartMonth.value - 1, selectedStartDay.value);
	if (props.ui.style !== DatePickerStyleValues.DatePickerDateRange) {
		//console.log("bugs!!",startDate.toLocaleDateString())
		return startDate.toLocaleDateString();
	}
	const endDate = new Date();
	endDate.setFullYear(selectedEndYear.value, selectedEndMonth.value - 1, selectedEndDay.value);
	return `${startDate.toLocaleDateString()} - ${endDate.toLocaleDateString()}`;
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
		case RangeSelectionState.CLEAR:
			initialize();
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
	}
	rangeSelectionState.value = RangeSelectionState.SELECT_END;
}

function selectEndDate(selectedDate: Date): void {
	selectedEndDay.value = selectedDate.getDate();
	selectedEndMonth.value = selectedDate.getMonth() + 1;
	selectedEndYear.value = selectedDate.getFullYear();
	rangeSelectionState.value = RangeSelectionState.CLEAR;
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

<template>
	<div v-if="!ui.invisible.value" class="relative" :style="frameStyles">
		<!-- Input field -->
		<InputWrapper
			:label="props.ui.label.value"
			:error="props.ui.errorText.value"
			:hint="props.ui.supportingText.value"
			:disabled="props.ui.disabled.value"
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
				<div class="absolute top-0 bottom-0 right-4 flex items-center pointer-events-none text-black h-full">
					<Calendar class="w-4" />
				</div>
			</div>
		</InputWrapper>

		<DatepickerOverlay
			:expanded="expanded"
			:range-mode="props.ui.style.value == DatePickerStyleValues.DatePickerDateRange"
			:label="props.ui.label.value"
			:start-date-selected="startDateSelected"
			:selected-start-day="selectedStartDay"
			:selected-start-month="selectedStartMonth"
			:selected-start-year="selectedStartYear"
			:end-date-selected="endDateSelected"
			:selected-end-day="selectedEndDay"
			:selected-end-month="selectedEndMonth"
			:selected-end-year="selectedEndYear"
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
import InputWrapper from '@/components/shared/InputWrapper.vue';
import { frameCSS } from '@/components/shared/frame';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import {
	DatePicker,
	DatePickerStyleValues,
	Ptr,
	Str,
	UpdateStateValueRequested,
	UpdateStateValues2Requested,
} from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: DatePicker;
}>();

const expanded = ref<boolean>(false);

const serviceAdapter = useServiceAdapter();
const selectedStartDay = ref<number>(0);
const selectedStartMonth = ref<number>(0);
const selectedStartYear = ref<number>(0);
const startDateSelected = ref<boolean>(false);
const selectedEndDay = ref<number>(0);
const selectedEndMonth = ref<number>(0);
const selectedEndYear = ref<number>(0);
const endDateSelected = ref<boolean>(false);

onMounted(initialize);

watch(
	() => props.ui,
	(newValue) => {
		initialize();
	}
);

function initialize(): void {
	selectedStartDay.value = props.ui.value.day.value;
	selectedStartMonth.value = props.ui.value.month.value;
	selectedStartYear.value = props.ui.value.year.value;

	startDateSelected.value =
		props.ui.style.value == DatePickerStyleValues.DatePickerDateRange && selectedStartYear.value != 0;

	selectedEndDay.value = props.ui.endValue.day.value;
	selectedEndMonth.value = props.ui.endValue.month.value;
	selectedEndYear.value = props.ui.endValue.year.value;

	endDateSelected.value =
		props.ui.style.value == DatePickerStyleValues.DatePickerDateRange && selectedEndYear.value != 0;

	if (props.ui.style.value == DatePickerStyleValues.DatePickerSingleDate) {
		startDateSelected.value = true;
		endDateSelected.value = true;
	}

	//console.log("start", selectedStartDay.value, selectedStartMonth.value, selectedStartYear.value)
	//console.log("ende", selectedEndDay.value, selectedEndMonth.value, selectedEndYear.value)
}

const dateFormatted = computed((): string | null => {
	if (props.ui.value.year.isZero()) {
		return null;
	}

	const startDate = new Date();
	startDate.setFullYear(selectedStartYear.value, selectedStartMonth.value - 1, selectedStartDay.value);
	if (props.ui.style.value !== DatePickerStyleValues.DatePickerDateRange) {
		//console.log("bugs!!",startDate.toLocaleDateString())
		return startDate.toLocaleDateString();
	}
	const endDate = new Date();
	endDate.setFullYear(selectedEndYear.value, selectedEndMonth.value - 1, selectedEndDay.value);
	return `${startDate.toLocaleDateString()} - ${endDate.toLocaleDateString()}`;
});

function showDatepicker(): void {
	if (!props.ui.disabled.value && !expanded.value) {
		expanded.value = true;
	}
}

function closeDatepicker(): void {
	expanded.value = false;
	// TODO fix me: range style-only: the range is not resetted when closed without submission
}

function selectDate(day: number, monthIndex: number, year: number): void {
	const selectedDate = new Date(year, monthIndex, day, 0, 0, 0, 0);
	if (props.ui.style.value != DatePickerStyleValues.DatePickerDateRange || !startDateSelected.value) {
		selectStartDate(selectedDate);
		return;
	}
	const currentStartDate: Date = new Date(
		selectedStartYear.value,
		selectedStartMonth.value - 1,
		selectedStartDay.value,
		0,
		0,
		0,
		0
	);
	if (selectedDate.getTime() > currentStartDate.getTime()) {
		// If the selected date is after the current start date, set it as the end date
		selectEndDate(selectedDate);
	} else if (selectedDate.getTime() < currentStartDate.getTime()) {
		// If the selected date is before the current start date, set is as the start date
		selectStartDate(selectedDate);
		if (!endDateSelected.value) {
			// If the no end date is selected yet, set the current start date as the end date
			selectEndDate(currentStartDate);
		}
	} else {
		if (!endDateSelected.value) {
			// If the selected date is equal to the current start date and no end date has been selected yet, set the selected
			// date as the start and end date
			selectStartDate(selectedDate);
			selectEndDate(selectedDate);
		} else {
			// If the selected date is equal to the current start date and an end date has been selected yet, set the current
			// end date as the start date
			const currentEndDate: Date = new Date(
				selectedEndYear.value,
				selectedEndMonth.value - 1,
				selectedEndDay.value,
				0,
				0,
				0,
				0
			);
			selectStartDate(currentEndDate);
		}
	}
}

interface MyDate {
	// Day
	d /*omitempty*/? /*Day*/ : number /*int*/;
	// Month
	m /*omitempty*/? /*Month*/ : number /*int*/;
	// Year
	y /*omitempty*/? /*Year*/ : number /*int*/;
}

function selectStartDate(selectedDate: Date): void {
	selectedStartDay.value = selectedDate.getDate();
	selectedStartMonth.value = selectedDate.getMonth() + 1;
	selectedStartYear.value = selectedDate.getFullYear();
	startDateSelected.value = true;
	if (props.ui.style.value !== DatePickerStyleValues.DatePickerDateRange) {
		serviceAdapter.sendEvent(
			new UpdateStateValueRequested(
				props.ui.inputValue,
				new Ptr(),
				nextRID(),
				new Str(
					JSON.stringify({
						d: selectedStartDay.value,
						m: selectedStartMonth.value,
						y: selectedStartYear.value,
					})
				)
			)
		);
	}
}

function selectEndDate(selectedDate: Date): void {
	selectedEndDay.value = selectedDate.getDate();
	selectedEndMonth.value = selectedDate.getMonth() + 1;
	selectedEndYear.value = selectedDate.getFullYear();
	endDateSelected.value = true;
}

function submitSelection(): void {
	expanded.value = false;

	switch (props.ui.style.value) {
		case DatePickerStyleValues.DatePickerSingleDate: {
			serviceAdapter.sendEvent(
				new UpdateStateValueRequested(
					props.ui.inputValue,
					new Ptr(),
					nextRID(),
					new Str(
						JSON.stringify({
							d: selectedStartDay.value,
							m: selectedStartMonth.value,
							y: selectedStartYear.value,
						})
					)
				)
			);

			return;
		}
		case DatePickerStyleValues.DatePickerDateRange: {
			serviceAdapter.sendEvent(
				new UpdateStateValues2Requested(
					props.ui.inputValue,
					new Str(
						JSON.stringify({
							d: selectedStartDay.value,
							m: selectedStartMonth.value,
							y: selectedStartYear.value,
						})
					),

					props.ui.endInputValue,
					new Str(
						JSON.stringify({
							d: selectedEndDay.value,
							m: selectedEndMonth.value,
							y: selectedEndYear.value,
						})
					),
					new Ptr(),
					nextRID()
				)
			);

			return;
		}
		default:
			throw 'unknown date picker style';
	}
}

const frameStyles = computed<string>(() => {
	return frameCSS(props.ui.frame).join(';');
});
</script>

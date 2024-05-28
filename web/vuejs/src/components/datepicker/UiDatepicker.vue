<template>
	<div v-if="ui.visible.v">
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
					@click="showDatepicker"
					@keydown.enter="showDatepicker">
					<p :class="{'text-placeholder-text': !dateFormatted}">{{ dateFormatted ?? $t('datepicker.select') }}</p>
					<div class="absolute top-0 bottom-0 right-4 flex items-center pointer-events-none text-black dark:text-white h-full">
						<Calendar class="w-4" />
					</div>
				</div>
			</InputWrapper>

			<DatepickerOverlay
				:expanded="props.ui.expanded.v"
				:range-mode="props.ui.rangeMode.v"
				:label="props.ui.label.v"
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
	</div>
</template>


<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import Calendar from '@/assets/svg/calendar.svg';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import DatepickerOverlay from '@/components/datepicker/DatepickerOverlay.vue';
import type {DatePicker} from "@/shared/protocol/ora/datePicker";
import { useServiceAdapter } from '@/composables/serviceAdapter';

const props = defineProps<{
	ui: DatePicker;
}>();

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

watch(() => props.ui.expanded.v, (newValue) => {
	if (!newValue) {
		initialize();
	}
});

function initialize(): void {
	selectedStartDay.value = props.ui.selectedStartDay.v;
	selectedStartMonth.value = props.ui.selectedStartMonth.v;
	selectedStartYear.value = props.ui.selectedStartYear.v;
	startDateSelected.value = props.ui.startDateSelected.v;
	selectedEndDay.value = props.ui.selectedEndDay.v;
	selectedEndMonth.value = props.ui.selectedEndMonth.v;
	selectedEndYear.value = props.ui.selectedEndYear.v;
	endDateSelected.value = props.ui.endDateSelected.v;
}

const dateFormatted = computed((): string|null => {
	if (!props.ui.startDateSelected.v || (props.ui.rangeMode.v && !props.ui.endDateSelected.v)) {
		return null;
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
		serviceAdapter.setPropertiesAndCallFunctions([
				{
					...props.ui.expanded,
					v: true,
				},
			], [props.ui.onClicked],
		);
	}
}

function closeDatepicker(): void {
	if (props.ui.expanded.v) {
		serviceAdapter.setProperties({
			...props.ui.expanded,
			v: false,
		});
	}
}

function selectDate(day: number, monthIndex: number, year: number): void {
	const selectedDate = new Date(year, monthIndex, day, 0, 0, 0, 0);
	if (!props.ui.rangeMode.v || !startDateSelected.value) {
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
		0,
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
				0,
			);
			selectStartDate(currentEndDate);
		}
	}
}

function selectStartDate(selectedDate: Date): void {
	selectedStartDay.value = selectedDate.getDate();
	selectedStartMonth.value = selectedDate.getMonth() + 1;
	selectedStartYear.value = selectedDate.getFullYear();
	startDateSelected.value = true;
	if (!props.ui.rangeMode.v) {
		serviceAdapter.setPropertiesAndCallFunctions([
				{
					...props.ui.selectedStartYear,
					v: selectedStartYear.value,
				},
				{
					...props.ui.selectedStartMonth,
					v: selectedStartMonth.value,
				},
				{
					...props.ui.selectedStartDay,
					v: selectedStartDay.value,
				},
				{
					...props.ui.startDateSelected,
					v: true,
				},
			], [props.ui.onSelectionChanged],
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
	serviceAdapter.setPropertiesAndCallFunctions([
			{
				...props.ui.selectedStartYear,
				v: selectedStartYear.value,
			},
			{
				...props.ui.selectedStartMonth,
				v: selectedStartMonth.value,
			},
			{
				...props.ui.selectedStartDay,
				v: selectedStartDay.value,
			},
			{
				...props.ui.startDateSelected,
				v: true,
			},
			{
				...props.ui.selectedEndYear,
				v: selectedEndYear.value,
			},
			{
				...props.ui.selectedEndMonth,
				v: selectedEndMonth.value,
			},
			{
				...props.ui.selectedEndDay,
				v: selectedEndDay.value,
			},
			{
				...props.ui.endDateSelected,
				v: true,
			},
			{
				...props.ui.expanded,
				v: false,
			},
		], [props.ui.onSelectionChanged],
	);
}
</script>

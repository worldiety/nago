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
	if (!props.ui.startDateSelected.value) {
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

function datepickerClicked(forceClose: boolean): void {
	if (!props.ui.disabled.value && (forceClose || !props.ui.expanded.value)) {
		networkStore.invokeFunc(props.ui.onClicked);
	}
}

function submit(selectedStartDate: Date, selectedEndDate: Date|null): void {
	if (!props.ui.rangeMode.value || !selectedEndDate) {
		selectStartDate(selectedStartDate);
		datepickerClicked(true);
		return;
	}

	if (selectedStartDate <= selectedEndDate) {
		selectStartDate(selectedStartDate);
		selectEndDate(selectedEndDate);
	} else if (selectedStartDate > selectedEndDate) {
		selectStartDate(selectedEndDate);
		selectEndDate(selectedStartDate);
	}
	datepickerClicked(true);
}

function selectDay(day: number, month: number, year: number): void {
	const selectedDate = new Date();
	selectedDate.setFullYear(year, month, day);
	selectedDate.setHours(0, 0, 0, 0);
	if (!props.ui.rangeMode.value || !props.ui.startDateSelected.value) {
		selectFirstDate(selectedDate);
		return;
	}

	selectSecondDate(selectedDate);
	networkStore.invokeFunc(props.ui.onSelectionChanged);
}

function selectFirstDate(selectedDate: Date): void {
	if (props.ui.endDateSelected.value) {
		const currentEndDate = new Date();
		currentEndDate.setFullYear(props.ui.selectedEndYear.value, props.ui.selectedEndMonth.value, props.ui.selectedEndDay.value);
		currentEndDate.setHours(0, 0, 0, 0);
		if (selectedDate > currentEndDate) {
			selectStartDate(currentEndDate);
			selectEndDate(selectedDate);
		} else if (selectedDate < currentEndDate) {
			selectStartDate(selectedDate);
		} else {
			const startDateSelected: PropertyBool = {
				...props.ui.startDateSelected,
				value: false,
			};
			networkStore.invokeSetProp(startDateSelected);
		}
	} else {
		selectStartDate(selectedDate);
	}
	networkStore.invokeFunc(props.ui.onSelectionChanged);
}

function selectSecondDate(selectedDate: Date): void {
	const currentStartDate = new Date();
	currentStartDate.setFullYear(props.ui.selectedStartYear.value, props.ui.selectedStartMonth.value, props.ui.selectedStartDay.value);
	currentStartDate.setHours(0, 0, 0, 0);
	if (selectedDate > currentStartDate) {
		selectEndDate(selectedDate);
	} else if (selectedDate < currentStartDate) {
		selectStartDate(selectedDate);
	} else {
		const startDateSelected: PropertyBool = {
			...props.ui.startDateSelected,
			value: false,
		};
		networkStore.invokeSetProp(startDateSelected);
	}
}

function selectStartDate(selectedDate: Date): void {
	networkStore.invokeSetProp({
		...props.ui.selectedStartYear,
		value: selectedDate.getFullYear(),
	});
	networkStore.invokeSetProp({
		...props.ui.selectedStartMonth,
		value: selectedDate.getMonth(),
	});
	networkStore.invokeSetProp({
		...props.ui.selectedStartDay,
		value: selectedDate.getDate(),
	});
	const startDateSelected: PropertyBool = {
		...props.ui.startDateSelected,
		value: true,
	};
	networkStore.invokeSetProp(startDateSelected);
}

function selectEndDate(selectedDate: Date): void {
	networkStore.invokeSetProp({
		...props.ui.selectedEndYear,
		value: selectedDate.getFullYear(),
	});
	networkStore.invokeSetProp({
		...props.ui.selectedEndMonth,
		value: selectedDate.getMonth(),
	});
	networkStore.invokeSetProp({
		...props.ui.selectedEndDay,
		value: selectedDate.getDate(),
	});
	const endDateSelected: PropertyBool = {
		...props.ui.endDateSelected,
		value: true,
	};
	networkStore.invokeSetProp(endDateSelected);
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
					@keydown.enter="datepickerClicked(false)">
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
				@submit="submit"
			/>
		</div>
	</div>
</template>

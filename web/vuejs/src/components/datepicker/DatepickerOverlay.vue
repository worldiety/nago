<!--
 Copyright (c) 2025 worldiety GmbH

 This file is part of the NAGO Low-Code Platform.
 Licensed under the terms specified in the LICENSE file.

 SPDX-License-Identifier: Custom-License
-->

<template>
	<div
		ref="datepicker"
		class="datepicker"
		:class="{ 'double-mode': doubleMode, 'range-mode': rangeMode }"
		@keydown.tab.exact="moveFocusForward"
		@keydown.shift.tab="moveFocusBackwards"
		@keydown.esc="$emit('close')"
	>
		<div class="relative bg-M1 rounded-xl shadow-lg p-6 z-10">
			<div class="h-[22.75rem]">
				<DatepickerHeader ref="datepickerHeader" :label="headerText" @close="emit('close')" />

				<hr class="border-I0" />

				<div class="datepicker-months">
					<!-- Datepicker content -->
					<div v-for="i in monthsToShow" :key="`date_picker_month_${i}`" class="datepicker-content">
						<div class="flex justify-between items-center mb-4 h-8">
							<button
								class="flex justify-center items-center rounded-full size-8"
								:class="{
									'opacity-50': lowerBoundReached && i === 1,
									'hover:bg-I0/15 cursor-pointer': !lowerBoundReached,
									'opacity-0 pointer-events-none': i !== 1,
								}"
								:tabindex="lowerBoundReached ? '-1' : '0'"
								@click="tryDecreaseMonth"
								@keydown.enter="tryDecreaseMonth"
							>
								<ArrowRight class="rotate-180 h-4" />
							</button>
							<div class="flex justify-center items-center text-lg h-full">
								<label
									class="relative hover:bg-I0/15 border-0 bg-M1 rounded-l-md select-none h-full focus-within flex items-center px-1"
								>
									<span class="">
										{{ getSelectedMonthName(i - 1) }}
									</span>
									<select
										:value="getSelectedMonth(i - 1)"
										class="absolute top-0 left-0 size-full opacity-0 overflow-hidden cursor-pointer"
										@change="onMonthChange($event, i - 1)"
									>
										<option
											v-for="(monthEntry, index) of monthNames.entries()"
											:key="index"
											:value="monthEntry[0]"
										>
											{{ monthEntry[1] }}
										</option>
									</select>
								</label>
								<input
									:value="getSelectedYear(i - 1)"
									type="number"
									step="1"
									:min="minYear || 0"
									class="hover:bg-I0/15 border-0 bg-M1 rounded-r-md w-14 h-full arrows-none text-center"
									@keydown.enter="onYearChange($event, i - 1)"
									@change="onYearChange($event, i - 1)"
									@blur="onYearChange($event, i - 1)"
								/>
							</div>
							<button
								class="hover:bg-I0/15 flex justify-center items-center cursor-pointer rounded-full size-8"
								:class="i !== monthsToShow ? 'opacity-0 pointer-events-none' : ''"
								tabindex="0"
								@click="increaseMonth"
								@keydown.enter="increaseMonth"
							>
								<ArrowRight class="h-4" />
							</button>
						</div>

						<div class="datepicker-grid grid grid-cols-7 gap-y-2 text-center leading-none">
							<span>Mo</span>
							<span>Di</span>
							<span>Mi</span>
							<span>Do</span>
							<span>Fr</span>
							<span>Sa</span>
							<span>So</span>

							<div
								v-for="(day, index) in getDatepickerDays(i - 1)"
								:key="index"
								class="relative flex justify-center items-center h-full w-full"
								:class="{
									'within-range-day': day.withinRange,
									'selected-start-day-container': day.selectedStart,
									'selected-end-day-container': day.selectedEnd,
									'other-month-day': day.otherMonth,
								}"
							>
								<div
									:ref="(el) => setLastDatepickerDayElement(el, index)"
									class="day flex justify-center items-center"
									:class="{
										'hover:bg-I0/15 cursor-pointer': day.selectable,
										'unselectable-day': !day.selectable,
										'selected-day': day.selectedStart || day.selectedEnd,
										'text-disabled-text':
											!day.withinRange && day.monthIndex !== getSelectedMonth(i - 1),
									}"
									:tabindex="day.selectable ? '0' : '-1'"
									@click="trySelectDate(day)"
									@keydown.enter="trySelectDate(day)"
								>
									<span class="select-none">{{ day.dayOfMonth }}</span>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>

			<template v-if="rangeMode">
				<hr class="border-ST0" />

				<div class="footer">
					<!-- Reset button when in range mode -->
					<button
						v-if="rangeMode"
						class="button-tertiary"
						:class="{ 'w-full': !doubleMode }"
						:disabled="rangeSelectionState !== RangeSelectionState.COMPLETE"
						@click="clearSelection"
					>
						<undo-icon class="h-4" aria-hidden="true" /> Auswahl zurücksetzen
					</button>

					<!-- Confirm button when in range mode -->
					<button
						ref="confirmButton"
						class="button-confirm button-primary"
						:disabled="rangeSelectionState === RangeSelectionState.SELECT_END"
						@click="emit('submitSelection')"
					>
						{{ t('datepicker.confirm') }}
					</button>
				</div>
			</template>
		</div>

		<!-- Blurred Background -->
		<div class="absolute top-0 left-0 bottom-0 right-0 bg-opacity-60 bg-black z-0" @click="emit('close')"></div>
	</div>
</template>

<script setup lang="ts">
import type { ComponentPublicInstance } from 'vue';
import { computed, ref, useTemplateRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import ArrowRight from '@/assets/svg/arrowRightBold.svg';
import UndoIcon from '@/assets/svg/undo.svg';
import DatepickerHeader from '@/components/datepicker/DatepickerHeader.vue';
import type DatepickerDay from '@/components/datepicker/datepickerDay';
import { RangeSelectionState } from '@/components/datepicker/rangeSelectionState';
import { weekNumber } from 'weeknumber';
import monthNames from '@/shared/monthNames';

const props = defineProps<{
	rangeMode: boolean;
	doubleMode?: boolean;
	label?: string;
	selectedStartDate?: Date;
	selectedEndDate?: Date;
	rangeSelectionState: RangeSelectionState;
	minYear?: number;
}>();

const emit = defineEmits<{
	(e: 'close'): void;
	(e: 'select', day: number, month: number, year: number): void;
	(e: 'submitSelection'): void;
	(e: 'clearSelection'): void;
}>();

const { t } = useI18n();
const datepickerHeader = useTemplateRef('datepickerHeader');
const confirmButton = useTemplateRef('confirmButton');
const datepicker = ref<HTMLElement | undefined>();

const selectedMonth = ref<number>(props.selectedStartDate ? props.selectedStartDate.getMonth() : new Date().getMonth());
const selectedYear = ref<number>(
	props.selectedStartDate ? props.selectedStartDate.getFullYear() : new Date().getFullYear()
);

const lastDatepickerDayIndex = ref<number | null>(null);
const lastDatepickerDayElement = ref<ComponentPublicInstance | Element | null>(null);
const monthsToShow = props.doubleMode ? 2 : 1;

const headerText = computed<string | undefined>(() => {
	if (!props.rangeMode) return props.label;

	switch (props.rangeSelectionState) {
		case RangeSelectionState.SELECT_START:
			return 'Bitte Start auswählen';
		case RangeSelectionState.SELECT_END:
			return 'Bitte Ende auswählen';
		default:
			return props.label;
	}
});

const lowerBoundReached = computed((): boolean => {
	return selectedMonth.value === 0 && selectedYear.value === (props.minYear || 0);
});

function getSelectedMonthName(offset: number): string {
	const monthIndex = getSelectedMonth(offset);
	return monthNames.get(monthIndex) || '';
}

function getSelectedMonth(offset: number): number {
	return (selectedMonth.value + offset) % 12;
}

function getSelectedYear(offset: number): number {
	return selectedYear.value + (selectedMonth.value + offset > 11 ? 1 : 0);
}

function onMonthChange(e: Event, offset: number) {
	const select = e.target as HTMLSelectElement;
	if (!select) return;

	const selectedMonthWithoutOffset = parseInt(select.value, 10) - offset;
	if (selectedMonthWithoutOffset < 0) {
		selectedMonth.value = selectedMonthWithoutOffset + 12;
		selectedYear.value--;
	} else {
		selectedMonth.value = selectedMonthWithoutOffset;
	}
}

function onYearChange(e: Event, offset: number) {
	const input = e.target as HTMLInputElement;
	if (!input || !input.valueAsNumber) return;
	selectedYear.value = Math.max(
		props.minYear || 0,
		input.valueAsNumber - (selectedMonth.value + offset > 11 ? 1 : 0)
	);
}

function setLastDatepickerDayElement(datepickerDay: ComponentPublicInstance | Element | null, index: number) {
	// Update the last element if its index is greater than the index of the current last element or the index of the
	// current last element is null
	if (lastDatepickerDayIndex.value === null || index > lastDatepickerDayIndex.value) {
		lastDatepickerDayIndex.value = index;
		lastDatepickerDayElement.value = datepickerDay;
	}
}

function getDatepickerDays(offset: number): DatepickerDay[] {
	const daysOfCurrentMonth: DatepickerDay[] = [];

	const month = (selectedMonth.value + offset) % 12;
	const year = selectedYear.value + (month > 11 ? 1 : 0);
	const firstDayOfSelectedMonth = new Date(year, month, 1);
	const lastDayOfSelectedMonth = new Date(year, month + 1, 0);

	const dayToShow = new Date(firstDayOfSelectedMonth);
	while (dayToShow.getDay() !== 1) {
		dayToShow.setDate(dayToShow.getDate() - 1);
	}

	while (dayToShow <= lastDayOfSelectedMonth || weekNumber(dayToShow) === weekNumber(lastDayOfSelectedMonth)) {
		const dayOfWeek = dayToShow.getDay() === 0 ? 7 : dayToShow.getDay();
		const dayOfMonth = dayToShow.getDate();

		const selectedStart = isSelectedStartDay(dayOfMonth, dayToShow.getMonth(), dayToShow.getFullYear());

		let selectedEnd = false;
		let withinRange = false;
		if (props.rangeMode) {
			selectedEnd = isSelectedEndDay(dayOfMonth, dayToShow.getMonth(), dayToShow.getFullYear());
			withinRange = isWithinRange(dayOfMonth, dayToShow.getMonth(), dayToShow.getFullYear());
		}

		const selectable = isSelectableDay(dayOfMonth, dayToShow.getMonth(), dayToShow.getFullYear());

		daysOfCurrentMonth.push({
			dayOfWeek: dayOfWeek,
			dayOfMonth: dayOfMonth,
			monthIndex: dayToShow.getMonth(),
			year: dayToShow.getFullYear(),
			selectedStart: selectedStart,
			selectedEnd: selectedEnd,
			withinRange: withinRange,
			selectable: selectable,
			otherMonth: dayToShow.getMonth() !== month,
		});

		dayToShow.setDate(dayToShow.getDate() + 1);
	}

	return daysOfCurrentMonth;
}

function isSelectableDay(day: number, monthIndex: number, year: number): boolean {
	return year >= (props.minYear || 0);
}

function isSelectedStartDay(day: number, monthIndex: number, year: number): boolean {
	return (
		!!props.selectedStartDate &&
		day === props.selectedStartDate.getDate() &&
		monthIndex === props.selectedStartDate.getMonth() &&
		year === props.selectedStartDate.getFullYear()
	);
}

function isSelectedEndDay(day: number, monthIndex: number, year: number): boolean {
	return (
		!!props.selectedEndDate &&
		props.rangeSelectionState !== RangeSelectionState.SELECT_END &&
		day === props.selectedEndDate.getDate() &&
		monthIndex === props.selectedEndDate.getMonth() &&
		year === props.selectedEndDate.getFullYear()
	);
}

function isWithinRange(day: number, monthIndex: number, year: number): boolean {
	if (
		!props.selectedStartDate ||
		!props.selectedEndDate ||
		props.rangeSelectionState === RangeSelectionState.SELECT_END
	) {
		return false;
	}

	const dateToCheck = new Date(year, monthIndex, day, 0, 0, 0, 0);
	return props.selectedStartDate <= dateToCheck && dateToCheck <= props.selectedEndDate;
}

function tryDecreaseMonth(): void {
	if (lowerBoundReached.value) return;

	if (selectedMonth.value === 0) {
		selectedMonth.value = 11;
		selectedYear.value--;
	} else {
		selectedMonth.value--;
	}
}

function increaseMonth(): void {
	if (selectedMonth.value === 11) {
		selectedMonth.value = 0;
		selectedYear.value++;
	} else {
		selectedMonth.value++;
	}
}

function trySelectDate(datepickerDay: DatepickerDay): void {
	if (!datepickerDay.selectable) {
		return;
	}

	emit('select', datepickerDay.dayOfMonth, datepickerDay.monthIndex, datepickerDay.year);
	if (!props.rangeMode) {
		emit('close');
	}
}

function moveFocusForward(event: Event) {
	if (document.activeElement === getLastFocusableElement()) {
		event.preventDefault();
		datepickerHeader.value?.closeButton?.focus();
	}
}

function getLastFocusableElement(): HTMLElement | null {
	if (props.rangeSelectionState === RangeSelectionState.COMPLETE) {
		return confirmButton.value;
	}
	if (lastDatepickerDayElement.value instanceof HTMLElement) {
		return lastDatepickerDayElement.value;
	}
	return null;
}

function moveFocusBackwards(event: Event) {
	if (document.activeElement === datepickerHeader.value?.closeButton) {
		event.preventDefault();
		getLastFocusableElement()?.focus();
	}
}

function clearSelection() {
	emit('clearSelection');
	datepickerHeader.value?.closeButton?.focus();
}

function showSelectedRange() {
	if (!props.selectedStartDate) return;

	for (let i = 0; i < monthsToShow; i++) {
		if (
			props.selectedStartDate.getFullYear() === selectedYear.value &&
			props.selectedStartDate.getMonth() === selectedMonth.value + i
		)
			return;
	}

	selectedMonth.value = props.selectedStartDate.getMonth();
	selectedYear.value = props.selectedStartDate.getFullYear();
}

watch(() => props.selectedStartDate, showSelectedRange);
</script>

<style scoped>
.datepicker {
	@apply fixed top-0 left-0 bottom-0 right-0 flex justify-center items-center z-30;

	hr {
		@apply my-3;
	}

	.datepicker-months {
		@apply grid grid-cols-1 gap-8;

		.datepicker-content {
			@apply max-w-72;
		}
	}

	.footer {
		@apply flex flex-col items-center justify-between w-full gap-2;

		.button-confirm {
			@apply w-full;
		}
	}

	&.double-mode {
		.datepicker-months {
			@apply grid-cols-2;
		}

		.footer {
			@apply flex-row items-center justify-end gap-2;

			.button-confirm {
				@apply w-auto;
			}
		}
	}

	&.range-mode {
		.footer {
			@apply justify-between;
		}
	}
}

.day {
	@apply relative size-8 rounded-full z-20;
}

.selected-day {
	@apply bg-I0 bg-opacity-25 text-M8;
}

.within-range-day {
	@apply bg-I0 bg-opacity-5 text-M8;
}

.unselectable-day {
	@apply text-M8/50;
}

/* Each day in the first column within the selection range except the selected days (after element) */
.datepicker-grid
	> .within-range-day:nth-of-type(7n - 6):not(.selected-start-day-container, .selected-start-day-container)::after {
	content: '';
	@apply absolute top-0 left-0 bottom-0 h-full w-1/2 bg-M1;
}

/* Each day in the first grid column within the selected range that is not a selected day (before element) */
.datepicker-grid > .within-range-day:nth-of-type(7n - 6) > .day:not(.selected-day)::before {
	content: '';
	@apply absolute top-0 left-0 bottom-0 h-full w-1/2 bg-I0 bg-opacity-5 rounded-l-full;
}

/* Each day in the last column within the selected range except the selected days (before element) */
.datepicker-grid
	> .within-range-day:nth-of-type(7n):not(.selected-start-day-container, .selected-end-day-container)::before {
	content: '';
	@apply absolute top-0 bottom-0 right-0 h-full w-1/2 bg-M1;
}

/* Each day in the last grid column within the selected range that is not a selected day (after element) */
.datepicker-grid > .within-range-day:nth-of-type(7n) > .day:not(.selected-day)::after {
	content: '';
	@apply absolute top-0 bottom-0 right-0 h-full w-1/2 bg-I0 bg-opacity-5 rounded-r-full;
}

/* First day of selected range (before element) */
.datepicker-grid > .selected-start-day-container::before {
	content: '';
	width: calc(50% + 1rem);
	@apply absolute top-0 left-0 bottom-0 bg-M1 rounded-r-full h-full;
}

/* Last day of selected range (after element) */
.datepicker-grid > .selected-end-day-container::after {
	content: '';
	width: calc(50% + 1rem);
	@apply absolute top-0 bottom-0 right-0 bg-M1 rounded-l-full h-full;
}

/* Selected start day container in last grid row */
.datepicker-grid > .within-range-day:nth-of-type(7n).selected-start-day-container,
	/* Selected end day container in first grid row */
.datepicker-grid > .within-range-day:nth-of-type(7n - 6).selected-end-day-container {
	@apply bg-transparent;
}

.other-month-day .day span {
	@apply opacity-25;
}
</style>

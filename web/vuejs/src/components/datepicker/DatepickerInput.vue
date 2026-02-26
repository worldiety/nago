<template>
	<InputWrapper :label="label" :error="errorText" :help="supportingText" :disabled="disabled" :no-hover-effect="true">
		<div class="input-field relative z-0 !pr-10">
			<div ref="datepickerInputContainer">
				<!-- Editable start date parts -->
				<template v-if="dateSelected">
					<select
						:value="editableStartDay"
						@input="startDayChanged"
						@blur="trySubmitSelection"
						class="cursor-pointer bg-transparent"
						:aria-label="rangeMode ? 'Startdatum Tag auswählen' : 'Datum Tag auswählen'"
						@click.stop
					>
						<option
							v-for="option in totalDaysForEditableStartMonth"
							:value="formatDateComponent(option)"
							:key="option"
						>
							{{ formatDateComponent(option) }}
						</option>
					</select>
					<span aria-hidden="true">.</span>
					<select
						:value="editableStartMonth"
						@input="startMonthChanged"
						@blur="trySubmitSelection"
						class="cursor-pointer bg-transparent"
						:aria-label="rangeMode ? 'Startdatum Monat auswählen' : 'Datum Monat auswählen'"
						@click.stop
					>
						<option v-for="option in 12" :value="formatDateComponent(option)" :key="option">
							{{ formatDateComponent(option) }}
						</option>
					</select>
					<span aria-hidden="true">.</span>
					<input
						:value="editableStartYear"
						type="text"
						inputmode="numeric"
						class="bg-transparent w-12"
						:aria-label="rangeMode ? 'Startdatum Jahr eingeben' : 'Datum Jahr eingeben'"
						@blur="startYearChanged"
						@click.stop
					/>
				</template>

				<!-- Editable end date parts-->
				<template v-if="dateSelected && rangeMode">
					<span class="mr-2">-</span>
					<select
						:value="editableEndDay"
						@input="endDayChanged"
						@blur="trySubmitSelection"
						class="cursor-pointer bg-transparent"
						aria-label="Enddatum Tag auswählen"
						@click.stop
					>
						<option
							v-for="option in totalDaysForEditableEndMonth"
							:value="formatDateComponent(option)"
							:key="option"
						>
							{{ formatDateComponent(option) }}
						</option>
					</select>
					<span aria-hidden="true">.</span>
					<select
						:value="editableEndMonth"
						@input="endMonthChanged"
						@blur="trySubmitSelection"
						class="cursor-pointer bg-transparent"
						aria-label="Enddatum Monat auswählen"
						@click.stop
					>
						<option v-for="option in 12" :value="formatDateComponent(option)" :key="option">
							{{ formatDateComponent(option) }}
						</option>
					</select>
					<span aria-hidden="true">.</span>
					<input
						:value="editableEndYear"
						type="text"
						inputmode="numeric"
						class="bg-transparent w-12"
						aria-label="Enddatum Jahr eingeben"
						@blur="endYearChanged"
						@click.stop
					/>
				</template>
			</div>

			<!-- Placeholder text -->
			<p v-if="!dateSelected" class="text-placeholder-text">
				{{ $t('datepicker.select') }}
			</p>

			<!-- Clickable calendar icon -->
			<div class="absolute top-0 bottom-0 right-4 flex items-center h-full">
				<div
					class="cursor-pointer hover:text-I0"
					tabindex="0"
					@click="$emit('showDatepicker')"
					@keydown.enter="$emit('showDatepicker')"
					role="button"
					:aria-label="datepickerCalendarAriaLabel"
				>
					<Calendar class="w-4" aria-hidden="true" />
				</div>
			</div>
		</div>
	</InputWrapper>
</template>

<script setup lang="ts">
import { computed, ref, useTemplateRef, watch } from 'vue';
import Calendar from '@/assets/svg/calendar.svg';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { isNumber } from '@tiptap/vue-3';
import type { DateData, DatePickerStyleValues } from '@/shared/proto/nprotoc_gen';
import { UpdateStateValueRequested, UpdateStateValues2Requested } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	// value containing the selected start date
	value?: DateData;
	label?: string;
	errorText?: string;
	supportingText?: string;
	disabled?: boolean;
	datepickerStyle?: DatePickerStyleValues;
	datepickerExpanded: boolean;
	selectedStartDate?: Date;
	selectedEndDate?: Date;
	rangeMode: boolean;
	// needed by Nago to submit updated dates
	inputValue?: number;
	// needed by Nago to submit updated dates
	endInputValue?: number;
}>();

defineEmits<{
	(e: 'showDatepicker'): void;
}>();

const serviceAdapter = useServiceAdapter();
const datepickerInputContainer = useTemplateRef('datepickerInputContainer');
const editableStartYear = ref<number>(0);
const editableStartMonth = ref<string>(formatDateComponent(1));
const editableStartDay = ref<string>(formatDateComponent(1));
const editableEndYear = ref<number>(0);
const editableEndMonth = ref<string>(formatDateComponent(1));
const editableEndDay = ref<string>(formatDateComponent(1));

const totalDaysForEditableStartMonth = computed((): number => {
	const editableStartMonthDate = new Date(
		editableStartYear.value,
		parseInt(editableStartMonth.value, 10),
		0,
		0,
		0,
		0,
		0
	);
	return editableStartMonthDate.getDate();
});

const totalDaysForEditableEndMonth = computed((): number => {
	const editableEndMonthDate = new Date(editableEndYear.value, parseInt(editableEndMonth.value, 10), 0, 0, 0, 0, 0);
	return editableEndMonthDate.getDate();
});

const dateSelected = computed((): boolean => {
	if (!props.value) {
		return false;
	}

	if (!props.value.year) {
		return false;
	}

	return !!props.selectedStartDate && (!props.rangeMode || !!props.selectedEndDate);
});

const datepickerCalendarAriaLabel = computed((): string => {
	if (props.rangeMode) {
		return `Zeitraumauswahldialog öffnen, ${screenreaderDateFormatted.value}`;
	}
	return `Datumsauswahldialog öffnen, ${screenreaderDateFormatted.value}`;
});

const screenreaderDateFormatted = computed((): string => {
	const editableStartDate = new Date(
		editableStartYear.value,
		parseInt(editableStartMonth.value, 10) - 1,
		parseInt(editableStartDay.value, 10),
		0,
		0,
		0,
		0
	);
	if (!props.rangeMode) {
		return `Ausgewähltes Datum: ${editableStartDate.toLocaleDateString()}`;
	}

	const editableEndDate = new Date(
		editableEndYear.value,
		parseInt(editableEndMonth.value, 10) - 1,
		parseInt(editableEndDay.value, 10),
		0,
		0,
		0,
		0
	);
	return `Ausgewählter Zeitraum: ${editableStartDate.toLocaleDateString()} bis ${editableEndDate.toLocaleDateString()}`;
});

function formatDateComponent(dateComponentRaw: number | string): string {
	if (isNumber(dateComponentRaw)) {
		dateComponentRaw = dateComponentRaw.toString(10);
	}
	if (dateComponentRaw.length === 1) {
		return `0${dateComponentRaw}`;
	}
	return dateComponentRaw;
}

function startDayChanged(event: Event) {
	if (!event.target) {
		return;
	}
	editableStartDay.value = (event.target as HTMLInputElement).value;
}

function endDayChanged(event: Event) {
	if (!event.target) {
		return;
	}
	editableEndDay.value = (event.target as HTMLInputElement).value;
}

function startMonthChanged(event: Event) {
	if (!event.target) {
		return;
	}
	editableStartMonth.value = (event.target as HTMLInputElement).value;

	adjustEditableStartDay();
}

function endMonthChanged(event: Event) {
	if (!event.target) {
		return;
	}
	editableEndMonth.value = (event.target as HTMLInputElement).value;

	adjustEditableEndDay();
}

function startYearChanged(event: FocusEvent) {
	if (!event.target) {
		return;
	}
	const newValue = (event.target as HTMLInputElement).value;
	if (!/^[0-9]+$/.test(newValue)) {
		return;
	}
	const updatedValue = parseInt(newValue, 10);
	if (updatedValue <= 1582) {
		// only support years after introduction of the gregorian calendar
		// also we have to set the value to 0 here first, otherwise the reset to the previous value will not work
		editableStartYear.value = 0;
		editableStartYear.value = props.selectedStartDate?.getFullYear() || new Date().getFullYear();
		return;
	}
	editableStartYear.value = updatedValue;

	adjustEditableStartDay();

	trySubmitSelection(event);
}

function endYearChanged(event: FocusEvent) {
	if (!event.target) {
		return;
	}
	const newValue = (event.target as HTMLInputElement).value;
	if (!/^[0-9]+$/.test(newValue)) {
		return;
	}
	const updatedValue = parseInt(newValue, 10);
	if (updatedValue <= 1582) {
		// only support years after introduction of the gregorian calendar
		// also we have to set the value to 0 here first, otherwise the reset to the previous value will not work
		editableEndYear.value = 0;
		editableEndYear.value = props.selectedEndDate?.getFullYear() || new Date().getFullYear();
		return;
	}
	editableEndYear.value = updatedValue;

	adjustEditableEndDay();

	trySubmitSelection(event);
}

function adjustEditableStartDay() {
	if (parseInt(editableStartDay.value, 10) > totalDaysForEditableStartMonth.value) {
		// current start day is greater than the amount of days in the current month
		// so we have to adjust the start day to this amount
		editableStartDay.value = formatDateComponent(totalDaysForEditableStartMonth.value);
	}
}

function adjustEditableEndDay() {
	if (parseInt(editableEndDay.value, 10) > totalDaysForEditableEndMonth.value) {
		// current end day is greater than the amount of days in the current month
		// so we have to adjust the end day to this amount
		editableEndDay.value = formatDateComponent(totalDaysForEditableEndMonth.value);
	}
}

function isEditableEndDateBeforeEditableStartDate() {
	if (!props.rangeMode) {
		return false;
	}

	const editableStartDate = new Date(
		editableStartYear.value,
		parseInt(editableStartMonth.value, 10),
		parseInt(editableStartDay.value, 10),
		0,
		0,
		0,
		0
	);
	const editableEndDate = new Date(
		editableEndYear.value,
		parseInt(editableEndMonth.value, 10),
		parseInt(editableEndDay.value, 10),
		0,
		0,
		0,
		0
	);
	return editableEndDate < editableStartDate;
}

function swapEditableDates() {
	const tempStartDay = editableStartDay.value;
	const tempStartMonth = editableStartMonth.value;
	const tempStartYear = editableStartYear.value;

	editableStartDay.value = editableEndDay.value;
	editableStartMonth.value = editableEndMonth.value;
	editableStartYear.value = editableEndYear.value;

	editableEndDay.value = tempStartDay;
	editableEndMonth.value = tempStartMonth;
	editableEndYear.value = tempStartYear;
}

function trySubmitSelection(event: FocusEvent) {
	if (!event.relatedTarget) {
		return;
	}

	const relatedNode = event.relatedTarget as Node;
	if (!datepickerInputContainer.value?.contains(relatedNode)) {
		// Focus left datepicker input container so update the selection
		submitSelection();
	}
}

function submitSelection() {
	if (isEditableEndDateBeforeEditableStartDate()) {
		swapEditableDates();
	}

	const updatedStartDay = parseInt(editableStartDay.value, 10);
	const updatedStartMonth = parseInt(editableStartMonth.value, 10);
	const updatedEndDay = parseInt(editableEndDay.value, 10);
	const updatedEndMonth = parseInt(editableEndMonth.value, 10);

	if (props.rangeMode) {
		serviceAdapter.sendEvent(
			new UpdateStateValues2Requested(
				props.inputValue,
				JSON.stringify({
					d: updatedStartDay,
					m: updatedStartMonth,
					y: editableStartYear.value,
				}),
				props.endInputValue,
				JSON.stringify({
					d: updatedEndDay,
					m: updatedEndMonth,
					y: editableEndYear.value,
				}),
				0,
				nextRID()
			)
		);
	} else {
		serviceAdapter.sendEvent(
			new UpdateStateValueRequested(
				props.inputValue,
				0,
				nextRID(),
				JSON.stringify({
					d: updatedStartDay,
					m: updatedStartMonth,
					y: editableStartYear.value,
				})
			)
		);
	}
}

function setInitialStartDate() {
	const now = new Date();
	now.setHours(0, 0, 0, 0);

	if (props.selectedStartDate) {
		editableStartYear.value = props.selectedStartDate.getFullYear();
		editableStartMonth.value = formatDateComponent(props.selectedStartDate.getMonth() + 1);
		editableStartDay.value = formatDateComponent(props.selectedStartDate.getDate());
	} else {
		editableStartYear.value = now.getFullYear();
		editableStartMonth.value = formatDateComponent(now.getMonth() + 1);
		editableStartDay.value = formatDateComponent(now.getDate());
	}
}

function setInitialEndDate() {
	const now = new Date();
	now.setHours(0, 0, 0, 0);

	if (props.selectedEndDate) {
		editableEndYear.value = props.selectedEndDate.getFullYear();
		editableEndMonth.value = formatDateComponent(props.selectedEndDate.getMonth() + 1);
		editableEndDay.value = formatDateComponent(props.selectedEndDate.getDate());
	} else {
		editableEndYear.value = now.getFullYear();
		editableEndMonth.value = formatDateComponent(now.getMonth() + 1);
		editableEndDay.value = formatDateComponent(now.getDate());
	}
}

function init() {
	setInitialStartDate();
	setInitialEndDate();

	watch(() => props.selectedStartDate, setInitialStartDate);
	watch(() => props.selectedEndDate, setInitialEndDate);
}

init();
</script>

<template>
	<InputWrapper :label="label" :error="errorText" :help="supportingText" :disabled="disabled" :no-hover-effect="true">
		<div class="input-field relative z-0 !pr-10">
			<div ref="datepickerInputContainer">
				<!-- Editable start date parts -->
				<template v-if="dateSelected">
					<input
						v-model="editableStartDay"
						type="number"
						step="1"
						min="1"
						:max="totalDaysForEditableStartMonth"
						class="input-day"
						:aria-label="rangeMode ? 'Startdatum Tag auswählen' : 'Datum Tag auswählen'"
						onfocus="this.select()"
						@input="onInputInput"
						@blur="onInputBlur"
						@click.stop
					/>
					<span aria-hidden="true">.</span>
					<input
						v-model="editableStartMonth"
						type="number"
						step="1"
						min="1"
						max="12"
						class="input-month"
						:aria-label="rangeMode ? 'Startdatum Monat auswählen' : 'Datum Monat auswählen'"
						onfocus="this.select()"
						@input="onInputInput"
						@blur="onInputBlur"
						@click.stop
					/>
					<span aria-hidden="true">.</span>
					<input
						v-model="editableStartYear"
						type="number"
						step="1"
						:min="minYear || 0"
						max="9999"
						class="input-year"
						:aria-label="rangeMode ? 'Startdatum Jahr eingeben' : 'Datum Jahr eingeben'"
						onfocus="this.select()"
						@input="onInputInput"
						@blur="onInputBlur"
						@click.stop
					/>
				</template>

				<!-- Editable end date parts-->
				<template v-if="dateSelected && rangeMode">
					<span class="mr-2">-</span>
					<input
						v-model="editableEndDay"
						type="number"
						step="1"
						min="1"
						:max="totalDaysForEditableEndMonth"
						class="input-day"
						aria-label="Enddatum Tag auswählen"
						onfocus="this.select()"
						@input="onInputInput"
						@blur="onInputBlur"
						@click.stop
					/>
					<span aria-hidden="true">.</span>
					<input
						v-model="editableEndMonth"
						type="number"
						step="1"
						min="1"
						max="12"
						class="input-month"
						aria-label="Enddatum Monat auswählen"
						onfocus="this.select()"
						@input="onInputInput"
						@blur="onInputBlur"
						@click.stop
					/>
					<span aria-hidden="true">.</span>
					<input
						v-model="editableEndYear"
						type="number"
						step="1"
						:min="minYear || 0"
						max="9999"
						class="input-year"
						aria-label="Enddatum Jahr eingeben"
						onfocus="this.select()"
						@input="onInputInput"
						@blur="onInputBlur"
						@click.stop
					/>
				</template>
			</div>

			<!-- Placeholder text -->
			<p v-if="!dateSelected" class="text-placeholder-text">
				{{ $t('datepicker.select') }}
			</p>

			<!-- Clickable calendar icon -->
			<button
				class="button-tertiary square small additional-right overlay-button"
				tabindex="0"
				role="button"
				:aria-label="datepickerCalendarAriaLabel"
				@click="$emit('showDatepicker')"
				@keydown.enter="$emit('showDatepicker')"
			>
				<Calendar class="w-4" aria-hidden="true" />
			</button>
		</div>
	</InputWrapper>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUpdated, ref, useTemplateRef, watch } from 'vue';
import Calendar from '@/assets/svg/calendar.svg';
import InputWrapper from '@/components/shared/InputWrapper.vue';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
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
	minYear?: number;
}>();

defineEmits<{
	(e: 'showDatepicker'): void;
}>();

const serviceAdapter = useServiceAdapter();
const datepickerInputContainer = useTemplateRef('datepickerInputContainer');
const editableStartYear = ref<number>(0);
const editableStartMonth = ref<number>(1);
const editableStartDay = ref<number>(1);
const editableEndYear = ref<number>(0);
const editableEndMonth = ref<number>(1);
const editableEndDay = ref<number>(1);

const totalDaysForEditableStartMonth = computed((): number => {
	const editableStartMonthDate = new Date(editableStartYear.value, editableStartMonth.value, 0, 0, 0, 0, 0);
	return editableStartMonthDate.getDate();
});

const totalDaysForEditableEndMonth = computed((): number => {
	const editableEndMonthDate = new Date(editableEndYear.value, editableEndMonth.value, 0, 0, 0, 0, 0);
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
		editableStartMonth.value - 1,
		editableStartDay.value,
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
		editableEndMonth.value - 1,
		editableEndDay.value,
		0,
		0,
		0,
		0
	);
	return `Ausgewählter Zeitraum: ${editableStartDate.toLocaleDateString()} bis ${editableEndDate.toLocaleDateString()}`;
});

function isEditableEndDateBeforeEditableStartDate() {
	if (!props.rangeMode) {
		return false;
	}

	const editableStartDate = new Date(
		editableStartYear.value,
		editableStartMonth.value,
		editableStartDay.value,
		0,
		0,
		0,
		0
	);
	const editableEndDate = new Date(editableEndYear.value, editableEndMonth.value, editableEndDay.value, 0, 0, 0, 0);
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

function onInputInput(event: Event) {
	const input = event.target as HTMLInputElement;
	if (!input) return;

	fixInputRange(input);
	fixAllInputDecimals();
}

function onInputBlur(event: FocusEvent) {
	const input = event.target as HTMLInputElement;
	if (!input) return;

	fixInputDecimals(input);
	trySubmitSelection(event);
}

function fixInputRange(input: HTMLInputElement) {
	const min = input.min;
	if (min && input.valueAsNumber < parseInt(min)) {
		input.value = min;
	}

	const max = input.max;
	if (max && input.valueAsNumber > parseInt(max)) {
		input.value = max;
	}
}

function fixInputDecimals(input: HTMLInputElement) {
	if (input.classList.contains('input-day') || input.classList.contains('input-month')) {
		if (isNaN(input.valueAsNumber)) input.valueAsNumber = 1;
		const value = input.value;
		if (value.length === 1) {
			input.value = `0${value}`;
		}
	}
	if (input.classList.contains('input.year')) {
		if (isNaN(input.valueAsNumber)) input.valueAsNumber = new Date().getFullYear();
		const value = input.value;
		while (value.length < 4) {
			input.value = `0${value}`;
		}
	}
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

	const updatedStartDay = editableStartDay.value;
	const updatedStartMonth = editableStartMonth.value;
	const updatedEndDay = editableEndDay.value;
	const updatedEndMonth = editableEndMonth.value;

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
		editableStartMonth.value = props.selectedStartDate.getMonth() + 1;
		editableStartDay.value = props.selectedStartDate.getDate();
	} else {
		editableStartYear.value = now.getFullYear();
		editableStartMonth.value = now.getMonth() + 1;
		editableStartDay.value = now.getDate();
	}
}

function setInitialEndDate() {
	const now = new Date();
	now.setHours(0, 0, 0, 0);

	if (props.selectedEndDate) {
		editableEndYear.value = props.selectedEndDate.getFullYear();
		editableEndMonth.value = props.selectedEndDate.getMonth() + 1;
		editableEndDay.value = props.selectedEndDate.getDate();
	} else {
		editableEndYear.value = now.getFullYear();
		editableEndMonth.value = now.getMonth() + 1;
		editableEndDay.value = now.getDate();
	}
}

function fixAllInputDecimals() {
	if (!datepickerInputContainer.value) return;
	const inputs = datepickerInputContainer.value.querySelectorAll('input');
	inputs.forEach((input) => {
		if (document.activeElement !== input) {
			fixInputDecimals(input as HTMLInputElement);
		}
	});
}

function init() {
	setInitialStartDate();
	setInitialEndDate();

	watch(() => props.selectedStartDate, setInitialStartDate);
	watch(() => props.selectedEndDate, setInitialEndDate);
}

init();
onMounted(fixAllInputDecimals);
onUpdated(fixAllInputDecimals);
watch(dateSelected, () => nextTick(fixAllInputDecimals));
</script>
<style scoped>
.input-day,
.input-month,
.input-year {
	@apply appearance-none bg-transparent text-center;
	@apply focus:outline-offset-2;
}

.additional-right {
	@apply absolute top-1/2 right-1.5 -translate-y-1/2;
}

.overlay-button {
	@apply size-8 p-1;
}
</style>

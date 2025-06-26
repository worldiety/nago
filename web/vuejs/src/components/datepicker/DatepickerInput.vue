<template>
	<InputWrapper
		:label="label"
		:error="errorText"
		:hint="supportingText"
		:disabled="disabled"
	>
		<div class="input-field relative z-0 !pr-10">
			<!-- Editable date parts -->
			<p v-if="dateSelected">
				<input v-model.trim="editableStartDay" type="text" inputmode="numeric" />
				<input v-model.trim="editableStartMonth" type="text" inputmode="numeric" />
				<input v-model.trim="editableStartYear" type="text" inputmode="numeric" />
			</p>

			<!-- Placeholder text -->
			<p v-else class="text-placeholder-text">
				{{ $t('datepicker.select') }}
			</p>

			<!-- Clickable calendar icon -->
			<div class="absolute top-0 bottom-0 right-4 flex items-center pointer-events-none h-full">
				<div
					tabindex="0"
					@click="$emit('showDatepicker')"
					@keydown.enter="$emit('showDatepicker')"
				>
					<Calendar class="w-4" />
				</div>
			</div>
		</div>
	</InputWrapper>
</template>

<script setup lang="ts">
import InputWrapper from '@/components/shared/InputWrapper.vue';
import Calendar from '@/assets/svg/calendar.svg';
import { computed, ref, useTemplateRef, watch } from 'vue';
import { DateData, DatePickerStyleValues } from '@/shared/proto/nprotoc_gen';
import { isNumber } from '@tiptap/vue-3';

const props = defineProps<{
	value?: DateData;
	label?: string;
	errorText?: string;
	supportingText?: string;
	disabled?: boolean;
	datepickerStyle?: DatePickerStyleValues;
	datepickerExpanded: boolean;
	selectedStartYear: number;
	selectedStartMonth: number;
	selectedStartDay: number;
	selectedEndYear: number;
	selectedEndMonth: number;
	selectedEndDay: number;
	rangeMode: boolean;
}>();

defineEmits<{
	(e: 'showDatepicker'): void;
}>();

const editableStartYear = ref<string>(formatDateComponent(props.selectedStartYear));
const editableStartMonth = ref<string>(formatDateComponent(props.selectedStartMonth));
const editableStartDay = ref<string>(formatDateComponent(props.selectedStartDay));

watch(() => props.selectedStartYear, (newValue) => {
	editableStartYear.value = formatDateComponent(newValue);
});

watch(() => props.selectedStartMonth, (newValue) => {
	editableStartMonth.value = formatDateComponent(newValue);
});

watch(() => props.selectedStartDay, (newValue) => {
	editableStartDay.value = formatDateComponent(newValue);
});

watch(editableStartDay, (newValue) => {
	editableStartDay.value = formatDateComponent(newValue);
});

function formatDateComponent(dateComponentRaw: number|string): string {
	if (isNumber(dateComponentRaw)) {
		dateComponentRaw = dateComponentRaw.toString(10);
	}
	if (dateComponentRaw.length === 1) {
		return `0${dateComponentRaw}`;
	}
	return dateComponentRaw;
}

const dateSelected = computed((): boolean => {
	if (!props.value) {
		return false;
	}

	if (!props.value.year) {
		return false;
	}

	const startDate = new Date();
	startDate.setFullYear(props.selectedStartYear, props.selectedStartMonth - 1, props.selectedStartDay);
	if (!props.selectedStartYear || !(props.selectedStartMonth - 1) || !props.selectedStartDay) {
		return false;
	}
	if (!props.rangeMode) {
		return true;
	}

	return !!(props.selectedEndYear && props.selectedEndMonth - 1 && props.selectedEndDay);
});
</script>

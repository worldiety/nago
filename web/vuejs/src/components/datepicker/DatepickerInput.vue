<template>
	<InputWrapper
		:label="label"
		:error="errorText"
		:hint="supportingText"
		:disabled="disabled"
	>
		<div
			class="input-field relative z-0 !pr-10"
			tabindex="0"
			@click="$emit('showDatepicker')"
			@keydown.enter="$emit('showDatepicker')"
		>
			<p :class="{ 'text-placeholder-text': !dateFormatted }">
				{{ dateFormatted ?? $t('datepicker.select') }}
			</p>
			<div class="absolute top-0 bottom-0 right-4 flex items-center pointer-events-none h-full">
				<Calendar class="w-4" />
			</div>
		</div>
	</InputWrapper>
</template>

<script setup lang="ts">
import InputWrapper from '@/components/shared/InputWrapper.vue';
import Calendar from '@/assets/svg/calendar.svg';
import { computed } from 'vue';
import { DateData, DatePickerStyleValues } from '@/shared/proto/nprotoc_gen';

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
}>();

defineEmits<{
	(e: 'showDatepicker'): void;
}>();

const dateFormatted = computed((): string | null => {
	if (!props.value) {
		return null;
	}

	if (!props.value.year) {
		return null;
	}

	const startDate = new Date();
	startDate.setFullYear(props.selectedStartYear, props.selectedStartMonth - 1, props.selectedStartDay);
	if (props.datepickerStyle !== DatePickerStyleValues.DatePickerDateRange) {
		//console.log("bugs!!",startDate.toLocaleDateString())
		return startDate.toLocaleDateString();
	}
	const endDate = new Date();
	endDate.setFullYear(props.selectedEndYear, props.selectedEndMonth - 1, props.selectedEndDay);
	return `${startDate.toLocaleDateString()} - ${endDate.toLocaleDateString()}`;
});
</script>

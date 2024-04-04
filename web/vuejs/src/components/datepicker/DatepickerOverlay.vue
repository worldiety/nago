<template>
	<div v-if="expanded" ref="datepicker" class="fixed top-0 left-0 bottom-0 right-0 flex justify-center items-center text-black dark:text-white z-30">
		<div class="relative bg-white dark:bg-gray-700 rounded-xl shadow-lg max-w-96 h-[25rem] p-6 z-10">
			<DatepickerHeader :label="label" @close="emit('close')" class="mb-4" />

			<!-- Datepicker content -->
			<div class="flex justify-between items-center mb-4 h-8">
				<div
					class="effect-hover flex justify-center items-center rounded-full size-8"
					tabindex="0"
					@click="decreaseMonth"
					@keydown.enter="decreaseMonth"
				>
					<ArrowRight class="rotate-180 h-4" />
				</div>
				<div class="flex justify-center items-center basis-2/3 gap-x-px text-lg h-full">
					<div class="basis-1/2 shrink-0 grow-0 h-full">
						<select v-model="currentMonthIndex" class="effect-hover border-0 bg-white dark:bg-darkmode-gray text-right cursor-default rounded-l-md w-full h-full px-2">
							<option v-for="(monthEntry, index) of monthNames.entries()" :key="index" :value="monthEntry[0]">
								{{ monthEntry[1] }}
							</option>
						</select>
					</div>
					<div class="basis-1/2 shrink-0 grow-0 h-full">
						<input v-model="yearInput" type="text" class="effect-hover border-0 bg-white dark:bg-darkmode-gray rounded-r-md text-left w-full h-full px-2">
					</div>
				</div>
				<div
					class="effect-hover flex justify-center items-center rounded-full size-8"
					tabindex="0"
					@click="increaseMonth"
					@keydown.enter="increaseMonth"
				>
					<ArrowRight class="h-4" />
				</div>
			</div>

			<div class="grid grid-cols-7 gap-2 text-center leading-none">
				<span>Mo</span>
				<span>Di</span>
				<span>Mi</span>
				<span>Do</span>
				<span>Fr</span>
				<span>Sa</span>
				<span>So</span>

				<div v-for="(fillingDay, index) in fillingDaysOfPreviousMonth" :key="index">
					<div class="flex justify-center items-center h-full w-full">
						<span class="text-disabled-text">{{ fillingDay }}</span>
					</div>
				</div>
				<div v-for="(day, index) in totalDaysInMonth" :key="index" class="flex justify-center items-center h-full w-full">
					<div
						class="day effect-hover flex justify-center items-center cursor-default"
						:class="{'selected-day': isSelectedDay(day)}"
						tabindex="0"
						@click="selectDay(day)"
						@keydown.enter="selectDay(day)"
					>
						<span>{{ day }}</span>
					</div>
				</div>
			</div>
		</div>

		<!-- Blurred Background -->
		<div class="absolute top-0 left-0 bottom-0 right-0 backdrop-blur z-0" @click="emit('close')"></div>
	</div>
</template>

<script setup lang="ts">
import monthNames from '@/shared/monthNames'
import ArrowRight from '@/assets/svg/arrowRightBold.svg';
import { computed, ref, watch } from 'vue';
import DatepickerHeader from '@/components/datepicker/DatepickerHeader.vue';

const props = defineProps<{
	expanded: boolean;
	label: string;
	selectedDay: number;
	selectedMonth: number;
	selectedYear: number;
}>();

const emit = defineEmits<{
	(e: 'close'): void;
	(e: 'select', day: number, month: number, year: number): void;
}>();

const datepicker = ref<HTMLElement|undefined>();
const currentDate = new Date(Date.now());
const currentYear = ref<number>(currentDate.getFullYear());
const currentMonthIndex = ref<number>(currentDate.getMonth());
const yearInput = ref<string>(currentYear.value.toString(10));

/**
 * Only allow year values with a length between 1 and 4.
 * Does also prevent values less than 1 and greater than 9999.
 */
watch(yearInput, (newValue, oldValue) => {
	if (newValue.match(/^[1-9][0-9]{0,3}$/)) {
		currentYear.value = parseInt(newValue, 10);
	} else {
		yearInput.value = oldValue;
	}
});

const totalDaysInMonth = computed((): number => {
	const lastDayOfMonthDate = new Date();
	lastDayOfMonthDate.setFullYear(currentYear.value, currentMonthIndex.value + 1, 0);
	return lastDayOfMonthDate.getDate();
});

const fillingDaysOfPreviousMonth = computed((): number[] => {
	const lastDayOfPreviousMonthDate = new Date();
	lastDayOfPreviousMonthDate.setFullYear(currentYear.value, currentMonthIndex.value, 0);
	const lastDayOfPreviousMonth = lastDayOfPreviousMonthDate.getDate();
	const fillingDays: number[] = [];
	for (let i = 0; i < dayStartOffsetInMonth.value; i++) {
		fillingDays.unshift(lastDayOfPreviousMonth - i);
	}
	return fillingDays;
});

const dayStartOffsetInMonth = computed((): number => {
	const firstDayOfMonthDate = new Date();
	firstDayOfMonthDate.setFullYear(currentYear.value, currentMonthIndex.value, 1);
	return firstDayOfMonthDate.getDay() === 0 ? 6 : firstDayOfMonthDate.getDay() - 1;
});

function selectDay(day: number): void {
	emit('select', day, currentMonthIndex.value + 1, currentYear.value);
}

function isSelectedDay(day: number): boolean {
	return day === props.selectedDay
		&& currentMonthIndex.value === props.selectedMonth - 1
		&& currentYear.value === props.selectedYear;
}

function decreaseMonth(): void {
	if (currentMonthIndex.value === 0) {
		currentMonthIndex.value = 11;
		currentYear.value -= 1;
		yearInput.value = currentYear.value.toString(10);
		return;
	}
	currentMonthIndex.value -= 1;
}

function increaseMonth(): void {
	if (currentMonthIndex.value === 11) {
		currentMonthIndex.value = 0;
		currentYear.value += 1;
		yearInput.value = currentYear.value.toString(10);
		return;
	}
	currentMonthIndex.value += 1;
}
</script>

<style scoped>
.day {
	@apply size-8 rounded-full
}

.selected-day {
	@apply bg-ora-orange bg-opacity-25 text-ora-orange;
}
</style>

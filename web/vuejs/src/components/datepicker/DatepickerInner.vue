<template>
	<div class="flex justify-between items-center mb-4 h-8">
		<button
			class="flex justify-center items-center rounded-full size-8"
			:class="lowerBoundReached ? 'opacity-50' : 'hover:bg-I0/15 cursor-pointer'"
			:tabindex="lowerBoundReached ? '-1' : '0'"
			@click="tryDecreaseMonth"
			@keydown.enter="tryDecreaseMonth"
		>
			<ArrowRight class="rotate-180 h-4" />
		</button>
		<div class="flex justify-center items-center basis-2/3 gap-x-px text-lg h-full">
			<div class="basis-1/2 shrink-0 grow-0 h-full">
				<select
					v-model="currentMonthIndexModel"
					class="hover:bg-I0/15 border-0 bg-M1 text-right cursor-pointer rounded-l-md select-none w-full h-full px-2"
				>
					<option v-for="(monthEntry, index) of monthNames.entries()" :key="index" :value="monthEntry[0]">
						{{ monthEntry[1] }}
					</option>
				</select>
			</div>
			<div class="basis-1/2 shrink-0 grow-0 h-full">
				<input
					v-model="yearInputModel"
					type="text"
					class="hover:bg-I0/15 border-0 bg-M1 rounded-r-md text-left w-full h-full px-2"
					@blur="trySubmitYearInput"
				/>
			</div>
		</div>
		<button
			class="hover:bg-I0/15 flex justify-center items-center cursor-pointer rounded-full size-8"
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
			v-for="(datepickerDay, index) in datepickerDays"
			:key="index"
			class="relative flex justify-center items-center h-full w-full"
			:class="{
				'within-range-day': datepickerDay.withinRange,
				'selected-start-day-container': datepickerDay.selectedStart,
				'selected-end-day-container': datepickerDay.selectedEnd,
			}"
		>
			<div
				:ref="(el) => setLastDatepickerDayElement(el, index)"
				class="day flex justify-center items-center"
				:class="{
					'hover:bg-I0/15 cursor-pointer': datepickerDay.selectable,
					'unselectable-day': !datepickerDay.selectable,
					'selected-day': datepickerDay.selectedStart || datepickerDay.selectedEnd,
					'text-disabled-text':
						!datepickerDay.withinRange && datepickerDay.monthIndex !== currentMonthIndexModel,
				}"
				:tabindex="datepickerDay.selectable ? '0' : '-1'"
				@click="trySelectDate(datepickerDay)"
				@keydown.enter="trySelectDate(datepickerDay)"
			>
				<span class="select-none">{{ datepickerDay.dayOfMonth }}</span>
			</div>
		</div>
	</div>
</template>
<script setup lang="ts">
import ArrowRight from '@/assets/svg/arrowRightBold.svg';
</script>

<template>
	<div class="flex flex-col-reverse">

		<!-- Input -->
		<div class="peer relative">
			<div class="peer input-field-wrapper" :class="inputFieldWrapperClasses">
				<slot />
			</div>
		</div>

		<!-- Label with optional hint -->
		<div class="flex justify-between items-end text-sm mb-1" :class="{'peer-focus-within:font-semibold': !disabled}">
			<div v-if="label" class="flex justify-start items-center gap-x-1" :class="labelClass">
				<LockIcon v-if="disabled" class="h-4" />
				<ErrorIcon v-else-if="error" class="h-2.5" />
				<span>{{ label }}</span>
			</div>
			<div v-if="!disabled && (error || hint)" class="font-normal">
				<span v-if="error" class="text-error">{{ t('inputWrapper.error') }}</span>
				<span v-else-if="hint" class="text-disabled-text">{{ hint }}</span>
			</div>
		</div>
	</div>

	<!-- Error message -->
	<div v-if="!disabled && (error || help)" class="mt-1 text-sm">
		<p v-if="error" class="text-error">{{ error }}</p>
		<p v-else-if="help" class="text-disabled-text">{{ help }}</p>
	</div>
</template>

<script setup lang="ts">
import LockIcon from '@/assets/svg/lock.svg';
import ErrorIcon from '@/assets/svg/closeBold.svg';
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps<{
	simple?: boolean;
	label?: string;
	error?: string;
	hint?: string;
	help?: string;
	disabled?: boolean;
}>();

const { t } = useI18n();

const labelClass = computed((): string => {
	if (props.disabled) {
		return 'text-disabled-text';
	}
	if (props.error) {
		return 'text-error';
	}
	return 'dark:text-white';
});

const inputFieldWrapperClasses = computed((): string|null => {
	const classes: string[] = [];
	if (!props.simple) {
		classes.push('input-field-wrapper-detailed');
	}
	if (props.disabled) {
		classes.push('input-field-wrapper-disabled');
	} else if (props.error) {
		classes.push('input-field-wrapper-error');
	}
	return classes.join(' ') || null;
});
</script>

<style>
.input-field-wrapper .input-field {
  @apply relative bg-transparent border-0 border-b border-b-black text-black cursor-default w-full px-0 py-2;
  @apply dark:border-b-white dark:text-white;
}

.input-field-wrapper.input-field-wrapper-detailed .input-field {
	@apply border border-black rounded-md px-4;
  @apply dark:border-white;
}

.input-field-wrapper.input-field-wrapper-error .input-field {
	@apply border-error;
  @apply dark:border-error;
}

.input-field-wrapper input::placeholder {
  @apply text-disabled-text;
}

.input-field-wrapper:hover .input-field,
.input-field-wrapper .input-field:focus {
  @apply border-ora-orange border-opacity-75;
  @apply dark:border-ora-orange;
}

.input-field-wrapper .input-field:focus {
  @apply outline-none ring-0;
}

.input-field-wrapper.input-field-wrapper-disabled .input-field {
  @apply border-b-disabled-text text-disabled-text;
}

.input-field-wrapper.input-field-wrapper-detailed.input-field-wrapper-disabled .input-field,
.input-field-wrapper.input-field-wrapper-detailed.input-field-wrapper-disabled .input-field::placeholder {
  @apply bg-disabled-background border-none;
  @apply dark:bg-disabled-text dark:text-disabled-background;
}

.input-field-wrapper {
	@apply text-black;
	@apply dark:text-white;
}

.input-field-wrapper.input-field-wrapper-disabled {
	@apply text-disabled-text pointer-events-none;
}
</style>

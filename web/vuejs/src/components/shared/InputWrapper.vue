<script setup lang="ts">
import LockIcon from '@/assets/svg/lock.svg';
import ErrorIcon from '@/assets/svg/closeBold.svg';
import { computed } from 'vue';

const props = defineProps<{
	simple?: boolean;
	label?: string;
	error?: string;
	hint?: string;
	disabled?: boolean;
}>();

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

const inputFieldOutlineClass = computed((): string|null => {
  if (!props.simple) {
    return 'input-field-outline-detailed';
  }
  return null;
});
</script>

<template>
  <div class="flex flex-col-reverse">

    <!-- Input -->
    <div class="peer relative">
      <div class="peer input-field-wrapper" :class="inputFieldWrapperClasses">
        <slot />
      </div>
      <div class="input-field-outline" :class="inputFieldOutlineClass"></div>
    </div>

    <!-- Label with optional hint -->
    <div class="flex justify-between items-end text-sm mb-1" :class="{'peer-focus-within:font-semibold': !disabled}">
      <div v-if="label" class="flex justify-start items-center gap-x-1" :class="labelClass">
        <LockIcon v-if="disabled" class="h-4" />
        <ErrorIcon v-else-if="error" class="h-2.5" />
        <span>{{ label }}</span>
      </div>
      <span v-if="hint" class="text-disabled-text font-normal">{{ hint }}</span>
    </div>
  </div>

  <!-- Error message -->
	<p v-if="!disabled && error" class="mt-1 text-sm text-end text-error">{{ error }}</p>
</template>

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

.input-field-wrapper:hover .input-field {
  @apply border-ora-orange border-opacity-75;
  @apply dark:border-ora-orange;
}

.input-field-wrapper .input-field:active {
  @apply border-opacity-65;
}

.input-field-wrapper .input-field:focus {
  @apply outline-none ring-0 border-transparent;
  @apply dark:border-transparent;
}

.input-field-wrapper.input-field-wrapper-disabled .input-field {
  @apply border-b-disabled-text text-disabled-text;
}

.input-field-wrapper.input-field-wrapper-detailed.input-field-wrapper-disabled .input-field,
.input-field-wrapper.input-field-wrapper-detailed.input-field-wrapper-disabled .input-field::placeholder {
  @apply bg-disabled-background border-none;
  @apply dark:bg-disabled-text dark:text-disabled-background;
}

.input-field-outline {
  @apply absolute top-0 -left-2 bottom-0 -right-2 rounded-sm pointer-events-none;
  @apply peer-focus-within:outline-none peer-focus-within:outline-2 peer-focus-within:outline-black peer-focus-within:ring-2 peer-focus-within:ring-white;
}

.input-field-outline.input-field-outline-detailed {
  @apply right-0 left-0;
}

.input-field-wrapper {
	@apply text-black;
	@apply dark:text-white;
}

.input-field-wrapper.input-field-wrapper-disabled {
	@apply text-disabled-text pointer-events-none;
}
</style>

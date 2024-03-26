<script setup lang="ts">

</script>

<template>
  <div class="relative">
    <div class="peer input-field-wrapper">
      <slot />
    </div>
    <div class="input-field-outline"></div>
  </div>
</template>

<style scoped>
.input-field-wrapper:deep(*) {
  @apply relative bg-transparent border-0 border-b border-b-black text-black cursor-default w-full px-0;
  @apply dark:border-b-white dark:text-white;
}

.input-field-wrapper:deep(*::placeholder) {
  @apply text-disabled-text;
}

.input-field-wrapper:deep(*:hover) {
  @apply border-b-ora-orange border-opacity-75;
}

.input-field-wrapper:deep(*:active) {
  @apply border-opacity-65;
}

.input-field-wrapper:deep(*:focus-visible) {
  @apply outline-none ring-0 border-b-transparent;
}

.input-field-wrapper:deep(*:disabled) {
  @apply border-b-disabled-text text-disabled-text;
}

.input-field-outline {
  @apply absolute top-0 -left-2 bottom-0 -right-2 rounded-sm pointer-events-none;
  @apply peer-focus-within:outline-none peer-focus-within:outline-2 peer-focus-within:outline-black peer-focus-within:ring-2 peer-focus-within:ring-offset-2 peer-focus-within:ring-white;
}
</style>

<script lang="ts" setup>
import { computed } from 'vue';
import { useNetworkStore } from '@/stores/networkStore';
import type { LiveButton } from '@/shared/model/liveButton';
import type { LivePage } from '@/shared/model/livePage';

const props = defineProps<{
    ui: LiveButton;
    page: LivePage;
}>();

const networkStore = useNetworkStore();

function onClick() {
    networkStore.invokeFunc(props.ui.action);
}

const clazz = computed<string>(() => {
    switch (props.ui.color.value) {
        case 'primary':
            return 'btn-primary';
        case 'secondary':
            return 'btn-secondary';
      case 'tertiary':
        return 'btn-tertiary'
        case 'subtile':
            return 'btn-subtile';
        case 'destructive':
            return 'btn-destructive';
        default:
            return 'btn-default';
    }
});

const iconOnly = computed<boolean>(() => {
    return props.ui.caption.value == '' && props.ui.preIcon.value != '';
});
</script>

<style>
.btn-primary,
.btn-secondary,
.btn.btn-tertiary,
.btn-subtile,
.btn-destructive {
    @apply inline-flex items-center justify-center px-5 py-2.5;
    @apply rounded-lg;
    @apply text-center text-sm font-medium;
    @apply focus:outline-none focus:outline-black focus:outline-offset-2;
}

.btn-primary {
  @apply rounded-3xl h-10 w-fit px-5 text-white bg-wdy-green;
  @apply hover:bg-wdy-green hover:opacity-90;
  @apply active:bg-opacity-75;
  @apply disabled:text-gray-400 disabled:bg-gray-200  disabled:hover:opacity-100;
  @apply dark:bg-opacity-0 dark:hover:bg-opacity-10 dark:focus:bg-opacity-25
}

.btn-secondary {
  @apply rounded-3xl h-10 w-fit px-5 text-black border border-black;
  @apply hover:text-wdy-green hover:border-wdy-green ;
  @apply active:bg-wdy-green active:bg-opacity-25;
  @apply focus:border-none focus:text-wdy-green;
  @apply disabled:text-gray-200 disabled:border-gray-200 disabled:focus:bg-white;
  @apply dark:text-white;
}

.btn-tertiary {
  @apply text-wdy-green rounded-3xl h-10 w-fit px-5;
  @apply hover:bg-wdy-green hover:bg-opacity-10;
  @apply active:bg-wdy-green active:bg-opacity-25;
  @apply focus:outline-none focus:outline-black focus:outline-offset-2;
  @apply disabled:text-gray-200 disabled:hover:bg-white disabled:focus:bg-white;

}

.btn-subtile {
    @apply hover:bg-gray-200;
    @apply focus:ring-gray-300;
    @apply border text-gray-900;
    @apply dark:border-gray-600 dark:text-white dark:hover:border-gray-600 dark:hover:bg-gray-700 dark:focus:ring-gray-700;
}

.btn-destructive {
    @apply focus:ring-gray-300;
    @apply bg-red-700 text-white hover:bg-red-800 focus:outline-none dark:bg-red-600 dark:hover:bg-red-700 dark:focus:ring-red-900;
}



.btn-default {
    @apply rounded-lg px-5  py-2.5 text-sm font-medium text-gray-900 hover:bg-gray-100 hover:text-blue-700 focus:z-10 focus:outline-none focus:ring-4 focus:ring-gray-200 dark:border-gray-600  dark:text-gray-400 dark:hover:bg-gray-700 dark:hover:text-white dark:focus:ring-gray-700;
}
</style>

<template>
    <button :class="clazz" @click="onClick" :disabled="props.ui.disabled.value">
        <svg
            v-inline
            class="me-2 h-3.5 w-3.5"
            v-html="props.ui.preIcon.value"
            v-if="props.ui.preIcon.value && !iconOnly"
        ></svg>
        {{ props.ui.caption.value }}
        <span class="ms-2 h-3.5 w-3.5" v-html="props.ui.postIcon.value" v-if="props.ui.postIcon.value"></span>
        <svg v-inline class="h-4 w-4" v-if="iconOnly" v-html="props.ui.preIcon.value"></svg>
    </button>
</template>

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

<script lang="ts" setup>

import { ButtonElement } from "@/shared/model";

const props = defineProps<{
    ui: ButtonElement;
}>();

</script>

<template>
    <button class="px-2 py-1 border">
        {{ props.ui.title.value }}
    </button>
</template>
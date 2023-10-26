<script lang="ts" setup>

import { TextElement } from "@/shared/model";

const props = defineProps<{
    ui: TextElement,
}>();

</script>

<template>
    <p>{{ props.ui.value }}</p>
</template>
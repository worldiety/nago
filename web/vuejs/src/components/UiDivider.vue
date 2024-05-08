<script lang="ts" setup>
import {Divider} from "@/shared/protocol/ora/divider";

const props = defineProps<{
	ui: Divider;
}>();
</script>

<template>
	<hr class="border-t-1 opacity-100 dark:opacity-50" />
</template>

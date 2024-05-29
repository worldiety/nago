<script lang="ts" setup>
import type { Divider } from '@/shared/protocol/ora/divider';

defineProps<{
	ui: Divider;
}>();
</script>

<template>
	<hr class="border-t-1 opacity-100" />
</template>

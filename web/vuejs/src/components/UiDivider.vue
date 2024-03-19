<script lang="ts" setup>
import type { Divider } from '@/shared/model/divider';
import type { LivePage } from '@/shared/model/livePage';

const props = defineProps<{
	ui: Divider;
	page: LivePage;
}>();
</script>

<template>
	<hr class="border-t-1 opacity-100 dark:opacity-50" />
</template>

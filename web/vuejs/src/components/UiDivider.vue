<script lang="ts" setup>
import type { LivePage } from '@/shared/model/livePage';
import {Divider} from "@/shared/protocol/gen/divider";

const props = defineProps<{
	ui: Divider;
	page: LivePage;
}>();
</script>

<template>
	<hr class="border-t-1 opacity-100 dark:opacity-50" />
</template>

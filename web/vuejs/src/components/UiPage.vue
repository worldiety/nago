<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import type { LivePage } from '@/shared/model/livePage';
import {Page} from "@/shared/protocol/gen/page";



const props = defineProps<{
	ui: Page;
	page: LivePage; // TODO refactor me, we should not need that
}>();

console.log("he in page",props.ui)
</script>

<template>
	<div class="grid grid-cols-1 gap-y-3">
		<ui-generic v-if="props.ui.body.v" :ui="props.ui.body.v" :page="page" />
	</div>
</template>

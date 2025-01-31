<template>
	<div class="flex justify-start items-center gap-x-2 text-black -ml-3">
		<template v-for="(item, index) in ui.items.v" :key="index">
			<UiBreadcrumbItem :breadcrumb-item="item" :icon="index === 0 ? ui.icon.v : null" />
			<ArrowRightBold v-if="index < ui.items.v.length - 1" class="shrink-0 h-3 *:h-full" />
		</template>
	</div>
</template>

<script setup lang="ts">
import ArrowRightBold from '@/assets/svg/arrowRightBold.svg';
import UiBreadcrumbItem from '@/components/breadcrumbs/UiBreadcrumbItem.vue';
import type { Breadcrumbs } from '@/shared/protocol/gen/breadcrumbs';

defineProps<{
	ui: Breadcrumbs;
}>();
</script>

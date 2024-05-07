<template>
	<div class="flex justify-start items-center gap-x-4 -ml-3">
		<template v-for="(item, index) in ui.items.v" :key="index">
			<div
				class="flex justify-start items-center gap-x-2 hover:bg-ora-orange hover:bg-opacity-15 active:bg-opacity-25 cursor-pointer rounded-full px-3"
				tabindex="0"
			>
				<div v-if="ui.icon.v && index === 0" class="h-4 *:h-full" v-html="ui.icon.v"></div>
				<span>{{ item }}</span>
			</div>
			<ArrowRightBold v-if="index < ui.items.v.length - 1" class="h-3 *:h-full"></ArrowRightBold>
		</template>
	</div>
</template>

<script setup lang="ts">
import type { Breadcrumbs } from '@/shared/protocol/gen/breadcrumbs';
import ArrowRightBold from '@/assets/svg/arrowRightBold.svg';

defineProps<{
	ui: Breadcrumbs;
}>();
</script>

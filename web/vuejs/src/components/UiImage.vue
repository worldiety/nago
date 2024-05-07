<script lang="ts" setup>
import type {Image} from "@/shared/protocol/ora/image";

const props = defineProps<{
	ui: Image;
}>();

function getSource(): string {
	if (props.ui.url.v === '/api/v1/download') {
		return props.ui.url.v + '?page=???' +  + '&download=' + props.ui.downloadToken.v; // TODO we had the page token here, but why not just omit that and just use the token? the backend can resolve that easily
	}

	return props.ui.url.v;
}
</script>

<template>
	<figure class="mx-auto max-w-lg">
		<img class="h-auto max-w-full rounded-lg" :src="getSource()" :alt="props.ui.caption.v" />
		<figcaption v-if="props.ui.caption.v" class="mt-2 text-center text-sm text-gray-500 dark:text-gray-400">
			{{ props.ui.caption.v }}
		</figcaption>
	</figure>
</template>

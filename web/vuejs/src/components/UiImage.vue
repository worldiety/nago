<script lang="ts" setup>
import type { Ref } from 'vue';
import { inject } from 'vue';
import { useNetworkStore } from '@/stores/networkStore';
import type { LivePage } from '@/shared/model/livePage';
import {Image} from "@/shared/protocol/gen/image";

const props = defineProps<{
	ui: Image;
	page: LivePage;
}>();

const networkStore = useNetworkStore();



function getSource(): string {
	if (props.ui.url.v === '/api/v1/download') {
		return props.ui.url.v + '?page=' + props.page.token + '&download=' + props.ui.downloadToken.v;
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

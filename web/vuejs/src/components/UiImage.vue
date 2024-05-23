<script lang="ts" setup>
import type {Image} from "@/shared/protocol/ora/image";

const props = defineProps<{
	ui: Image;
}>();

console.log(props.ui.uri)

</script>

<template>
	<figure class="">
		<img class="h-auto max-w-full rounded-lg" :src="props.ui.uri.v" :alt="props.ui.caption.v" />
		<figcaption v-if="props.ui.caption.v" class="mt-2 text-center text-sm text-gray-500 dark:text-gray-400">
			{{ props.ui.caption.v }}
		</figcaption>
	</figure>
</template>

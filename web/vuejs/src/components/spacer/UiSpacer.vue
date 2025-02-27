<script lang="ts" setup>
import { computed } from 'vue';
import { borderCSS } from '@/components/shared/border';
import { frameCSS } from '@/components/shared/frame';
import {Spacer} from "@/shared/proto/nprotoc_gen";
import {colorValue} from "@/components/shared/colors";

const props = defineProps<{
	ui: Spacer;
}>();

const styles = computed<string>(() => {
	let styles = borderCSS(props.ui.border);
	styles.push(...frameCSS(props.ui.frame));
	if (props.ui.backgroundColor) {
		styles.push(`background-color: ${colorValue(props.ui.backgroundColor)}`);
	}

	styles.push('object-fit: cover');

	return styles.join(';');
});
</script>

<template v-if="props.ui.children">
	<!-- spacer -->
	<div class="grow shrink" :style="styles"></div>
</template>

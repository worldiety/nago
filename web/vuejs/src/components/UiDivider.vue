<script lang="ts" setup>
import { computed } from 'vue';
import { borderCSS } from '@/components/shared/border';
import { frameCSS } from '@/components/shared/frame';
import { marginCSS } from '@/components/shared/padding';
import type { Divider } from '@/shared/protocol/ora/divider';

const props = defineProps<{
	ui: Divider;
}>();

const styles = computed<string>(() => {
	let styles = borderCSS(props.ui.b);
	styles.push(...frameCSS(props.ui.f));
	styles.push(...marginCSS(props.ui.p));

	return styles.join(';');
});
</script>

<template>
	<div :style="styles" />
</template>

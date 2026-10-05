<template>
	<component :is="tag || 'span'" ref="container" />
</template>
<script lang="ts" setup>
import { onMounted, ref, watch } from 'vue';
import { sanitizeMarkup } from '@/shared/sanitize';

interface Props {
	html: string;
	tag?: string;
}

const props = defineProps<Props>();
const container = ref<HTMLSpanElement>();

onMounted(loadSecureHtml);
watch(() => props.html, loadSecureHtml);

// Inserts sanitized HTML or SVG into the DOM, see sanitizeMarkup.
function loadSecureHtml(): void {
	if (!container.value) {
		return;
	}

	container.value.innerHTML = sanitizeMarkup(props.html);
}
</script>

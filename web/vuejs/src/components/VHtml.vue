<template>
	<component :is="tag || 'span'" ref="container" />
</template>
<script lang="ts" setup>
import { onMounted, ref } from 'vue';

interface Props {
	html: string;
	tag?: string;
}

const props = defineProps<Props>();
const container = ref<HTMLSpanElement>();

onMounted(loadSecureHtml);

// Removing script tags before inserting html into DOM.
function loadSecureHtml(): void {
	const elem = document.createElement('span');
	elem.innerHTML = props.html;
	const scripts = elem.querySelectorAll('script');
	scripts.forEach((script) => script.remove());
	if (container.value) {
		container.value.innerHTML = elem.innerHTML;
	}
}
</script>

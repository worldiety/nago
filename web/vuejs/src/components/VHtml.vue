<template>
	<component :is="tag || 'span'" ref="container" />
</template>
<script lang="ts" setup>
import { onMounted, ref, watch } from 'vue';


interface Props {
	html: string;
	tag?: string;
}

const props = defineProps<Props>();
const container = ref<HTMLSpanElement>();

onMounted(loadSecureHtml);
watch(() => props.html, loadSecureHtml);

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

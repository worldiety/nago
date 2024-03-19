<script lang="ts" setup>
import { ref } from 'vue';
import type { CustomError } from '@/composables/errorhandling';

const props = defineProps<{
	error: CustomError;
}>();

const showAdditionalInformation = ref<boolean>(false);

async function toggleErrorInfo(): Promise<void> {
	showAdditionalInformation.value = !showAdditionalInformation.value;
}
</script>

<template>
	<div class="rounded-md border border-gray-600 bg-gray-50 p-4">
		<p>{{ props.error.message }}</p>
		<button class="border border-black bg-white p-0.5 text-sm" @click="toggleErrorInfo">Mehr Informationen</button>
		<p v-if="showAdditionalInformation">{{ props.error.additionalInformation }}</p>
	</div>
</template>

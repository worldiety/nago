<script setup lang="ts">
import Page from '@/views/Page.vue';
import UiErrorMessage from '@/components/UiErrorMessage.vue';
import { ApplicationError, type CustomError, useErrorHandling } from '@/composables/errorhandling';


const errorHandler = useErrorHandling();
//TODO: Torben baut zukünftig /health ein, der einen 200er und eine json-response zurückgibt, wenn der Service grundsätzlich läuft

async function init(): Promise<void> {

}

init();
</script>

<template>
	<div v-if="errorHandler.error.value" class="flex h-screen items-center justify-center">
		<UiErrorMessage :error="errorHandler.error.value"> </UiErrorMessage>
	</div>
	<Page/>
</template>

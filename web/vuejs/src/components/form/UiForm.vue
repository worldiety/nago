<template v-if="props.ui.iv">
	<form :id="ui.id" @submit.prevent="handleSubmit" :autocomplete="ui.autocomplete ? 'on' : 'off'">
		<ui-generic v-for="ui in props.ui.children?.value" :ui="ui" />
	</form>
</template>

<script setup lang="ts">
import UiGeneric from '@/components/UiGeneric.vue';
import { useServiceAdapter } from '@/composables/serviceAdapter';
import { nextRID } from '@/eventhandling';
import { Form, FunctionCallRequested } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: Form;
}>();

const serviceAdapter = useServiceAdapter();

function handleSubmit(event: Event) {
	if (props.ui.action) {
		event.stopPropagation();
		serviceAdapter.sendEvent(new FunctionCallRequested(props.ui.action, nextRID()));
	}
}
</script>

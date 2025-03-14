<script setup lang="ts">
import { computed } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';
import { frameCSS } from '@/components/shared/frame';
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

const frameStyles = computed<string>(() => {
	let styles = frameCSS(props.ui.frame);

	return styles.join(';');
});
</script>

<template v-if="props.ui.iv">
	<form
		:style="frameStyles"
		:id="ui.id"
		@submit.prevent="handleSubmit"
		:autocomplete="ui.autocomplete ? 'on' : 'off'"
	>
		<ui-generic v-for="ui in props.ui.children?.value" :ui="ui" />
	</form>
</template>

<script lang="ts" setup>
import type { Component as OraComponent } from '@/shared/proto/nprotoc_gen';

const props = defineProps<{
	ui: OraComponent; // if we don't rename the import, Vue confuses everything up
}>();
</script>

<template>
	<div v-if="props.ui">[Protocol type has not been implemented: '{{ props.ui.constructor.name }}']</div>
	<div v-else>ui component is null</div>
</template>

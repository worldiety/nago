<script lang="ts" setup>
import type { Component as OraComponent } from '@/shared/proto/nprotoc_gen';
import uiComponentsMap from '@/shared/uiComponents';

const props = defineProps<{
	ui: OraComponent; // if we don't rename the import, Vue confuses everything up
}>();
</script>

<template>
	<div>[Protocol type has not been implemented: '{{ props.ui.constructor.name }}']</div>
</template>

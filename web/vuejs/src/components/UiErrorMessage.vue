<script lang="ts" setup>
import {ref} from 'vue';
import type {CustomError} from "@/composables/errorhandling";

const props = defineProps<{
  error: CustomError
}>()

const showAdditionalInformation = ref<boolean>( false);

async function toggleErrorInfo():Promise<void> {
  showAdditionalInformation.value = !showAdditionalInformation.value
}

</script>


<template>
  <div class="border border-gray-600 p-4 rounded-md bg-gray-50">
    <p> {{props.error.message}}</p>
    <button class="border border-black p-0.5 text-sm bg-white" @click="toggleErrorInfo">Mehr Informationen</button>
    <p v-if="showAdditionalInformation">{{ props.error.additionalInformation }}</p>
  </div>
</template>
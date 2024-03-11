<script lang="ts" setup>
import UiGeneric from '@/components/UiGeneric.vue';
import { useFocus } from '@vueuse/core'
import { ref } from 'vue'
import type { LiveDialog, LivePage } from '@/shared/model';

const props = defineProps<{
    ui: LiveDialog;
    page: LivePage;
}>();

// enables focus on an element that has "dialog" as a ref, as soon as it gets created
const dialog = ref()
useFocus(dialog, { initialValue: true })

</script>

<template>
    <div ref="dialog" class="bg-white text-left shadow-xl transition-all dark:bg-gray-500" tabindex="0">
        <div class="px-4 pb-4 pt-5 dark:bg-gray-500 sm:p-6 sm:pb-4">
            <div class="sm:flex sm:items-start">
                <div
                    v-if="props.ui.icon.value"
                    class="mx-auto flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full bg-red-100 sm:mx-0 sm:h-10 sm:w-10"
                >
                    <svg v-inline class="h-6 w-6 text-red-600" v-html="props.ui.icon.value"></svg>
                </div>
                <div class="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left">
                    <h3 class="text-base font-semibold leading-6 text-gray-900" id="modal-title">
                        {{ props.ui.title.value }}
                    </h3>
                    <div class="mt-2">
                        <div class="text-sm text-gray-500">
                            <ui-generic v-if="props.ui.body.value" :ui="props.ui.body.value" :page="page" />
                        </div>
                    </div>
                </div>
            </div>
        </div>
        <div class="bg-gray-50 px-4 py-3 dark:bg-gray-600 sm:flex sm:flex-row-reverse sm:px-6">
            <ui-generic
                class="inline-flex w-full justify-center rounded-md px-3 py-2 text-sm font-semibold text-white shadow-sm sm:ml-3 sm:w-auto"
                v-for="action in props.ui.actions.value"
                :ui="action"
                :page="page"
            />
        </div>
    </div>
</template>


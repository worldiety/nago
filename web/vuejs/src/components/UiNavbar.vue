<script lang="ts" setup>
import type { NavbarElement } from '@/shared/model';
import { computed } from 'vue';
import UiGeneric from '@/components/UiGeneric.vue';

const props = defineProps<{
    ui: NavbarElement;
}>();

//TODO we get into trouble using tailwind pre-processor here
const style = computed<string>(
    () => `

`
);
</script>

<template>
    <header class="flex flex-wrap items-center bg-white px-4 py-4 shadow-md lg:px-16">
        <div class="flex flex-1 items-center justify-between">
            <a href="#" class="text-xl"> <ui-generic :ui="props.ui.caption" /></a>
        </div>

        <label for="menu-toggle" class="pointer-cursor block md:hidden">
            <svg
                class="fill-current text-gray-900"
                xmlns="http://www.w3.org/2000/svg"
                width="20"
                height="20"
                viewBox="0 0 20 20"
            >
                <title>menu</title>
                <path d="M0 3h20v2H0V3zm0 6h20v2H0V9zm0 6h20v2H0v-2z"></path>
            </svg>
        </label>
        <input id="menu-toggle" class="hidden" type="checkbox" />

        <div id="menu" class="hidden w-full md:flex md:w-auto md:items-center">
            <nav>
                <ul class="items-center justify-between pt-4 text-base text-gray-700 md:flex md:pt-0">
                    <li v-for="ui in props.ui.menuItems">
                        <a class="block px-0 py-3 md:p-4" href="#"> <ui-generic :ui="ui" /></a>
                    </li>
                </ul>
            </nav>
        </div>
    </header>
</template>

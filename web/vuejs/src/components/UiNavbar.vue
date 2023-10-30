<script lang="ts" setup>

import {NavbarElement} from "@/shared/model";
import {computed} from "vue";
import UiGeneric from "@/components/UiGeneric.vue";

const props = defineProps<{
  ui: NavbarElement,
}>();

//TODO we get into trouble using tailwind pre-processor here
const style = computed<string>(() => `

`);

</script>

<template>
  <header class="lg:px-16 px-4 bg-white flex flex-wrap items-center py-4 shadow-md">

    <div class="flex-1 flex justify-between items-center">
      <a href="#" class="text-xl"> <ui-generic :ui="props.ui.caption" /></a>
    </div>


    <label for="menu-toggle" class="pointer-cursor md:hidden block">
      <svg class="fill-current text-gray-900"
           xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 20 20">
        <title>menu</title>
        <path d="M0 3h20v2H0V3zm0 6h20v2H0V9zm0 6h20v2H0v-2z"></path>
      </svg>
    </label>
    <input class="hidden" type="checkbox" id="menu-toggle" />

    <div class="hidden md:flex md:items-center md:w-auto w-full" id="menu">
      <nav>
        <ul class="md:flex items-center justify-between text-base text-gray-700 pt-4 md:pt-0">
          <li v-for="ui in props.ui.menuItems">
            <a class="md:p-4 py-3 px-0 block" href="#"> <ui-generic :ui="ui" /></a>


          </li>
        </ul>
      </nav>
    </div>





  </header>
</template>
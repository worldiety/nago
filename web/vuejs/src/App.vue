<script setup lang="ts">
import {createRouter, createWebHashHistory, createWebHistory, RouterView} from "vue-router";
import { useAppNavigation } from "./stores/appNavigation";



//?
//const appNavigation = useAppNavigation();
//appNavigation.updateRoute("");
</script>

<template>
    <header>
        <h1 class="text-3xl font-bold underline">Hello world!</h1>
        <!-- <span>{{ appNavigation.route }}</span> -->
    </header>

    <RouterView />
</template>

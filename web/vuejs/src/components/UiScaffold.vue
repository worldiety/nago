<script lang="ts" setup>

import { Scaffold } from "@/shared/model";
import {ref} from "vue";
import UiTableRow from "@/components/UiTableRow.vue";

const props = defineProps<{
    ui: Scaffold,
}>();

</script>

<template>
  <v-app class="rounded rounded-md">
    <v-app-bar :title="props.ui.title"></v-app-bar>

    <v-navigation-drawer
        expand-on-hover
        rail
    >
      <v-list>
        <v-divider></v-divider>
        <v-list-item :prepend-icon="navItem.icon.name" :href="navItem.action.target" link v-for="navItem in props.ui.navigation" :ui="navItem" :title="navItem.title">

        </v-list-item>
      </v-list>
    </v-navigation-drawer>

    <v-main class="d-flex align-center justify-center" style="min-height: 300px;">
    </v-main>


    <v-bottom-navigation class="d-flex d-lg-none">

      <v-btn :href="navItem.action.target" link v-for="navItem in props.ui.navigation" :ui="navItem" >
      <v-icon>{{navItem.icon.name}}</v-icon>
        <span>{{navItem.title}}</span>
      </v-btn>

    </v-bottom-navigation>

  </v-app>

</template>
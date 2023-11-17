<script lang="ts" setup>
import type {Scaffold} from '@/shared/model';
import {UiElement} from "@/shared/model";
import {ref} from 'vue';
import UiGeneric from "@/components/UiGeneric.vue";

const props = defineProps<{
  ui: Scaffold;
}>();

const children = ref<UiElement[]>()

async function init(): Promise<void> {
  let res = new Array<UiElement>();
  for (const url of props.ui.children) {
    const e = await fetch(url).then((r) => r.json());
    res.push(e)
  }
  children.value = res
}


init()
</script>

<template>
  <v-app class="rounded rounded-md">
    <v-app-bar :title="props.ui.title"></v-app-bar>

    <v-navigation-drawer expand-on-hover rail>
      <v-list>
        <v-divider></v-divider>
        <v-list-item
            v-for="navItem in props.ui.navigation"
            :prepend-icon="navItem.icon.name"
            :href="navItem.link"
            link
            :ui="navItem"
            :title="navItem.title"
        >
        </v-list-item>
      </v-list>
    </v-navigation-drawer>

    <v-main class="d-flex align-center justify-center" style="min-height: 300px">
      <v-container>
           <ui-generic v-for="e in children" :ui="e" />
      </v-container>
    </v-main>

    <v-bottom-navigation class="d-flex d-lg-none">
      <v-btn v-for="navItem in props.ui.navigation" :href="navItem.link" link :ui="navItem">
        <v-icon>{{ navItem.icon.name }}</v-icon>
        <span>{{ navItem.title }}</span>
      </v-btn>
    </v-bottom-navigation>
  </v-app>
</template>

<script lang="ts" setup>
import type {TextElement, TimelineElement} from '@/shared/model';

const props = defineProps<{
    ui: TimelineElement;
}>();
</script>

<template>
  <v-timeline direction="horizontal">

    <v-timeline-item :dot-color="item.color" :icon="item.icon?.name" v-for="item in props.ui.items">
      <template v-if="item.alternateDotText" v-slot:icon >
        <span>{{item.alternateDotText}}</span>
      </template>

      <template v-slot:opposite>
        <router-link v-if="item.target" :to="item.target">{{item.title}}</router-link>
        <span v-else>{{item.title}}</span>
      </template>

    </v-timeline-item>

  </v-timeline>
</template>

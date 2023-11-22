<script lang="ts" setup>
import type {CardView} from '@/shared/model';
import router from "@/router";
import UiGeneric from "@/components/UiGeneric.vue";

const props = defineProps<{
  ui: CardView;
}>();
</script>

<template>

  <v-container>
    <v-row align="center" justify="center">
      <v-col cols="auto"
             v-for="card in props.ui.cards">

        <v-card
            class="mx-auto"
            max-width="368"
            min-width="300"
            :title="card.title"
            :subtitle="card.subtitle"
            :prepend-icon="card.prependIcon?.name"
            :append-icon="card.appendIcon?.name"
            :href="card.primaryAction?.target"
        >


          <ui-generic v-if="card.content" :ui="card.content"/>



          <v-card-actions v-if="card.actions?.length>0">

            <v-btn v-for="btn in card.actions"
                   variant="text"

                   @click="router.push(btn.action?.target)"
            >
              {{ btn.caption }}
            </v-btn>
          </v-card-actions>

        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

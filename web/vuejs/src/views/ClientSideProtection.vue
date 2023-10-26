<script lang="ts" setup>

import { useAuth } from "@/stores/auth";
import { computed } from "vue";

const auth = useAuth();
const name = computed(() => auth.user?.profile.given_name);

</script>

<template>
    <div class="flex flex-col gap-4 m-4">
        <p>Hi {{ name }}, this page can only be seen by logged in users.</p>
        <router-link to="/" class="underline">Back</router-link>
    </div>
</template>

<script lang="ts" setup>

import { useAuth } from "@/stores/auth";
import { useHttp } from "@/stores/http";
import { computed } from "vue";
import { useRouter } from "vue-router";

const auth = useAuth();
const http = useHttp();
const router = useRouter();

// A list of all dynamically loaded routes, for convenience
const dynamicRoutes = computed<string[]>(() =>
    router
        .getRoutes()
        .filter((r) => r.meta.page !== undefined)
        .map((r) => r.path)
);

async function login() {
    await auth.signIn();
}

async function logout() {
    await auth.signOut();
}

async function loadPrivateData() {
    const response = await http.request("http://localhost:3000/private");
    const text = await response.text();
    alert(text);
}

</script>

<template>
    <div class="flex flex-col items-start gap-4 m-4">
        <p>This page is public.</p>
        <template v-if="auth.user">
            <p>You are logged in as {{ auth.user?.profile.given_name }}.</p>
            <button @click="logout" class="underline">Logout</button>
        </template>
        <template v-else>
            <button @click="login" class="underline">Login</button>
        </template>
        <a href="http://localhost:8080" target="_blank" class="underline">Open Keycloak</a>
        <p>These routes were loaded from the server:</p>
        <ul>
            <li v-for="route in dynamicRoutes"><router-link :to="route" class="underline">{{ route }}</router-link></li>
        </ul>
    </div>
</template>

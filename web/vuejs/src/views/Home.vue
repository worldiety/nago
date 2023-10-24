<script lang="ts" setup>

import { useAuth } from "@/stores/auth";
import { useHttp } from "@/stores/http";

const auth = useAuth();
const http = useHttp();

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
        <router-link to="/protected-client" class="underline">Go to a client side protected page</router-link>
        <button @click="loadPrivateData" class="underline">Load private information from the server</button>
        <a href="http://localhost:8080" target="_blank" class="underline">Open Keycloak</a>
    </div>
</template>

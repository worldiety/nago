<script lang="ts" setup>
import { useAuth } from '@/stores/authStore';
import { useHttp } from '@/shared/http';
import { computed } from 'vue';
import { useRouter } from 'vue-router';

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
    const response = await http.request(import.meta.env.VITE_HOST_BACKEND + 'private');
    const text = await response.text();
    alert(text);
}
</script>

<template>
  <div>
    <p>This page is public.</p>

    <template v-if="auth.user">
      <p>You are logged in as {{ auth.user?.profile.given_name }}.</p>
      <button class="underline" @click="logout">Logout</button>
    </template>
    <template v-else>
      <button class="underline" @click="login">Login</button>
    </template>
    <a href="http://localhost:8080" target="_blank" class="underline">Open Keycloak</a>
    <p>These routes were loaded from the server:</p>
    <ul>
      <li v-for="(route,idx) in dynamicRoutes" :key="idx">
        <router-link :to="route" class="underline">{{ route }}</router-link>
      </li>
    </ul>
  </div>
</template>

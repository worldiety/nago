<script lang="ts" setup>
import { useAuth } from '@/stores/auth';
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
    <v-app class="rounded rounded-md">
        <v-app-bar title="Application bar"></v-app-bar>

        <v-navigation-drawer expand-on-hover rail>
            <v-list>
                <v-list-item title="My Application" subtitle="Vuetify"></v-list-item>
                <v-divider></v-divider>
                <v-list-item prepend-icon="mdi-folder" link title="List Item 1"></v-list-item>
                <v-list-item link title="List Item 2"></v-list-item>
                <v-list-item link title="List Item 3"></v-list-item>
            </v-list>
        </v-navigation-drawer>

        <v-main class="d-flex align-center justify-center" style="min-height: 300px">
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
        </v-main>

        <v-bottom-navigation class="d-flex d-lg-none">
            <v-btn value="recent">
                <v-icon>mdi-history</v-icon>

                <span>Recent</span>
            </v-btn>

            <v-btn value="favorites">
                <v-icon>mdi-heart</v-icon>

                <span>Favorites</span>
            </v-btn>

            <v-btn value="nearby">
                <v-icon>mdi-map-marker</v-icon>

                <span>Nearby</span>
            </v-btn>
        </v-bottom-navigation>
    </v-app>
</template>

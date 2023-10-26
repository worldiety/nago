<!--
  This is the OAuth redirect page for exchanging tokens.
  Our IDP will redirect here after login, so we can pass the PKCE challenge and obtain an access token.
-->

<script lang="ts" setup>
import { useAuth } from "@/stores/auth";
import { useRouter } from "vue-router";

const auth = useAuth();
const router = useRouter();

async function init() {
    try {
        await auth.signInCallback();
    } catch (e) {
        // Something went wrong, go back to the home page.
        console.log("handle signInCallback", e);
        await router.replace("/");
    }
}

init();
</script>

<template></template>

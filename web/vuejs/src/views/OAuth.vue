<!--
  This is the OAuth redirect page for exchanging tokens.
  Our IDP will redirect here after login, so we can pass the PKCE challenge and obtain an access token.
-->

<script lang="ts" setup>
import { useAuth } from '@/stores/authStore';
import { useRoute, useRouter } from 'vue-router';
import { watch } from 'vue';

const auth = useAuth();
const router = useRouter();
const route = useRoute();

async function init() {
	try {
		await auth.signInCallback();
	} catch (e) {
		// Something went wrong, go back to the home page.
		console.log('handle signInCallback', e);
		await router.replace('/');
	}
}

init();
watch(route, () => {
	init();
});
</script>

<template></template>

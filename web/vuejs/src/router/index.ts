import type { PageConfiguration } from '@/shared/model';
import { useAuth } from '@/stores/authStore';
import Home from '@/views/Home.vue';
import OAuth from '@/views/OAuth.vue';
import { createRouter, createWebHistory } from 'vue-router';

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/oauth',
            component: OAuth,
        },
        {
            path: '/',
            component: Home,
        },
    ],
});

interface PageMeta {
    page?: PageConfiguration;
}


// Make sure users only enter authenticated pages when they are signed in.
router.beforeEach(async (to, from, next) => {
    const meta = to.meta as PageMeta | undefined;
    const authenticated = meta?.page?.authenticated || false;

    console.log("before push",to)
    if (authenticated) {
        console.log("router: require auth")
        const auth = useAuth();
        const user = await auth.getUser;
        const accessToken = user?.access_token;
        if (accessToken) {
            console.log("router: access token is here, dispatching next")
            next();
            return
        } else {
            console.log("requires authentication but access token is null",user,accessToken)
            await auth.signIn(to.fullPath);
            //next(false)

            return;
        }
    }

    next();
});

export default router;

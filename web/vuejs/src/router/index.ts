import { useAuth } from "@/stores/auth";
import Home from "@/views/Home.vue";
import OAuth from "@/views/OAuth.vue";
import Secret from "@/views/ClientSideProtection.vue";
import { createRouter, createWebHistory, NavigationGuardWithThis } from "vue-router";

const checkAuthentication: NavigationGuardWithThis<undefined> = async (to, from, next) => {
    const auth = useAuth();
    const user = await auth.getUser();
    if (user && user.access_token) {
        next();
    } else {
        await auth.signIn(to.fullPath);
    }
};

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: "/oauth",
            component: OAuth,
        },
        {
            path: "/",
            component: Home,
        },
        {
            path: "/protected-client",
            component: Secret,
            beforeEnter: checkAuthentication,
        }
    ]
});

export default router;

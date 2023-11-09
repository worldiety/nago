import { GetRoute } from '@/components/usecases/GetRoute';
import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useAppNavigation = defineStore('appNavigation', () => {
    const route = ref('');
    const updateRoute = async (routeName: string) => {
        route.value = await GetRoute.apply(routeName);
    };

    return { route, updateRoute };
});

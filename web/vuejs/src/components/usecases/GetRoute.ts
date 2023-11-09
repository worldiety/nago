import { get } from '@/shared/FetchApi';

type Test = {
    test: string;
};

export class GetRoute {
    static async apply(route: string): Promise<string> {
        const response = (await get(`/api/v1/nago/route/render/${route}`)) as Test;
        return response.test;
    }
}

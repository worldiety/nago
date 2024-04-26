import type ServiceAdapter from '@/shared/network/serviceAdapter';
import { inject } from 'vue';
import { networkAdapterKey } from '@/shared/injectionKeys';

export function useServiceAdapter(): ServiceAdapter {
	const serviceAdapter = inject(networkAdapterKey);
	if (!serviceAdapter) {
		throw new Error('Could not inject ServiceAdapter as it is undefined');
	}
	return serviceAdapter;
}

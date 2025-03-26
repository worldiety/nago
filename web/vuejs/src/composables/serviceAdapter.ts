/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import { inject } from 'vue';
import { serviceAdapterKey } from '@/shared/injectionKeys';
import type ServiceAdapter from '@/shared/network/serviceAdapter';

export function useServiceAdapter(): ServiceAdapter {
	const serviceAdapter = inject(serviceAdapterKey);
	if (!serviceAdapter) {
		throw new Error('Could not inject ServiceAdapter as it is undefined');
	}
	return serviceAdapter;
}

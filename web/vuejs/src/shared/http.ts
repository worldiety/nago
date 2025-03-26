/**
 * Copyright (c) 2025 worldiety GmbH
 *
 * This file is part of the NAGO Low-Code Platform.
 * Licensed under the terms specified in the LICENSE file.
 *
 * SPDX-License-Identifier: Custom-License
 */
import type { CustomError } from '@/composables/errorhandling';
import type { URL } from 'node:url';

export class HttpRequest<T> {
	private readonly method: 'GET' | 'POST' | 'PUT' | 'DELETE';
	private readonly url: string | URL;
	private headers: Record<string, string> = {};
	private payload?: string | FormData;
	private auth?: () => Promise<string>;

	constructor(method: 'GET' | 'POST' | 'PUT' | 'DELETE', url: string | URL) {
		this.method = method;
		this.url = url;
	}

	public static get<T = undefined>(url: string | URL): HttpRequest<T> {
		return new HttpRequest<T>('GET', url);
	}

	public static post<T = undefined>(url: string | URL): HttpRequest<T> {
		return new HttpRequest<T>('POST', url);
	}

	public static put<T = undefined>(url: string | URL): HttpRequest<T> {
		return new HttpRequest<T>('PUT', url);
	}

	public static delete<T = undefined>(url: string | URL): HttpRequest<T> {
		return new HttpRequest<T>('DELETE', url);
	}

	public header(name: string, value: string): HttpRequest<T> {
		this.headers[name] = value;
		return this;
	}

	public authenticated(auth: () => Promise<string>): HttpRequest<T> {
		this.auth = auth;
		return this;
	}

	public body(body: string | Blob | FormData | object, contentType?: string): HttpRequest<T> {
		if (!['POST', 'PUT'].includes(this.method)) {
			throw new Error(`Request method ${this.method} does not support request body!`);
		}

		if (body instanceof Blob) {
			if (this.payload && this.payload instanceof FormData) {
				this.payload.append('files', body);
			} else {
				this.payload = new FormData();
				this.payload.append('files', body);
			}
			// https://muffinman.io/blog/uploading-files-using-fetch-multipart-form-data/
			delete this.headers['content-type'];
		} else if (body instanceof FormData) {
			this.payload = body;
		} else if (typeof body === 'string') {
			this.payload = body;
			this.headers['content-type'] = contentType || 'text/plain';
		} else if (typeof body === 'object') {
			this.payload = JSON.stringify(body);
			this.headers['content-type'] = contentType || 'application/json';
		} else {
			throw new Error('Unsupported request body!');
		}
		return this;
	}

	public async fetch(): Promise<T> {
		if (this.auth) {
			const token = await this.auth();
			this.headers['Authorization'] = `Bearer ${token}`;
		}

		let response: Response;
		try {
			response = await fetch(this.url, {
				method: this.method,
				body: this.payload,
				headers: this.headers,
				credentials: 'include',
			});
		} catch (e) {
			const customError: CustomError = {
				errorCode: '001',
			};
			throw customError;
		}

		if (!response.ok) {
			throw response;
		}

		try {
			switch (response.headers.get('content-type')?.toLowerCase()) {
				case 'application/json':
					return (await response.clone().json()) as T;
				case 'text/plain':
				case 'text/csv':
					return (await response.clone().text()) as T;
				default:
					//	return await response.clone().json() as T;
					return undefined as T;
			}
		} catch (e) {
			const customError: CustomError = {
				errorCode: '002',
			};
			throw customError;
		}
	}
}

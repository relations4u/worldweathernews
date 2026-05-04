import { PUBLIC_API_BASE_URL } from '$env/static/public';

import type { components } from './types.gen';

export type PingResponse = components['schemas']['PingResponse'];
export type Location = components['schemas']['Location'];
export type Problem = components['schemas']['Problem'];

export class ApiError extends Error {
	constructor(
		public status: number,
		public problem: Partial<Problem>
	) {
		super(`API ${status}: ${problem.title ?? 'unknown'}`);
		this.name = 'ApiError';
	}
}

interface RequestOptions extends RequestInit {
	timeout?: number;
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
	const { timeout = 10_000, ...init } = options;
	const controller = new AbortController();
	const timer = setTimeout(() => controller.abort(), timeout);

	try {
		const response = await fetch(`${PUBLIC_API_BASE_URL}${path}`, {
			...init,
			signal: controller.signal,
			headers: {
				Accept: 'application/json',
				'Content-Type': 'application/json',
				...init.headers
			}
		});

		if (!response.ok) {
			let problem: Partial<Problem> = {
				title: response.statusText,
				status: response.status
			};
			try {
				problem = (await response.json()) as Partial<Problem>;
			} catch {
				// nicht-JSON-Body, ignorieren
			}
			throw new ApiError(response.status, problem);
		}

		return (await response.json()) as T;
	} finally {
		clearTimeout(timer);
	}
}

export function ping(): Promise<PingResponse> {
	return request<PingResponse>('/api/v1/ping');
}

export function searchLocations(q: string, limit = 10): Promise<{ results: Location[] }> {
	const params = new URLSearchParams({ q, limit: String(limit) });
	return request<{ results: Location[] }>(`/api/v1/locations?${params}`);
}

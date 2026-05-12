import { PUBLIC_API_BASE_URL } from '$env/static/public';

import type { components } from './types.gen';

export type PingResponse = components['schemas']['PingResponse'];
export type Location = components['schemas']['Location'];
export type Observation = components['schemas']['Observation'];
export type ForecastEntry = components['schemas']['ForecastEntry'];
export type LocationDetail = components['schemas']['LocationDetail'];
export type Problem = components['schemas']['Problem'];

export type LocationsListResponse = {
	results: Location[];
	attribution: string;
};

export class ApiError extends Error {
	constructor(
		public status: number,
		public problem: Partial<Problem>
	) {
		super(`API ${status}: ${problem.title ?? 'unknown'}`);
		this.name = 'ApiError';
	}
}

type Fetch = typeof globalThis.fetch;

interface RequestOptions extends RequestInit {
	timeout?: number;
	fetch?: Fetch;
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
	const { timeout = 10_000, fetch: customFetch, ...init } = options;
	const doFetch: Fetch = customFetch ?? globalThis.fetch;
	const controller = new AbortController();
	const timer = setTimeout(() => controller.abort(), timeout);

	try {
		const response = await doFetch(`${PUBLIC_API_BASE_URL}${path}`, {
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

export function ping(options?: RequestOptions): Promise<PingResponse> {
	return request<PingResponse>('/api/v1/ping', options);
}

export function listLocations(options?: RequestOptions): Promise<LocationsListResponse> {
	return request<LocationsListResponse>('/api/v1/locations', options);
}

export function getLocationDetail(slug: string, options?: RequestOptions): Promise<LocationDetail> {
	return request<LocationDetail>(`/api/v1/locations/${encodeURIComponent(slug)}`, options);
}

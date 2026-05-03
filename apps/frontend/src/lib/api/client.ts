import { PUBLIC_API_BASE_URL } from '$env/static/public';

export class ApiError extends Error {
	constructor(
		public status: number,
		public detail: string,
		public traceId?: string
	) {
		super(`API ${status}: ${detail}`);
		this.name = 'ApiError';
	}
}

interface ApiOptions extends RequestInit {
	timeout?: number;
}

export async function apiFetch<T>(path: string, options: ApiOptions = {}): Promise<T> {
	const { timeout = 10_000, ...rest } = options;
	const controller = new AbortController();
	const timer = setTimeout(() => controller.abort(), timeout);

	try {
		const response = await fetch(`${PUBLIC_API_BASE_URL}${path}`, {
			...rest,
			signal: controller.signal,
			headers: {
				Accept: 'application/json',
				'Content-Type': 'application/json',
				...rest.headers
			}
		});

		if (!response.ok) {
			let detail = response.statusText;
			try {
				const body = await response.json();
				detail = body.detail || body.title || detail;
			} catch {
				// nicht-JSON-Body, ignorieren
			}
			throw new ApiError(response.status, detail);
		}

		return (await response.json()) as T;
	} finally {
		clearTimeout(timer);
	}
}

export interface PingResponse {
	message: string;
	traceId: string;
}

export async function ping(): Promise<PingResponse> {
	return apiFetch<PingResponse>('/api/v1/ping');
}

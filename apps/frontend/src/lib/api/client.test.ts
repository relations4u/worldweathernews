import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { ApiError, ping } from './client';

describe('ApiError', () => {
	it('formatiert die Message aus Status und Problem-Title', () => {
		const err = new ApiError(404, { title: 'Not Found', status: 404 });
		expect(err.message).toBe('API 404: Not Found');
		expect(err.status).toBe(404);
		expect(err.problem.title).toBe('Not Found');
	});

	it('fällt auf "unknown" zurück wenn Title fehlt', () => {
		const err = new ApiError(500, {});
		expect(err.message).toBe('API 500: unknown');
	});

	it('ist Instanz von Error mit Name "ApiError"', () => {
		const err = new ApiError(500, { title: 'boom', status: 500 });
		expect(err).toBeInstanceOf(Error);
		expect(err.name).toBe('ApiError');
	});
});

describe('request via ping()', () => {
	const originalFetch = globalThis.fetch;

	beforeEach(() => {
		vi.useFakeTimers();
	});

	afterEach(() => {
		vi.useRealTimers();
		globalThis.fetch = originalFetch;
	});

	it('parsed erfolgreichen JSON-Response als PingResponse', async () => {
		globalThis.fetch = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ message: 'pong', traceId: 't-1' }), {
				status: 200,
				headers: { 'Content-Type': 'application/json' }
			})
		);

		const result = await ping();
		expect(result).toEqual({ message: 'pong', traceId: 't-1' });
	});

	it('wirft ApiError mit problem-Body bei HTTP-Fehler', async () => {
		globalThis.fetch = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ title: 'oops', status: 500, detail: 'something broke' }), {
				status: 500,
				statusText: 'Server Error',
				headers: { 'Content-Type': 'application/problem+json' }
			})
		);

		await expect(ping()).rejects.toMatchObject({
			status: 500,
			problem: { title: 'oops', detail: 'something broke' }
		});
	});
});

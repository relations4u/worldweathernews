import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest';
import { ApiError, apiFetch } from './client';

describe('ApiError', () => {
	it('formatiert die Message aus Status und Detail', () => {
		const err = new ApiError(404, 'Not Found');
		expect(err.message).toBe('API 404: Not Found');
		expect(err.status).toBe(404);
		expect(err.detail).toBe('Not Found');
	});

	it('ist Instanz von Error', () => {
		const err = new ApiError(500, 'boom');
		expect(err).toBeInstanceOf(Error);
		expect(err.name).toBe('ApiError');
	});
});

describe('apiFetch', () => {
	const originalFetch = globalThis.fetch;

	beforeEach(() => {
		vi.useFakeTimers();
	});

	afterEach(() => {
		vi.useRealTimers();
		globalThis.fetch = originalFetch;
	});

	it('parsed erfolgreichen JSON-Response', async () => {
		globalThis.fetch = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ ok: true }), {
				status: 200,
				headers: { 'Content-Type': 'application/json' }
			})
		);

		const result = await apiFetch<{ ok: boolean }>('/test');
		expect(result).toEqual({ ok: true });
	});

	it('wirft ApiError mit detail-Feld bei HTTP-Fehler', async () => {
		globalThis.fetch = vi.fn().mockResolvedValue(
			new Response(JSON.stringify({ detail: 'oops' }), {
				status: 500,
				statusText: 'Server Error',
				headers: { 'Content-Type': 'application/json' }
			})
		);

		await expect(apiFetch('/test')).rejects.toMatchObject({
			status: 500,
			detail: 'oops'
		});
	});
});

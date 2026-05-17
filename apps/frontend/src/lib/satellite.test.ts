import { describe, expect, it } from 'vitest';

import { bboxCorners, resolveFrameIndex } from './satellite';

describe('bboxCorners', () => {
	it('liefert TL, TR, BR, BL in MapLibre-Reihenfolge', () => {
		const corners = bboxCorners({ lonMin: -18, latMin: 30, lonMax: 40, latMax: 70 });
		expect(corners).toEqual([
			[-18, 70], // top-left
			[40, 70], // top-right
			[40, 30], // bottom-right
			[-18, 30] // bottom-left
		]);
	});
});

describe('resolveFrameIndex', () => {
	it('folgt dem neuesten Frame, wenn nichts gewählt ist', () => {
		expect(resolveFrameIndex(null, 96)).toBe(95);
	});

	it('respektiert eine konkrete Auswahl', () => {
		expect(resolveFrameIndex(10, 96)).toBe(10);
	});

	it('klemmt in [0, count-1]', () => {
		expect(resolveFrameIndex(200, 96)).toBe(95);
		expect(resolveFrameIndex(-5, 96)).toBe(0);
	});

	it('gibt 0 bei leerer Frame-Liste', () => {
		expect(resolveFrameIndex(null, 0)).toBe(0);
		expect(resolveFrameIndex(3, 0)).toBe(0);
	});
});

import { describe, expect, it } from 'vitest';

import { compass, windArrowRotationDeg } from './wind';

describe('compass', () => {
	it('mappt die Kardinalrichtungen', () => {
		expect(compass(0)).toBe('N');
		expect(compass(90)).toBe('O');
		expect(compass(180)).toBe('S');
		expect(compass(270)).toBe('W');
	});

	it('rundet auf den nächsten 45°-Sektor', () => {
		expect(compass(22)).toBe('N');
		expect(compass(23)).toBe('NO');
		expect(compass(45)).toBe('NO');
		expect(compass(315)).toBe('NW');
	});

	it('wrappt bei 360° zurück auf N', () => {
		expect(compass(360)).toBe('N');
	});

	it('gibt "—" für fehlende Richtung', () => {
		expect(compass(null)).toBe('—');
		expect(compass(undefined)).toBe('—');
	});
});

describe('windArrowRotationDeg', () => {
	it('dreht um +180° (zeigt wohin der Wind weht)', () => {
		expect(windArrowRotationDeg(0)).toBe(180);
		expect(windArrowRotationDeg(90)).toBe(270);
		expect(windArrowRotationDeg(180)).toBe(0);
		expect(windArrowRotationDeg(270)).toBe(90);
	});

	it('normalisiert auf 0–360', () => {
		expect(windArrowRotationDeg(200)).toBe(20);
		expect(windArrowRotationDeg(359)).toBe(179);
	});

	it('gibt null für fehlende Richtung', () => {
		expect(windArrowRotationDeg(null)).toBeNull();
		expect(windArrowRotationDeg(undefined)).toBeNull();
	});
});

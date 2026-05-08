import { browser } from '$app/environment';

const STORAGE_KEY = 'wwn-cookie-consent-v1';
const SCHEMA_VERSION = 1;

export type CookieCategory = 'essential' | 'functional' | 'analytics' | 'marketing';

export type CookieSelection = {
	functional: boolean;
	analytics: boolean;
	marketing: boolean;
};

export type CookieConsent = CookieSelection & {
	essential: true;
	timestamp: number;
	version: number;
};

function loadConsent(): CookieConsent | null {
	if (!browser) return null;
	const raw = localStorage.getItem(STORAGE_KEY);
	if (!raw) return null;
	try {
		const parsed = JSON.parse(raw) as Partial<CookieConsent>;
		if (parsed.version !== SCHEMA_VERSION) return null;
		return {
			essential: true,
			functional: !!parsed.functional,
			analytics: !!parsed.analytics,
			marketing: !!parsed.marketing,
			timestamp: typeof parsed.timestamp === 'number' ? parsed.timestamp : Date.now(),
			version: SCHEMA_VERSION
		};
	} catch {
		return null;
	}
}

function persistConsent(consent: CookieConsent): void {
	if (!browser) return;
	localStorage.setItem(STORAGE_KEY, JSON.stringify(consent));
}

class CookieConsentStore {
	#hydrated = $state(false);
	consent = $state<CookieConsent | null>(null);

	showBanner = $derived(this.#hydrated && this.consent === null);
	hasDecided = $derived(this.#hydrated && this.consent !== null);

	hydrate(): void {
		if (this.#hydrated) return;
		this.consent = loadConsent();
		this.#hydrated = true;
	}

	acceptAll(): void {
		this.#commit({ functional: true, analytics: true, marketing: true });
	}

	rejectAll(): void {
		this.#commit({ functional: false, analytics: false, marketing: false });
	}

	save(selection: CookieSelection): void {
		this.#commit(selection);
	}

	#commit(selection: CookieSelection): void {
		const consent: CookieConsent = {
			essential: true,
			...selection,
			timestamp: Date.now(),
			version: SCHEMA_VERSION
		};
		this.consent = consent;
		persistConsent(consent);
	}
}

export const cookieConsent = new CookieConsentStore();

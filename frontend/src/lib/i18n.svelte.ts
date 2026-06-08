import { en } from './translations/en';
import { de } from './translations/de';
import type { Translations } from './translations/index';
import { SUPPORTED_LANGS, type Lang } from './translations/index';

const translations: Record<string, Translations> = { en, de };
const STORAGE_KEY = 'wm-lang';

class Locale {
	lang = $state<Lang>('en');

	constructor() {
		if (typeof window === 'undefined') return;
		const stored = localStorage.getItem(STORAGE_KEY);
		if (stored && (SUPPORTED_LANGS as readonly string[]).includes(stored)) {
			this.lang = stored as Lang;
		} else {
			this.lang = this.detect();
		}
		if (document.documentElement.lang !== this.lang) {
			document.documentElement.lang = this.lang;
		}
	}

	private detect(): Lang {
		const pref = navigator.language.slice(0, 2);
		return (SUPPORTED_LANGS as readonly string[]).includes(pref)
			? (pref as Lang)
			: 'en';
	}

	set(lang: Lang) {
		this.lang = lang;
		try {
			localStorage.setItem(STORAGE_KEY, lang);
		} catch {}
		document.documentElement.lang = lang;
	}

	get current(): Translations {
		return translations[this.lang] ?? translations.en;
	}
}

export const locale = new Locale();

function resolve(obj: unknown, key: string): string | undefined {
	const parts = key.split('.');
	let cur: unknown = obj;
	for (const p of parts) {
		if (cur == null || typeof cur !== 'object') return undefined;
		cur = (cur as Record<string, unknown>)[p];
	}
	return typeof cur === 'string' ? cur : undefined;
}

export function t(
	key: string,
	params?: Record<string, string | number>
): string {
	let val = resolve(locale.current, key);
	if (val === undefined) val = resolve(translations.en, key);
	if (val === undefined) return key;
	if (params) {
		for (const [k, v] of Object.entries(params)) {
			val = val!.replace(`{${k}}`, String(v));
		}
	}
	return val!;
}

export function ordinal(n: number): string {
	if (locale.lang === 'de') return `${n}.`;
	const s = ['th', 'st', 'nd', 'rd'];
	const v = n % 100;
	return n + (s[(v - 20) % 10] || s[v] || s[0]);
}

export function stageLabel(stage: string): string {
	const map: Record<string, string> = {
		group: 'stages.groupStage',
		R32: 'stages.R32',
		R16: 'stages.R16',
		QF: 'stages.QF',
		SF: 'stages.SF',
		'3RD': 'stages.thirdPlace',
		FINAL: 'stages.final'
	};
	return t(map[stage] ?? stage);
}

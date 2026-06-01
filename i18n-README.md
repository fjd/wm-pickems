# i18n — Adding a New Language

The app supports English (`en`) and German (`de`). This guide explains how to add another language.

## Overview

- **Translation files** live in `frontend/src/lib/translations/` — one `.ts` file per language.
- **Type safety**: Every translation file must satisfy the `Translations` interface defined in `index.ts`. TypeScript will catch missing or extra keys at compile time.
- **Core module**: `frontend/src/lib/i18n.svelte.ts` provides `t()`, `ordinal()`, and `stageLabel()`.
- **Language selector**: Settings page (`settings/+page.svelte`) has a button group. New languages appear automatically once wired up.

## Step-by-step

### 1. Add the language code

In `frontend/src/lib/translations/index.ts`, add the new code to `SUPPORTED_LANGS`:

```typescript
export const SUPPORTED_LANGS = ['en', 'de', 'fr'] as const;
```

### 2. Create the translation file

Copy `en.ts` to a new file (e.g., `fr.ts`) and translate all values:

```typescript
import type { Translations } from './index';

export const fr: Translations = {
  nav: {
    home: 'Accueil',
    tips: 'Pronostics',
    // ... translate every field
  },
  // ... every section
};
```

TypeScript will error on any missing key — fix until it compiles cleanly.

### 3. Register the translation

In `frontend/src/lib/i18n.svelte.ts`, import and register the new file:

```typescript
import { fr } from './translations/fr';

const translations: Record<string, Translations> = { en, de, fr };
```

### 4. Add the language selector entry

In `frontend/src/routes/settings/+page.svelte`, add an entry to the `languages` array:

```typescript
const languages: { code: Lang; label: string }[] = [
  { code: 'en', label: 'English' },
  { code: 'de', label: 'Deutsch' },
  { code: 'fr', label: 'Français' }  // native name
];
```

Use the **native language name** (e.g., "Français", not "French").

### 5. Handle locale-specific formatting

The app already uses `locale.lang` for:
- **Date/time**: `new Date(...).toLocaleString(locale.lang, opts)`
- **Ordinals**: The `ordinal()` function in `i18n.svelte.ts` — add a branch for the new language if it doesn't follow English (`1st`) or German (`1.`) patterns.

### 6. Verify

```bash
cd frontend && npx svelte-check --threshold error
```

Zero errors = all keys are present and correctly typed.

## Translation key structure

Keys use dot-notation (`t('nav.home')`) to traverse nested objects:

| Section | Purpose |
|---|---|
| `nav` | Navigation labels |
| `stages` | Tournament stage names (R32, QF, SF, etc.) |
| `common` | Shared strings (loading, saving, buttons, etc.) |
| `auth` | Login, register, password reset |
| `settings` | Settings page (including language selector label) |
| `home` | Dashboard |
| `tips` / `tipCard` | Match predictions |
| `groupStandings` | Projected group tables |
| `forecast` | Tournament forecast (editable) |
| `forecastView` | Tournament forecast (read-only, friend view) |
| `tournament` | Live tournament tables and bracket |
| `leagues` | League list |
| `leagueDetail` | League detail + leaderboard + scoring legend |
| `pwa` | PWA install prompts |
| `join` | Invite link flow |
| `dev` | Dev tools (only visible with `WMP_DEV=1`) |
| `userMenu` | User dropdown menu |
| `errors` | All error messages |

## Interpolation

Use `{variable}` placeholders in translation values:

```typescript
// en.ts
{ greeting: 'Hi, {name}' }

// Usage
t('home.greeting', { name: auth.user?.name ?? '' })
```

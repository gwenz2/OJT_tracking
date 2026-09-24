# Frontend — React + Vite + TypeScript + Tailwind v4

Feature-based SPA frontend for the full-stack template.

## Stack

- React 19 + Vite + TypeScript
- Tailwind CSS v4 (via `@tailwindcss/vite`)
- React Router v7
- TanStack Query v5 (server state)
- Zustand (global UI state)
- React Hook Form + Zod (forms + validation)
- Axios (centralized API client)
- Vitest + React Testing Library (tests)
- oxlint (linting)

## Architecture

```
Page
 ↓
Feature Component
 ↓
Feature Hook
 ↓
TanStack Query
 ↓
API Layer (features/<x>/api)
 ↓
API Client (lib/api-client)
 ↓
Go API
```

### State ownership

| Concern         | Owner                                   |
|-----------------|-----------------------------------------|
| Server state    | TanStack Query                          |
| Global UI/auth  | Zustand stores (`auth-store`, `ui-store`) |
| Local UI        | React `useState`/`useReducer`           |

Never duplicate server data into Zustand.

## Path aliases

`@/*` → `src/*` (configured in `tsconfig.app.json` and `vite.config.ts`).

## Commands

```bash
npm install
npm run dev        # start Vite dev server (http://localhost:5173)
npm run build      # type-check + production build
npm run preview    # preview the production build
npm run lint       # oxlint
npm test           # vitest (watch)
npm test -- --run  # vitest (one-shot)
```

## Environment

Copy `.env.example` to `.env`:

```env
VITE_API_URL=http://localhost:8080/api/v1
```

Never expose backend secrets through `VITE_*` variables.

## Feature structure

Each feature follows:

```
features/<name>/
├── api/<name>.api.ts
├── components/
├── hooks/
├── pages/
├── schemas/
├── types/
└── index.ts        # public barrel
```

## Query keys

Use the centralized factory in `src/lib/query-keys.ts`. Never scatter raw
string keys across the codebase.

## Optimistic UI

See `features/users/hooks/use-users.ts` for the canonical optimistic update
pattern (cancel → snapshot → optimistic update → mutate → rollback on error
→ invalidate on settle).

Only use optimistic UI for safe, reversible mutations. For payments,
checkout, stock deduction, permission changes, or any irreversible /
security-sensitive operation, use a pending → server-confirmed flow.

## Adding a new feature

1. `src/features/products/` with the structure above
2. Add query keys to `src/lib/query-keys.ts`
3. Add routes to `src/app/router.tsx`
4. Add tests in `features/products/__tests__/`

---
name: nextjs
description: Next.js App Router development standards — Server Components, data fetching, route handlers, caching, layout patterns, and platform-specific AI traps. Load when writing TypeScript/React code in the frontend/ directory.
---

# Next.js App Router Standards

You are writing Next.js code using the App Router with Server Components
as the default. Follow these conventions exactly. When your instinct (or
a plausible-looking API) contradicts this document, trust this document.

## Project Layout (Monorepo)

```
frontend/
├── src/
│   ├── app/
│   │   ├── layout.tsx            # Root layout (Server Component)
│   │   ├── page.tsx              # Home page
│   │   ├── loading.tsx           # Root loading (applies to all)
│   │   ├── error.tsx             # Root error boundary (use client)
│   │   ├── not-found.tsx         # 404 page
│   │   ├── (marketing)/          # Route group (doesn't affect URL)
│   │   │   ├── layout.tsx
│   │   │   ├── about/
│   │   │   │   └── page.tsx
│   │   │   └── pricing/
│   │   │       └── page.tsx
│   │   ├── dashboard/
│   │   │   ├── layout.tsx
│   │   │   ├── page.tsx
│   │   │   ├── loading.tsx
│   │   │   └── error.tsx
│   │   └── api/
│   │       ├── users/
│   │       │   └── route.ts      # GET/POST /api/users
│   │       └── users/[id]/
│   │           └── route.ts      # GET/PUT/DELETE /api/users/:id
│   ├── components/
│   │   ├── ui/                   # Reusable UI components
│   │   │   ├── Button.tsx
│   │   │   └── Input.tsx
│   │   └── forms/                # Form components with use client
│   │       └── UserForm.tsx
│   ├── hooks/                    # Client-side hooks
│   │   └── useAuth.ts
│   ├── lib/                      # Shared utilities (works server + client)
│   │   ├── api.ts                # API client
│   │   └── utils.ts
│   └── types/                    # TypeScript types
│       └── index.ts
├── public/                       # Static assets
├── package.json
├── tsconfig.json
└── next.config.ts
```

### `.nvmrc` — pin Node version

Always include a `.nvmrc` file in the frontend root so `nvm use` picks
the correct Node version automatically:

```
# frontend/.nvmrc
24
```

Match the Node version to Next.js requirements. Next.js 16 requires
`>=20.9.0`. Use the latest active LTS (24 Krypton as of 2026-08).

**The agent shell does not auto-load nvm.** Always verify the active
Node version matches `.nvmrc` before running `npm` or `node` commands.
If mismatched, prefix with the nvm path:
```bash
node --version  # verify before proceeding
# If mismatch, use:
PATH="$HOME/.nvm/versions/node/v$(cat .nvmrc)/bin:$PATH" npm install
```

## Server Components by DEFAULT

### Rule: Never add `"use client"` unless you have to

Start every component as a Server Component. Only add `"use client"` when
the component needs ONE of:

| Need | Fix |
|------|-----|
| `useState`, `useReducer` | `"use client"` |
| `useEffect`, `useLayoutEffect` | `"use client"` |
| Event handlers (`onClick`, `onSubmit`) | `"use client"` |
| Browser APIs (`window`, `localStorage`) | `"use client"` |
| React Context (`createContext`, `useContext`) | `"use client"` |
| Custom hooks that use any of the above | `"use client"` |
| `useSearchParams()` | `"use client"` |

### DO — Data fetching in Server Components

```tsx
// ✅ Server Component: no "use client", async, fetch directly
export default async function UsersPage() {
  const users = await fetch("http://localhost:8080/api/users").then(r => r.json());
  return (
    <ul>
      {users.map((u) => (
        <li key={u.id}>{u.name}</li>
      ))}
    </ul>
  );
}
```

### DO — Pass data from Server to Client Component

```tsx
// ✅ Parent: Server Component
import { UserList } from "./UserList";

export default async function UsersPage() {
  const users = await fetch("http://localhost:8080/api/users").then(r => r.json());
  return <UserList users={users} />;
}

// ✅ Child: Client Component (has onClick, so needs "use client")
"use client";
export function UserList({ users }: { users: User[] }) {
  const [selected, setSelected] = useState<string | null>(null);
  return (
    <ul>
      {users.map((u) => (
        <li key={u.id} onClick={() => setSelected(u.id)}>
          {u.name}
        </li>
      ))}
    </ul>
  );
}
```

### DO NOT — Common Server/Client mistakes

| ❌ Wrong | ✅ Right |
|---|---|
| `"use client"` on a component that just renders children | Keep it as Server Component |
| `async` function in a `"use client"` component | Client Components can't be async |
| `useState` in a Server Component | Extract interactive part into a Client Component |
| `"use client"` at the top of `layout.tsx` | Layouts can be Server Components unless they use context |
| `"use server"` directive on a component | "use server" is for Server Actions only, not components |
| `export { metadata }` from a Client Component | metadata must be exported from Server Components |

## Data Fetching Decision Tree

```
Need data for display?
├── Yes, can fetch in Server Component → async SC, fetch directly
├── Yes, need interactivity too → fetch in parent SC, pass as props
├── Yes, need real-time → Route Handler + polling/SSE/WebSocket
└── No → just render

Need to mutate data?
├── Yes, from form → Server Action with useActionState
├── Yes, from route handler → app/api/.../route.ts with POST/PATCH/DELETE
└── No → just render
```

### DO — Server Action with useActionState

```tsx
// app/actions.ts
"use server";
import { revalidatePath } from "next/cache";

export async function createUser(prevState: any, formData: FormData) {
  const name = formData.get("name");
  const email = formData.get("email");
  const res = await fetch("http://localhost:8080/api/users", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name, email }),
  });
  if (!res.ok) {
    return { error: "Failed to create user" };
  }
  revalidatePath("/users");
  return { success: true };
}
```

```tsx
// app/users/new/page.tsx
"use client";
import { useActionState } from "react"; // NOT react-dom
import { createUser } from "../actions";

export default function NewUserPage() {
  const [state, formAction, pending] = useActionState(createUser, null);
  return (
    <form action={formAction}>
      <input name="name" required />
      <input name="email" type="email" required />
      <button disabled={pending}>Create</button>
      {state?.error && <p>{state.error}</p>}
    </form>
  );
}
```

### DO NOT — useActionState mistakes

| ❌ Wrong | ✅ Right |
|---|---|
| `import { useActionState } from "react-dom"` | `import { useActionState } from "react"` |
| `revalidatePath("/users")` with relative path | Use absolute path from app root |
| `fetch("api/users")` in Server Component | Use absolute URL or `process.env.API_URL` |

## Route Handlers

### DO — Standard REST route handler

```tsx
// app/api/users/route.ts
import { NextRequest, NextResponse } from "next/server";

export async function GET(request: NextRequest) {
  const { searchParams } = request.nextUrl;
  const page = searchParams.get("page") || "1";
  const res = await fetch(`http://localhost:8080/api/users?page=${page}`);
  const users = await res.json();
  return NextResponse.json(users);
}

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const res = await fetch("http://localhost:8080/api/users", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    const data = await res.json();
    return NextResponse.json(data, { status: res.status });
  } catch {
    return NextResponse.json({ error: "Invalid request body" }, { status: 400 });
  }
}
```

### DO — Dynamic route with params (Next.js 15+)

```tsx
// app/api/users/[id]/route.ts
export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;  // params is a Promise in Next.js 15+
  const res = await fetch(`http://localhost:8080/api/users/${id}`);
  const user = await res.json();
  if (!res.ok) return NextResponse.json(user, { status: res.status });
  return NextResponse.json(user);
}
```

### DO NOT — Route handler mistakes

| ❌ Wrong | ✅ Right |
|---|---|
| `params.id` (not await) | `const { id } = await params` in Next.js 15+ |
| `new NextResponse(...)` | `NextResponse.json(...)` |
| `NextAPIResponse` (doesn't exist) | `NextResponse` |
| `request.body` for JSON | `await request.json()` |
| No try/catch around `.json()` | Always wrap `request.json()` in try/catch |

## Layout / Page / Template Hierarchy

### DO — Nested layouts

```tsx
// app/layout.tsx — Root layout (applies everywhere)
export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <nav>Global Nav</nav>
        {children}
      </body>
    </html>
  );
}

// app/dashboard/layout.tsx — Dashboard layout (applies under /dashboard/*)
export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="dashboard">
      <aside>Dashboard Sidebar</aside>
      {children}
    </div>
  );
}
```

### DO — Loading, error, not-found

```tsx
// app/users/loading.tsx — Shown while page loads
export default function Loading() {
  return <div>Loading users...</div>;
}

// app/users/error.tsx — Must be "use client"; catches errors in children
"use client";
export default function Error({ error, reset }: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div>
      <h2>Something went wrong!</h2>
      <button onClick={reset}>Try again</button>
    </div>
  );
}

// app/users/[id]/not-found.tsx — Per-entity 404
export default function NotFound() {
  return <h1>User not found</h1>;
}
```

**Route-segment completeness rule:** every route segment whose page
fetches data server-side MUST ship `error.tsx` and `loading.tsx`, and
`not-found.tsx` when the page calls `notFound()`. Also ship a root
`app/error.tsx` and `app/not-found.tsx` — without them, a page that
re-throws lands on Next's unbranded default screen. Audit:

```bash
# every segment with a server-fetching page should appear with all three
ls src/app/**/page.tsx src/app/**/error.tsx src/app/**/loading.tsx
```

### DO NOT — Layout mistakes

| ❌ Wrong | ✅ Right |
|---|---|
| `searchParams` prop in layout | `searchParams` is only available in `page.tsx` |
| Layout without `children` prop | All layouts MUST accept and render `children` |
| `export const metadata` in `"use client"` layout | metadata only works in Server Components |

## Caching Traps (Next.js 15 Changed Defaults)

### DO NOT — Assume caching is still automatic

In Next.js 15, `fetch` no longer caches by default. `GET` route handlers
are no longer cached by default. Pages that read `cookies()` or `headers()`
are no longer statically rendered by default.

### DO — Explicit caching

```tsx
// Cache this fetch for 1 hour
const res = await fetch(url, { next: { revalidate: 3600 } });

// Never cache (dynamic, server-rendered on every request)
const res = await fetch(url, { cache: "no-store" });

// Use unstable_cache for non-fetch operations
import { unstable_cache } from "next/cache";
const getCachedUsers = unstable_cache(
  async () => { /* ... */ },
  ["users-list"],
  { revalidate: 3600 }
);
```

### DO — Explicit dynamic behavior

```tsx
// Force dynamic rendering
export const dynamic = "force-dynamic";

// Force static (for pages that can be pre-rendered)
export const dynamic = "force-static";

// Revalidate at most every N seconds
export const revalidate = 3600;
```

## Metadata

### DO — Static and dynamic metadata

```tsx
// Static metadata
export const metadata: Metadata = {
  title: "Users",
  description: "Manage your users",
};

// Dynamic metadata (awaited params in Next.js 15+)
export async function generateMetadata(
  { params }: { params: Promise<{ id: string }> }
): Promise<Metadata> {
  const { id } = await params;
  const user = await fetch(`http://localhost:8080/api/users/${id}`).then(r => r.json());
  return {
    title: user.name,
    description: `Profile for ${user.name}`,
    openGraph: { images: [user.avatar] },
  };
}
```

### DO NOT — Metadata mistakes

| ❌ Wrong | ✅ Right |
|---|---|
| `export const metadata` in Client Component | Must be in Server Component (or layout) |
| `params.id` in `generateMetadata` | `const { id } = await params` in Next.js 15+ |

## Suspense & useSearchParams

**`useSearchParams()` requires a `<Suspense>` boundary** above the
component that calls it — otherwise `next build` fails during static
generation (the error is masked while something else forces the whole
app dynamic, then bites later). Pattern: export a wrapper.

```tsx
function SearchContent() {
  const searchParams = useSearchParams();
  // ...
}

export default function SearchPage() {
  return (
    <Suspense fallback={<main className="..." />}>
      <SearchContent />
    </Suspense>
  );
}
```

The same applies to shared client components (grids, filters) that read
search params — wrap inside the component itself so every caller is
safe.

**Root-layout dynamic trap:** calling `cookies()`/`headers()` in the
root layout forces EVERY route in the app to dynamic rendering. Read
request data as deep in the tree as possible.

## Middleware

### DO — Edge-compatible middleware

```tsx
// src/middleware.ts
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
  // Auth guard
  const token = request.cookies.get("token")?.value;
  if (!token && request.nextUrl.pathname.startsWith("/dashboard")) {
    return NextResponse.redirect(new URL("/login", request.url));
  }
  return NextResponse.next();
}

export const config = {
  matcher: ["/dashboard/:path*"],  // only run on these paths
};
```

### DO NOT — Middleware edge runtime traps

| ❌ Wrong | ✅ Right |
|---|---|
| `import fs from "fs"` in middleware | Middleware runs on Edge; no Node.js APIs |
| `import { PrismaClient }` in middleware | No database clients in middleware |
| `import jwt from "jsonwebtoken"` | Use `jose` for JWT in Edge runtime |
| `request.page` | Use `request.nextUrl.pathname` |
| Matcher as regex | Matcher is glob patterns only |

## Common AI Hallucinations — Complete Reference

### Fake Next.js APIs

| ❌ Hallucination | ✅ Reality |
|---|---|
| `NextAPIResponse` | `NextResponse` |
| `NextAPIRequest` | `NextRequest` |
| `useRouter` from `next/navigation` with `router.query` | `useSearchParams()` or `useParams()` |
| `getServerSideProps` | Does not exist in App Router (Pages Router only) |
| `getStaticProps` | Does not exist in App Router |
| `getStaticPaths` | Use `generateStaticParams` |
| `useAppRouter` | `useRouter` from `next/navigation` |
| `useHistory` | `useRouter` from `next/navigation` |
| `redirect` from `next/navigation` in Server Component | `redirect` from `next/navigation` works in both, but typically used in Server Components/Actions |
| `ImageResponse` from `next/server` | `ImageResponse` is fine but needs explicit import |
| `NextApiRequest`, `NextApiResponse` | Pages Router types. App Router uses `NextRequest` |

### "use client" / "use server" Fabrications

| ❌ Hallucination | ✅ Reality |
|---|---|
| `"use static"` directive | Does not exist |
| `"use edge"` directive | Does not exist |
| `"use cache"` directive | Does not exist (use `unstable_cache` or `cache` import) |
| `"use server"` on a React component | `"use server"` is for Server Action functions only |

### Import Traps

| ❌ Hallucination | ✅ Reality |
|---|---|
| `import { useRouter } from "next/router"` | `import { useRouter } from "next/navigation"` |
| `import { useActionState } from "react-dom"` | `import { useActionState } from "react"` |
| `import { Link } from "next/link"` | `import Link from "next/link"` (default export) |
| `import { headers, cookies } from "next/headers"` and not awaiting | `headers()` and `cookies()` return Promises in Next.js 15+ |

## API Layer Discipline

**One API client module; pages and components never call `fetch`
directly for backend data.** Raw fetches bypass typed responses, error
classes, and shared config — and they inevitably re-implement helper
functions that already exist unused in the client.

- Define every request as an exported, typed function in `lib/api.ts`
  (fetcher + request/response types from `types/`).
- Export the base URL constant from the client module. Grep before
  adding another `process.env.NEXT_PUBLIC_API_URL || "..."` — divergent
  fallbacks across files (3000 in one, 8081 in another) send users to
  different origins.
- Throw a typed error class (`ApiError` with `status`) from the client;
  callers use `error instanceof ApiError` — never structural checks
  like `"status" in error` with casts.
- Audit: `grep -rn "fetch(" src/app src/components` — every hit outside
  the API client needs a justification (streaming, third-party).

**Error state ≠ empty state.** A failed fetch must never render the
"nothing here yet!" empty state — catch failures separately and show an
error block (or throw to the segment's error boundary):

```tsx
// ❌ Backend down renders "No items yet. Create the first one!"
catch { return { items: [] }; }

// ✅ Distinguish failure from genuinely empty
catch { return { items: [], loadFailed: true }; }
```

## WebSocket / Effect Lifecycle

**The disposed-flag pattern is mandatory for WebSockets with reconnect
logic.** The browser fires `onclose` AFTER your cleanup calls
`ws.close()` — an unguarded `onclose` schedules a reconnect that
re-creates sockets and intervals which outlive the component (a leak
that also explains "StrictMode double-connect" symptoms).

```tsx
useEffect(() => {
  let disposed = false;
  let ws: WebSocket | null = null;
  let reconnectTimer: ReturnType<typeof setTimeout>;
  let attempt = 0;

  function connect() {
    ws = new WebSocket(url);
    ws.onopen = () => { attempt = 0; };
    ws.onclose = () => {
      if (disposed) return;                       // ← the fix
      if (attempt < 10) {
        const delay = Math.min(1000 * 2 ** attempt, 30_000);
        attempt += 1;
        reconnectTimer = setTimeout(connect, delay); // backoff + cap
      }
    };
    ws.onerror = () => ws?.close();
  }
  connect();

  return () => {
    disposed = true;              // 1. mark disposed FIRST
    clearTimeout(reconnectTimer); // 2. kill pending reconnects
    if (ws) {                     // 3. detach handlers, then close
      ws.onclose = null;
      ws.onerror = null;
      ws.onmessage = null;
      ws.close();
    }
  };
}, [url]);
```

Rules that fall out of this:
- Reconnects always use exponential backoff with a max-attempt cap —
  a fixed short delay hammers a downed backend forever.
- WS URLs derive from env config (`NEXT_PUBLIC_WS_URL`), never a
  hardcoded `ws://localhost:...` — that breaks every non-local deploy
  and is mixed-content-blocked under https.
- Any interval/timeout created inside a socket handler must be cleared
  in the same cleanup.

## Fetch Races (AbortController)

Any user-triggered fetch that can overlap (search-as-you-type, submit +
load-more, filter changes) must abort the previous request or stale
responses will overwrite fresh ones:

```tsx
const abortRef = useRef<AbortController | null>(null);

async function performSearch(q: string) {
  abortRef.current?.abort();
  const controller = new AbortController();
  abortRef.current = controller;
  try {
    const result = await searchItems({ query: q }, { signal: controller.signal });
    setState(/* ... */);
  } catch (err) {
    if (err instanceof DOMException && err.name === "AbortError") return; // superseded
    setState({ status: "error", query: q });
  }
}
```

Companion rules:
- **Pagination uses the executed query, not the live input.** Store the
  query that produced the current results in the results state; "load
  more" reads that, never the input box value (typing without submitting
  must not mix result sets).
- Auto-search effects key off the actual URL param value — a one-shot
  `useRef(false)` guard breaks client-side `?q=a → ?q=b` navigation.
- Pass `{ signal }` through your API-client functions as an optional
  options param.

## Component Patterns & Anti-Patterns

### DO — Extract duplicated components and helpers

If you find the same component/function defined in 2+ files, extract to
a shared file (`lib/format.ts`, `components/ui/`). Common culprits:
- `Spinner` / `LoadingSpinner`
- `formatCount` / `formatNumber` / `formatDuration` / date formatting
- `EmptyState` with icon + message
- Skeleton markup duplicated between `loading.tsx` and in-component
  loading states

Duplicated helpers WILL drift (one file's `1.2k` becomes another's
`1.2K`) — that is a visible UI inconsistency, not just tech debt.
Before writing a formatting helper: `grep -rn "function format" src/`.

```tsx
// ❌ Wrong — duplicated in 3 files
function Spinner() { return <svg className="animate-spin" ... />; }

// ✅ Right — shared component
// components/ui/Spinner.tsx
export function Spinner() { return <svg className="animate-spin" ... />; }
```

### DO — img onError fallbacks

Every `<img>` whose `src` can 404 (avatars from third parties,
generated thumbnails) needs an `onError` fallback to an inline data-URI
placeholder — otherwise users see the broken-image icon. Keep the
fallbacks in one module (`lib/fallbacks.ts`):

```tsx
<img
  src={avatarUrl}
  onError={(e) => {
    const target = e.target as HTMLImageElement;
    target.onerror = null; // prevent loops if the fallback also errors
    target.src = AVATAR_FALLBACK;
  }}
/>
```

Note: an `onError` handler makes the component a Client Component —
say so in the `"use client"` comment (not a made-up reason like
"needed for Link navigation"; `<Link>` works in Server Components).

### DO — Derive, don't duplicate, state

State that is computable from other state is a bug factory — compute it
during render instead:

```tsx
// ❌ charCount can drift from message
const [message, setMessage] = useState("");
const [charCount, setCharCount] = useState(0);

// ✅ derived
const [message, setMessage] = useState("");
const charCount = message.length;
```

### DO NOT — Module-level mutable state

Module-level variables (`let userPromise: Promise<User> | null = null`) are
shared across all SSR requests and concurrent renders. During HMR or
concurrent rendering, one request can read another's data.

```tsx
// ❌ Wrong — shared across all renders
let userPromise: Promise<User> | null = null;
function fetchUser() {
  if (!userPromise) userPromise = getMe();
  return userPromise;
}

// ✅ Right — per-component state or React.cache()
function Component() {
  const [user, setUser] = useState<User | null>(null);
  useEffect(() => { getMe().then(setUser); }, []);
  // ...
}

// ✅ Also right — Server Component data fetching
import { cache } from "react";
const getUser = cache(async () => { /* fetch from DB */ });
```

### DO NOT — Inline style objects when Tailwind classes exist

Tailwind v4 with `@theme inline` maps CSS custom properties to utility
classes. Prefer classes over inline `style={{}}`:

```tsx
// ❌ Wrong — runtime style object
<div style={{ backgroundColor: "var(--color-surface-raised)" }}>

// ✅ Right — Tailwind class
<div className="bg-surface-raised">
```

### DO NOT — Use viewport units for component-internal layout changes

Components should not use `vw`, `vh`, `w-screen`, or `h-screen` to
implement features that affect page layout (theater mode, expanded view).
These units ignore the parent container and page structure. Instead:

- Lift the state to the parent component via a callback prop
- Let the parent decide how to rearrange the layout
- Use `%`, `flex`, or `max-w-*` within the component to fill available space

```tsx
// ❌ Wrong — component forces viewport width, breaks all layouts
<div className={isExpanded ? "!w-screen" : ""}>

// ✅ Right — component expands within parent, parent controls layout
<div className={isExpanded ? "!max-w-full" : ""}>
// Parent: <div className={isTheater ? "flex-col" : "lg:flex-row"}>
```

### DO — Stop both onClick and onDoubleClick on controls inside clickable parents

When a container has `onClick` or `onDoubleClick` handlers, every interactive
child (button, input, slider) must stop propagation for BOTH event types.
Stopping only `onClick` still allows `onDoubleClick` to bubble up to the
parent and trigger unwanted behavior.

```tsx
// ❌ Wrong — double-click on button triggers parent's onDoubleClick
<button onClick={(e) => e.stopPropagation()}>

// ✅ Right — both event types are blocked
<button
  onClick={(e) => e.stopPropagation()}
  onDoubleClick={(e) => e.stopPropagation()}
>
```

### Testing — vi.mock and class exports

When a module exports both functions AND classes (an API client with an
`ApiError` class), a factory mock that only stubs the functions makes
the class `undefined` — `instanceof` checks in the code under test then
throw or silently fail. Preserve real exports with `importOriginal`:

```ts
// ❌ ApiError becomes undefined; `err instanceof ApiError` breaks
vi.mock("@/lib/api", () => ({ getChannel: vi.fn() }));

// ✅ keep the real module, stub only what the test controls
vi.mock("@/lib/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/api")>();
  return { ...actual, getChannel: vi.fn() };
});
```

And reject with the REAL error class in tests
(`new ApiError(404, {...})`), not a structurally-similar plain object.

### Pre-Deploy Checklist

Before opening a PR for a frontend component:
- [ ] No `"use client"` unless the component uses hooks, event handlers, browser APIs, or img onError — and the comment states the REAL reason
- [ ] Server Components fetch data, Client Components receive it via props
- [ ] Each component has: loading state (skeleton), empty state, error state — and error is NOT rendered as empty
- [ ] Every server-fetching route segment has `error.tsx` + `loading.tsx` (+ `not-found.tsx` if it 404s); root `error.tsx`/`not-found.tsx` exist
- [ ] All backend calls go through the shared API client; no raw `fetch(` in pages/components
- [ ] WebSockets/effects use the disposed-flag cleanup; reconnects have backoff + cap
- [ ] Overlapping fetches abort stale requests (AbortController)
- [ ] `useSearchParams()` callers are wrapped in `<Suspense>`
- [ ] Internal navigation uses `<Link>`, not `<a href>`
- [ ] No module-level mutable variables — use `useState` + `useEffect` or React `cache()`
- [ ] No duplicated utility functions or UI components — extract to shared files
- [ ] No hardcoded origins — URLs derive from the shared base constant / env vars
- [ ] `npx tsc --noEmit` passes
- [ ] `npm run lint` passes (with `--max-warnings 0` — CI enforces this)
- [ ] `npm run build` passes (catches Suspense/prerender errors tsc misses)
- [ ] At minimum a render test for each component
- [ ] **Test-first:** the render test for each distinct state (loading,
  empty, error, populated) is written and shown failing BEFORE the
  component renders that state; bug fixes get a failing regression test
  first (AGENTS.md §5.2, spec-driven skill Phase 4)

### React Patterns — `useRef` vs `useState`

When a boolean flag exists only to prevent re-execution of a side effect
(not to drive rendering), use `useRef` instead of `useState`. Refs don't
trigger re-renders and don't need to be in dependency arrays:

```tsx
// ❌ Triggers re-render → re-creates useCallback → re-runs useEffect
const [didRun, setDidRun] = useState(false);
const fetchData = useCallback(() => {
  if (!didRun) { setDidRun(true); doWork(); }
}, [didRun]); // dep changes on first run → double-fetch

// ✅ No re-render, no dep array change
const hasRunRef = useRef(false);
const fetchData = useCallback(() => {
  if (!hasRunRef.current) { hasRunRef.current = true; doWork(); }
}, []); // stable
```

### ESLint — underscore-prefix does NOT suppress unused-vars

ESLint's `@typescript-eslint/no-unused-vars` does **not** treat `_` prefix
as "intentionally unused" by default (unlike Go or Python).

```tsx
// ❌ Still triggers @typescript-eslint/no-unused-vars
const Mock = ({ _unused, used }: Props) => <div>{used}</div>;

// ✅ Use the full props object to avoid the warning
const Mock = (props: Props) => <div>{props.used}</div>;
```

Alternatively, configure `argsIgnorePattern: "^_"` in `eslint.config` to
enable underscore-prefix suppression globally.

*Last updated: 2026-08-12*

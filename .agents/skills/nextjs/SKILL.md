---
name: nextjs
description: Next.js App Router development standards — Server Components, data fetching, route handlers, caching, layout patterns, and platform-specific AI traps. Load when writing TypeScript/React code in the frontend/ directory.
---

# Next.js App Router Standards

You are writing Next.js code using the App Router with Server Components
as the default. Follow these conventions exactly.

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

| ❌ DeepSeek says | ✅ Reality |
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

| ❌ DeepSeek says | ✅ Reality |
|---|---|
| `"use static"` directive | Does not exist |
| `"use edge"` directive | Does not exist |
| `"use cache"` directive | Does not exist (use `unstable_cache` or `cache` import) |
| `"use server"` on a React component | `"use server"` is for Server Action functions only |

### Import Traps

| ❌ DeepSeek says | ✅ Reality |
|---|---|
| `import { useRouter } from "next/router"` | `import { useRouter } from "next/navigation"` |
| `import { useActionState } from "react-dom"` | `import { useActionState } from "react"` |
| `import { Link } from "next/link"` | `import Link from "next/link"` (default export) |
| `import { headers, cookies } from "next/headers"` and not awaiting | `headers()` and `cookies()` return Promises in Next.js 15+ |

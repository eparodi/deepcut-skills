# Next.js App Router Reference — Traps, Mistakes & Best Practices

> **Audience:** AI coding agents (DeepSeek) and developers writing production Next.js App Router applications.
> **Versions:** Next.js 15.x, React 19, TypeScript 5.x

---

## Table of Contents

1. [Server Components Are DEFAULT](#1-server-components-are-default)
2. [Data Fetching Patterns](#2-data-fetching-patterns)
3. [Route Handlers (app/api)](#3-route-handlers-appapi)
4. [Layout / Page / Template Hierarchy](#4-layout--page--template-hierarchy)
5. [Caching Traps](#5-caching-traps)
6. [Metadata](#6-metadata)
7. [Testing](#7-testing)
8. [Monorepo File Structure](#8-monorepo-file-structure)
9. [Common AI Hallucinations](#9-common-ai-hallucinations)
10. [Middleware](#10-middleware)

---

## 1. Server Components Are DEFAULT

**Every component in the App Router is a Server Component unless it has `"use client"` at the top of the file.** No exceptions. This is the single biggest paradigm shift from Pages Router, and DeepSeek gets it wrong constantly.

### DO: Know when `"use client"` is REQUIRED

You MUST add `"use client"` when the component uses:

| Trigger | Concrete Example |
|---------|-----------------|
| `useState` | Toggle, form input state, modal open/close |
| `useEffect` | Side effects, subscriptions, browser APIs |
| `useContext` / `createContext` | Theme, auth, any React context |
| Event handlers (`onClick`, `onChange`, etc.) | Buttons, forms, interactive elements |
| Custom hooks using any of the above | `useLocalStorage`, `useMediaQuery`, `useDebounce` |
| Browser-only APIs | `window`, `document`, `localStorage`, `navigator`, `location` |
| `useRef` | DOM refs, mutable refs for intervals/timeouts |
| `useReducer` | Complex state machines |
| `useSyncExternalStore` | External store subscriptions |

### DO NOT: Add `"use client"` when it's NOT needed

| ❌ Wrong | ✅ Right |
|----------|----------|
| Adding `"use client"` to a component that only fetches data with `async/await` | Keep it as a Server Component — it runs on the server |
| Adding `"use client"` because the component receives `children` | Server Components can render `children` just fine |
| Adding `"use client"` to use `async` in the component | Server Components can be `async function` directly |
| Adding `"use client"` to pass props down | Props pass seamlessly from Server → Client components |

### DO: Server Component (async, data fetching, NO "use client")

```tsx
// app/users/page.tsx
import { db } from "@/lib/db";
import { UserCard } from "./user-card"; // client component

export default async function UsersPage() {
  // ✅ async/await directly in the component — server-only
  const users = await db.user.findMany({ take: 50 });

  return (
    <div className="grid grid-cols-3 gap-4">
      {users.map((user) => (
        <UserCard key={user.id} user={user} />
      ))}
    </div>
  );
}
```

### DO: Client Component ("use client" at the very top)

```tsx
// app/users/user-card.tsx
"use client";

import { useState } from "react";
import type { User } from "@/lib/db/schema";

export function UserCard({ user }: { user: User }) {
  // ✅ useState is fine because this is a Client Component
  const [expanded, setExpanded] = useState(false);

  return (
    <div className="rounded border p-4">
      <h3>{user.name}</h3>
      <button onClick={() => setExpanded(!expanded)}>
        {expanded ? "Collapse" : "Expand"}
      </button>
      {expanded && <p>{user.bio}</p>}
    </div>
  );
}
```

### DO NOT: This is what DeepSeek hallucinates (MIXING patterns)

```tsx
// ❌ DEEPSEEK HALLUCINATION — "use client" AND async/await in the SAME file
"use client";

import { db } from "@/lib/db";

export default async function UsersPage() { // ❌ Client Components CANNOT be async
  const users = await db.user.findMany();    // ❌ This won't work in a Client Component
  // ...render...
}
```

**The fix:** Remove `"use client"` and make it a Server Component. Or fetch data in a parent Server Component and pass it as props.

### DO NOT: Using hooks in a Server Component

```tsx
// ❌ DEEPSEEK HALLUCINATION — hooks in Server Component
import { useState, useEffect } from "react";

export default function Page() {
  const [count, setCount] = useState(0);       // ❌ Runtime error
  useEffect(() => { /* ... */ }, []);            // ❌ Runtime error
  return <button onClick={() => setCount(c => c + 1)}>Click</button>; // ❌ onClick needs "use client"
}
```

**The fix:** Extract the interactive part into a Client Component:

```tsx
// app/page.tsx — Server Component
import { Counter } from "./counter";

export default function Page() {
  return (
    <main>
      <h1>Dashboard</h1>
      <Counter /> {/* ✅ Client Component rendered inside Server Component */}
    </main>
  );
}
```

```tsx
// app/counter.tsx — Client Component
"use client";

import { useState } from "react";

export function Counter() {
  const [count, setCount] = useState(0);
  return <button onClick={() => setCount((c) => c + 1)}>Count: {count}</button>;
}
```

### DO: Context providers MUST be in a Client Component

```tsx
// ❌ DEEPSEEK HALLUCINATION — createContext in a Server Component
import { createContext, useContext } from "react";

export const ThemeContext = createContext("light");

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <ThemeContext.Provider value="dark"> {/* ❌ Provider in Server Component */}
      {children}
    </ThemeContext.Provider>
  );
}
```

**The fix:** Extract the provider into its own Client Component:

```tsx
// app/providers.tsx
"use client";

import { createContext, useContext, useState, type ReactNode } from "react";

const ThemeContext = createContext<{
  theme: "light" | "dark";
  toggle: () => void;
} | null>(null);

export function useTheme() {
  const ctx = useContext(ThemeContext);
  if (!ctx) throw new Error("useTheme must be used within Providers");
  return ctx;
}

export function Providers({ children }: { children: ReactNode }) {
  const [theme, setTheme] = useState<"light" | "dark">("light");
  const toggle = () => setTheme((t) => (t === "light" ? "dark" : "light"));

  return (
    <ThemeContext.Provider value={{ theme, toggle }}>
      {children}
    </ThemeContext.Provider>
  );
}
```

```tsx
// app/layout.tsx — Server Component
import { Providers } from "./providers";

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <Providers>{children}</Providers> {/* ✅ Client Provider wraps children */}
      </body>
    </html>
  );
}
```

### DO: "use client" boundary is a FILE-level directive

- `"use client"` marks the ENTIRE file and ALL its imports as client-rendered.
- A Client Component can import Server Components only as `children` or props passed from a Server Component.
- Once you cross the boundary INTO a Client Component, everything below is client-rendered.

### DO NOT: DeepSeek invents fake directives

```tsx
// ❌ HALLUCINATION — these DO NOT EXIST
"use server";   // ❌ Not for components (only for Server Actions in a separate file/module)
"use static";   // ❌ Does not exist
"use streaming"; // ❌ Does not exist
"use edge";     // ❌ Does not exist (use `export const runtime = "edge"` instead)
```

### DO: Marking a Server Action module

```tsx
// app/actions.ts
"use server"; // ✅ Valid — marks all exports as Server Actions

import { revalidatePath } from "next/cache";
import { db } from "@/lib/db";

export async function deleteUser(id: string) {
  await db.user.delete({ where: { id } });
  revalidatePath("/users");
}
```

---

## 2. Data Fetching Patterns

### DO: Use the right tool for the job

| Task | Pattern | Location |
|------|---------|----------|
| Read data for a page | `async` Server Component | `app/**/page.tsx` |
| Mutate data | Server Action or Route Handler | `app/actions.ts` or `app/api/**/route.ts` |
| External webhooks / API | Route Handler | `app/api/**/route.ts` |
| Revalidate after mutation | `revalidatePath` / `revalidateTag` | Inside Server Action or Route Handler |
| Optimistic UI | `useOptimistic` + Server Action | Client Component calling a Server Action |

### DO: Fetch data in an async Server Component (READ operations)

```tsx
// app/posts/page.tsx
import { db } from "@/lib/db";
import type { Post } from "@/lib/db/schema";

export default async function PostsPage() {
  const posts: Post[] = await db.post.findMany({
    orderBy: { createdAt: "desc" },
  });

  if (posts.length === 0) {
    return <p>No posts yet.</p>;
  }

  return (
    <ul>
      {posts.map((post) => (
        <li key={post.id}>
          <h2>{post.title}</h2>
          <p>{post.excerpt}</p>
        </li>
      ))}
    </ul>
  );
}
```

### DO: Server Actions for mutations (Next.js 14+ recommended pattern)

```tsx
// app/actions/posts.ts
"use server";

import { revalidatePath } from "next/cache";
import { db } from "@/lib/db";
import { z } from "zod";

const CreatePostSchema = z.object({
  title: z.string().min(1).max(200),
  content: z.string().min(1),
});

export async function createPost(formData: FormData) {
  const parsed = CreatePostSchema.safeParse({
    title: formData.get("title"),
    content: formData.get("content"),
  });

  if (!parsed.success) {
    return { error: parsed.error.flatten().fieldErrors };
  }

  await db.post.create({ data: parsed.data });
  revalidatePath("/posts"); // ✅ revalidate the page cache after mutation
  return { success: true };
}
```

```tsx
// app/posts/new/page.tsx — uses useActionState (Next.js 15 / React 19)
"use client";

import { useActionState } from "react";
import { createPost } from "@/app/actions/posts";

export default function NewPostPage() {
  const [state, formAction, isPending] = useActionState(createPost, null);

  return (
    <form action={formAction}>
      <input name="title" placeholder="Title" required />
      {state?.error?.title && <p className="text-red-500">{state.error.title[0]}</p>}

      <textarea name="content" placeholder="Content" required />
      {state?.error?.content && <p className="text-red-500">{state.error.content[0]}</p>}

      <button type="submit" disabled={isPending}>
        {isPending ? "Creating..." : "Create Post"}
      </button>

      {state?.success && <p className="text-green-500">Post created!</p>}
    </form>
  );
}
```

### DO: Route Handlers for external-facing APIs or webhooks

```tsx
// app/api/posts/route.ts
import { NextRequest, NextResponse } from "next/server";
import { db } from "@/lib/db";

export async function GET(request: NextRequest) {
  const { searchParams } = request.nextUrl;
  const page = parseInt(searchParams.get("page") ?? "1");

  const posts = await db.post.findMany({
    take: 10,
    skip: (page - 1) * 10,
    orderBy: { createdAt: "desc" },
  });

  return NextResponse.json({ posts, page });
}

export async function POST(request: NextRequest) {
  const body = await request.json();
  const post = await db.post.create({ data: body });
  return NextResponse.json(post, { status: 201 });
}
```

### DO: revalidatePath vs revalidateTag

```tsx
// With tags — more surgical
import { unstable_cache } from "next/cache";

const getPost = unstable_cache(
  async (id: string) => db.post.findUnique({ where: { id } }),
  ["post"],          // cache key parts
  { tags: [`post-${id}`] } // tags for targeted revalidation
);

// Later, after mutation:
import { revalidateTag } from "next/cache";
revalidateTag(`post-${postId}`); // ✅ only this one post's cache is cleared
```

```tsx
// With paths — broader
import { revalidatePath } from "next/cache";
revalidatePath("/posts");     // revalidates /posts (the page)
revalidatePath("/posts/[id]"); // revalidates all dynamic post pages
revalidatePath("/", "layout"); // revalidates layout + all nested pages
```

### DO NOT: Attempt `fetch` inside a Client Component for page data

```tsx
// ❌ DEEPSEEK HALLUCINATION — fetch in Client Component + useState
"use client";

import { useState, useEffect } from "react";

export default function PostsPage() {
  const [posts, setPosts] = useState([]);

  useEffect(() => {
    fetch("/api/posts").then((res) => res.json()).then(setPosts);
  }, []);

  return <ul>{posts.map(/* ... */)}</ul>;
}
```

**The fix:** Make this a Server Component and fetch directly, or use a data-fetching library like TanStack Query if the data must be client-side (e.g., real-time polling).

### DO NOT: Call `fetch` with relative URLs in a Server Component

```tsx
// ❌ DEEPSEEK HALLUCINATION — relative URL in Server Component fetch
export default async function Page() {
  const data = await fetch("/api/users").then(r => r.json()); // ❌ Server has no origin
}
```

**The fix:** Server Components should call the database/ORM directly, NOT fetch an internal API route. If you must call the API route, construct an absolute URL:

```tsx
import { headers } from "next/headers";

export default async function Page() {
  const headersList = await headers();
  const host = headersList.get("host") ?? "localhost:3000";
  const protocol = process.env.NODE_ENV === "development" ? "http" : "https";
  const data = await fetch(`${protocol}://${host}/api/users`).then((r) => r.json());
}
```

But again: **prefer calling the database directly in Server Components.**

### DO NOT: Confuse `useActionState` import location

```tsx
// ✅ React 19 (Next.js 15) — correct import
import { useActionState } from "react";

// ❌ DEEPSEEK HALLUCINATION — old import paths that don't exist anymore
import { useActionState } from "react-dom";         // ❌ Wrong in React 19
import { experimental_useFormState } from "react-dom"; // ❌ Removed in React 19
```

---

## 3. Route Handlers (app/api)

### DO: Standard CRUD pattern with correct exports

```tsx
// app/api/users/route.ts
import { NextRequest, NextResponse } from "next/server";
import { db } from "@/lib/db";

export async function GET(request: NextRequest) {
  const { searchParams } = request.nextUrl;
  const q = searchParams.get("q");

  const users = q
    ? await db.user.findMany({ where: { name: { contains: q } } })
    : await db.user.findMany();

  return NextResponse.json(users);
}

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const user = await db.user.create({ data: body });
    return NextResponse.json(user, { status: 201 });
  } catch (error) {
    return NextResponse.json(
      { error: "Failed to create user" },
      { status: 500 }
    );
  }
}
```

```tsx
// app/api/users/[id]/route.ts
import { NextRequest, NextResponse } from "next/server";
import { db } from "@/lib/db";

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params; // ✅ In Next.js 15, params is a Promise
  const user = await db.user.findUnique({ where: { id } });

  if (!user) {
    return NextResponse.json({ error: "Not found" }, { status: 404 });
  }

  return NextResponse.json(user);
}

export async function PATCH(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const body = await request.json();
  const user = await db.user.update({ where: { id }, data: body });
  return NextResponse.json(user);
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  await db.user.delete({ where: { id } });
  return NextResponse.json(null, { status: 204 });
}
```

### DO NOT: DeepSeek hallucinates `params` as synchronous

```tsx
// ❌ DEEPSEEK HALLUCINATION — params is NOT synchronous in Next.js 15
export async function GET(
  request: NextRequest,
  { params }: { params: { id: string } }  // ❌ Wrong type — params is Promise<{ id: string }>
) {
  const { id } = params; // ❌ TypeScript error in Next.js 15
}
```

### DO: Cookie handling

```tsx
import { NextRequest, NextResponse } from "next/server";

export async function POST(request: NextRequest) {
  // Read cookies
  const token = request.cookies.get("session-token")?.value;
  const allCookies = request.cookies.getAll();

  const response = NextResponse.json({ ok: true });

  // Set cookies
  response.cookies.set("session-token", "abc123", {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    maxAge: 60 * 60 * 24 * 7, // 7 days
    path: "/",
  });

  // Delete cookies
  response.cookies.delete("old-cookie");

  return response;
}
```

### DO: Reading headers

```tsx
import { NextRequest, NextResponse } from "next/server";

export async function GET(request: NextRequest) {
  const userAgent = request.headers.get("user-agent");
  const authorization = request.headers.get("authorization");
  const contentType = request.headers.get("content-type");

  if (!authorization?.startsWith("Bearer ")) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  const token = authorization.slice(7);
  // verify token...

  return NextResponse.json({ userAgent, authenticated: true });
}
```

### DO: Request body parsing — always handle errors

```tsx
export async function POST(request: NextRequest) {
  // ❌ DEEPSEEK HALLUCINATION — doesn't handle parse errors
  // const body = await request.json();

  // ✅ Always wrap in try/catch
  let body: unknown;
  try {
    body = await request.json();
  } catch {
    return NextResponse.json(
      { error: "Invalid JSON body" },
      { status: 400 }
    );
  }

  // ✅ Validate with zod after parsing
  const parsed = CreateUserSchema.safeParse(body);
  if (!parsed.success) {
    return NextResponse.json(
      { error: parsed.error.flatten() },
      { status: 422 }
    );
  }

  const user = await db.user.create({ data: parsed.data });
  return NextResponse.json(user, { status: 201 });
}
```

### DO NOT: Importing the wrong `NextResponse`

```tsx
// ✅ Correct
import { NextResponse } from "next/server";

// ❌ HALLUCINATION — does not exist
import { Response } from "next/server";          // ❌
import { NextAPIResponse } from "next/server";   // ❌
```

### DO NOT: Common security mistakes in Route Handlers

```tsx
// ❌ SECURITY MISTAKE — no input validation
export async function POST(request: NextRequest) {
  const { username, role } = await request.json();
  // ❌ Directly spreading user input into DB — mass assignment vulnerability
  const user = await db.user.create({ data: { username, role } });
  return NextResponse.json(user);
}

// ✅ Use zod + explicit field picking
const CreateUserSchema = z.object({
  username: z.string().min(3).max(50),
  // NOTE: role is NOT in the schema — cannot be set by user
});

export async function POST(request: NextRequest) {
  const body = await request.json();
  const parsed = CreateUserSchema.safeParse(body);
  if (!parsed.success) {
    return NextResponse.json({ error: parsed.error.flatten() }, { status: 422 });
  }

  // ✅ Only specific fields, with server-enforced role
  const user = await db.user.create({
    data: { ...parsed.data, role: "user" },
  });
  return NextResponse.json(user, { status: 201 });
}
```

### DO: CORS in Route Handlers

```tsx
export async function GET(request: NextRequest) {
  const origin = request.headers.get("origin") ?? "";

  const allowedOrigins = [
    process.env.NEXT_PUBLIC_APP_URL ?? "http://localhost:3000",
  ];

  const isAllowed = allowedOrigins.includes(origin);

  return NextResponse.json(
    { data: "public" },
    {
      headers: {
        "Access-Control-Allow-Origin": isAllowed ? origin : allowedOrigins[0],
        "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
        "Access-Control-Allow-Headers": "Content-Type, Authorization",
      },
    }
  );
}

export async function OPTIONS() {
  return NextResponse.json(null, {
    headers: {
      "Access-Control-Allow-Origin": "*",
      "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
      "Access-Control-Allow-Headers": "Content-Type, Authorization",
    },
  });
}
```

---

## 4. Layout / Page / Template Hierarchy

### DO: Understand the rendering order

```
RootLayout (app/layout.tsx)
├── Nested Layout (app/dashboard/layout.tsx)
│   ├── Page (app/dashboard/page.tsx)
│   ├── template.tsx          ← re-mounts on every navigation
│   ├── loading.tsx           ← shown while page loads
│   ├── error.tsx             ← catches errors in children
│   └── not-found.tsx
```

### DO: Nested layouts for shared UI

```tsx
// app/dashboard/layout.tsx
import { Sidebar } from "./sidebar";
import { TopNav } from "./top-nav";
import { requireAuth } from "@/lib/auth";

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const user = await requireAuth(); // ✅ layouts CAN do async data fetching

  return (
    <div className="flex h-screen">
      <Sidebar user={user} />
      <div className="flex-1 flex flex-col">
        <TopNav user={user} />
        <main className="flex-1 p-6">{children}</main>
      </div>
    </div>
  );
}
```

### DO NOT: Use `searchParams` in layouts

```tsx
// ❌ DEEPSEEK HALLUCINATION — layouts cannot access searchParams
export default function DashboardLayout({
  children,
  searchParams, // ❌ Runtime error! layouts don't receive searchParams
}: {
  children: React.ReactNode;
  searchParams: { q: string }; // ❌ This prop does not exist on layouts
}) {
  return <div>{children}</div>;
}
```

**The fix:** Use `searchParams` in the `page.tsx` and pass data down, or use middleware to read search params.

### DO: Parallel Routes

```tsx
// app/dashboard/layout.tsx
export default function DashboardLayout({
  children,       // app/dashboard/page.tsx
  analytics,      // app/dashboard/@analytics/page.tsx
  team,           // app/dashboard/@team/page.tsx
}: {
  children: React.ReactNode;
  analytics: React.ReactNode;
  team: React.ReactNode;
}) {
  return (
    <div className="grid grid-cols-3 gap-4">
      <div>{team}</div>
      <div className="col-span-2">{children}</div>
      <div className="col-span-3">{analytics}</div>
    </div>
  );
}
```

File structure:
```
app/dashboard/
├── layout.tsx
├── page.tsx
├── @analytics/
│   ├── page.tsx
│   └── loading.tsx       # per-slot loading state
├── @team/
│   ├── page.tsx
│   └── default.tsx       # fallback for unmatched routes
```

### DO: Intercepting Routes

```
app/
├── photos/
│   ├── page.tsx           # /photos — gallery grid
│   └── [id]/
│       └── page.tsx       # /photos/[id] — full page view
├── @modal/
│   ├── default.tsx        # returns null when no modal
│   └── (.)photos/
│       └── [id]/
│           └── page.tsx   # intercepts /photos/[id] when navigated from /photos
```

Use `(.)` to match the same level, `(..)` for one level up, `(...)` for root.

### DO: Route Groups `(groupName)` — organize without affecting URL

```
app/
├── (marketing)/
│   ├── layout.tsx         # shared marketing layout
│   ├── page.tsx           # /
│   ├── pricing/
│   │   └── page.tsx       # /pricing
│   └── about/
│       └── page.tsx       # /about
├── (dashboard)/
│   ├── layout.tsx         # shared dashboard layout (auth required)
│   ├── page.tsx           # / (but rendered with dashboard layout)
│   └── settings/
│       └── page.tsx       # /settings
```

### DO: loading.tsx patterns

```tsx
// app/posts/loading.tsx — automatically wrapped around page.tsx
import { Skeleton } from "@/components/ui/skeleton";

export default function PostsLoading() {
  return (
    <div className="space-y-4">
      {Array.from({ length: 5 }).map((_, i) => (
        <div key={i} className="flex gap-4 p-4 border rounded animate-pulse">
          <Skeleton className="h-16 w-16 rounded-full" />
          <div className="flex-1 space-y-2">
            <Skeleton className="h-4 w-3/4" />
            <Skeleton className="h-3 w-full" />
            <Skeleton className="h-3 w-1/2" />
          </div>
        </div>
      ))}
    </div>
  );
}
```

### DO: error.tsx patterns

```tsx
// app/posts/error.tsx
"use client"; // ✅ MUST be a Client Component

import { useEffect } from "react";

export default function PostsError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error("Posts page error:", error);
  }, [error]);

  return (
    <div className="flex flex-col items-center justify-center py-20 gap-4">
      <h2 className="text-xl font-semibold">Something went wrong!</h2>
      <p className="text-muted-foreground">
        {error.message || "Failed to load posts."}
      </p>
      <button
        onClick={reset} // ✅ re-renders the page (retries)
        className="px-4 py-2 bg-primary text-primary-foreground rounded"
      >
        Try again
      </button>
    </div>
  );
}
```

### DO: not-found.tsx patterns

```tsx
// app/not-found.tsx — global 404
import Link from "next/link";

export default function NotFound() {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen gap-4">
      <h1 className="text-4xl font-bold">404</h1>
      <p>Page not found</p>
      <Link href="/" className="text-primary underline">
        Go home
      </Link>
    </div>
  );
}
```

```tsx
// app/users/[id]/page.tsx — trigger not-found per-entity
import { notFound } from "next/navigation";
import { db } from "@/lib/db";

export default async function UserPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const user = await db.user.findUnique({ where: { id } });

  if (!user) {
    notFound(); // ✅ triggers the closest not-found.tsx
  }

  return <h1>{user.name}</h1>;
}
```

### DO NOT: Forget `error.tsx` must be a Client Component

```tsx
// ❌ This will fail — error.tsx needs "use client"
export default function ErrorPage({
  error,
  reset,
}: {
  error: Error;
  reset: () => void;
}) {
  return <button onClick={reset}>Retry</button>; // ❌ onClick without "use client"
}
```

---

## 5. Caching Traps

**Next.js 15 changed the defaults dramatically.** This is where DeepSeek gets things wrong most often because its training data is saturated with Next.js 13/14 patterns.

### The Big Changes in Next.js 15

| What changed | Before (13/14) | Now (15) |
|-------------|----------------|----------|
| `fetch()` caching | Cached by default | **NOT cached** by default — `cache: "no-store"` is the new default |
| GET route handlers | Cached by default | **NOT cached** by default — dynamic by default |
| Pages using `cookies()` / `headers()` | Could still be static | **Always dynamic** — no opt-out |
| `fetchCache` config | `"default-cache"` | Removed entirely |

### DO: Explicitly opt INTO caching when you want it

```tsx
// ✅ Opt into data cache with fetch (Next.js 15)
const data = await fetch("https://api.example.com/data", {
  cache: "force-cache", // explicitly cached
});
```

```tsx
// ✅ Opt into data cache with unstable_cache
import { unstable_cache } from "next/cache";
import { db } from "@/lib/db";

const getExpensiveQuery = unstable_cache(
  async () => {
    return db.analytics.groupBy({ /* expensive aggregation */ });
  },
  ["expensive-analytics"],
  {
    revalidate: 3600, // revalidate every hour
    tags: ["analytics"],
  }
);

export default async function AnalyticsPage() {
  const data = await getExpensiveQuery();
  return <pre>{JSON.stringify(data, null, 2)}</pre>;
}
```

### DO: Understand the four caches

| Cache | What it stores | Where | Duration | How to clear |
|-------|---------------|-------|----------|-------------|
| **Data Cache** | `fetch()` results and `unstable_cache` | Server | Persistent (survives deploys) | `revalidatePath()`, `revalidateTag()` |
| **Full Route Cache** | Rendered HTML + RSC payload | Server | Persistent | `revalidatePath()`, `revalidateTag()`, `dynamicParams` |
| **Router Cache** | RSC payload in browser memory | Client | Session (30s for dynamic, 5min for static) | `router.refresh()`, `revalidatePath()` |
| **Request Memoization** | `fetch()` dedup within a single render | Server | Single render pass | N/A — automatic |

### DO NOT: Assume `fetch()` is cached

```tsx
// ❌ DEEPSEEK HALLUCINATION — assumes cached by default (was true in Next.js 13/14)
export default async function Page() {
  // In Next.js 15, this fetches on EVERY request
  const data = await fetch("https://api.example.com/data").then((r) => r.json());
  return <pre>{JSON.stringify(data)}</pre>;
}

// ✅ Explicit cache opt-in for Next.js 15
export default async function Page() {
  const data = await fetch("https://api.example.com/data", {
    cache: "force-cache",
  }).then((r) => r.json());
  return <pre>{JSON.stringify(data)}</pre>;
}
```

### DO: Use `dynamic` config correctly

```tsx
// ❌ DEEPSEEK HALLUCINATION
export const dynamic = "static";   // valid but rarely used in Next.js 15
export const dynamic = "auto";     // ❌ NOT a valid value — removed in 15

// ✅ Valid values
export const dynamic = "force-dynamic";  // always server-render on every request
export const dynamic = "force-static";   // force static rendering (build error if page uses dynamic APIs)
export const dynamic = "error";          // throw error if page uses dynamic APIs (for debugging)
```

### DO NOT: Mix `noStore()` with `unstable_cache` on the same data

```tsx
// ❌ DEEPSEEK HALLUCINATION — contradictory caching directives
import { unstable_noStore as noStore } from "next/cache";

const getData = unstable_cache(
  async () => {
    noStore(); // ❌ This cancels the cache you just set up!
    return db.query();
  },
  ["my-data"]
);
```

### DO: Revalidate patterns

```tsx
// Time-based revalidation (ISR-like)
const getPosts = unstable_cache(
  async () => db.post.findMany(),
  ["posts"],
  { revalidate: 60 } // revalidate at most every 60 seconds
);
```

```tsx
// On-demand revalidation (after mutation)
// In a Server Action:
"use server";

import { revalidatePath, revalidateTag } from "next/cache";
import { db } from "@/lib/db";

export async function updatePost(id: string, data: { title: string }) {
  await db.post.update({ where: { id }, data });

  revalidateTag(`post-${id}`);   // ✅ revalidate the cached post
  revalidatePath(`/posts/${id}`); // ✅ revalidate the post page
  revalidatePath("/posts");       // ✅ revalidate the listing page
}
```

### DO: `router.refresh()` in Client Components

```tsx
"use client";

import { useRouter } from "next/navigation";

export function RefreshButton() {
  const router = useRouter();

  return (
    <button onClick={() => router.refresh()}>
      Refresh Data
    </button>
  );
}
// ✅ Refetches Server Components without full page reload
// ✅ Clears Router Cache and re-renders server parts
```

---

## 6. Metadata

### DO: Static metadata

```tsx
// app/page.tsx
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Home | My App",
  description: "Welcome to My App",
};
```

### DO: Dynamic metadata with `generateMetadata`

```tsx
// app/posts/[id]/page.tsx
import type { Metadata } from "next";
import { db } from "@/lib/db";
import { notFound } from "next/navigation";

interface Props {
  params: Promise<{ id: string }>;
  searchParams: Promise<{ [key: string]: string | string[] | undefined }>;
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { id } = await params;
  const post = await db.post.findUnique({ where: { id } });

  if (!post) {
    return { title: "Post Not Found" };
  }

  return {
    title: post.title,
    description: post.excerpt,
    openGraph: {
      title: post.title,
      description: post.excerpt,
      type: "article",
      publishedTime: post.createdAt.toISOString(),
      images: post.coverImage ? [{ url: post.coverImage }] : [],
    },
    twitter: {
      card: "summary_large_image",
      title: post.title,
      description: post.excerpt,
      images: post.coverImage ? [post.coverImage] : [],
    },
  };
}

export default async function PostPage({ params }: Props) {
  const { id } = await params;
  const post = await db.post.findUnique({ where: { id } });

  if (!post) notFound();

  return (
    <article>
      <h1>{post.title}</h1>
      <div dangerouslySetInnerHTML={{ __html: post.content }} />
    </article>
  );
}
```

### DO: Template-based titles (inherit from parent)

```tsx
// app/layout.tsx
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: {
    template: "%s | My App",  // child pages fill in %s
    default: "My App",         // fallback for pages without title
  },
  description: "The best app ever",
};
```

```tsx
// app/about/page.tsx
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "About", // renders as "About | My App"
};
```

### DO NOT: Export `metadata` from a Client Component

```tsx
// ❌ DEEPSEEK HALLUCINATION
"use client";

import type { Metadata } from "next";

export const metadata: Metadata = {   // ❌ Runtime error — metadata only in Server Components
  title: "Client Page",
};

export default function ClientPage() {
  return <div>...</div>;
}
```

**The fix:** Move the metadata export to a `layout.tsx` or wrap in a Server Component.

### DO: `generateMetadata` runs on the server

- `generateMetadata` only executes on the server.
- It CAN access the database, filesystem, environment variables.
- It CANNOT use hooks, `useState`, `useEffect`, browser APIs.
- The `params` and `searchParams` are Promises in Next.js 15.

### DO NOT: Use `headers()` or `cookies()` in metadata without awaiting

```tsx
// ✅ Correct in Next.js 15
import { headers } from "next/headers";

export async function generateMetadata(): Promise<Metadata> {
  const headersList = await headers(); // ✅ Must await
  const pathname = headersList.get("x-pathname");
  return { title: pathname ?? "Default" };
}
```

### DO NOT: DeepSeek hallucinates `generateMetadata` signature

```tsx
// ❌ HALLUCINATION
export async function generateMetadata({ params: { id } }: { params: { id: string } }) {
  // ❌ params is not a Promise — TypeScript error in Next.js 15
}

// ❌ HALLUCINATION
export const generateMetadata = async (context) => { // ❌ no types
  return {};
};

// ✅ CORRECT Next.js 15
export async function generateMetadata({
  params,
}: {
  params: Promise<{ id: string }>;
}): Promise<Metadata> {
  const { id } = await params;
  // ...
}
```

---

## 7. Testing

### DO: Server Component testing with `render` (async)

```tsx
// __tests__/posts-page.test.tsx
import { render, screen } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import PostsPage from "@/app/posts/page";

// Mock the database
vi.mock("@/lib/db", () => ({
  db: {
    post: {
      findMany: vi.fn(),
    },
  },
}));

import { db } from "@/lib/db";

describe("PostsPage", () => {
  it("renders posts from database", async () => {
    const mockPosts = [
      { id: "1", title: "Post One", excerpt: "First post", createdAt: new Date() },
      { id: "2", title: "Post Two", excerpt: "Second post", createdAt: new Date() },
    ];

    vi.mocked(db.post.findMany).mockResolvedValue(mockPosts);

    // ✅ Server Components are async — render them as async
    const jsx = await PostsPage();
    render(jsx);

    expect(screen.getByText("Post One")).toBeDefined();
    expect(screen.getByText("Post Two")).toBeDefined();
  });

  it("shows empty state when no posts", async () => {
    vi.mocked(db.post.findMany).mockResolvedValue([]);

    const jsx = await PostsPage();
    render(jsx);

    expect(screen.getByText("No posts yet.")).toBeDefined();
  });
});
```

### DO: Testing Client Components with React Testing Library

```tsx
// __tests__/counter.test.tsx
import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { Counter } from "@/app/counter";

describe("Counter", () => {
  it("increments count on button click", () => {
    render(<Counter />);

    const button = screen.getByRole("button");
    expect(button).toHaveTextContent("Count: 0");

    fireEvent.click(button);
    expect(button).toHaveTextContent("Count: 1");

    fireEvent.click(button);
    expect(button).toHaveTextContent("Count: 2");
  });
});
```

### DO: Route Handler testing

```tsx
// __tests__/api-users.test.ts
import { describe, it, expect, vi, beforeEach } from "vitest";
import { GET, POST } from "@/app/api/users/route";
import { db } from "@/lib/db";
import { NextRequest } from "next/server";

vi.mock("@/lib/db", () => ({
  db: {
    user: {
      findMany: vi.fn(),
      create: vi.fn(),
    },
  },
}));

function createRequest(url: string, init?: RequestInit) {
  return new NextRequest(new URL(url, "http://localhost:3000"), init);
}

describe("GET /api/users", () => {
  it("returns users", async () => {
    const mockUsers = [{ id: "1", name: "Alice" }];
    vi.mocked(db.user.findMany).mockResolvedValue(mockUsers);

    const request = createRequest("http://localhost:3000/api/users");
    const response = await GET(request);

    expect(response.status).toBe(200);
    const body = await response.json();
    expect(body).toEqual(mockUsers);
  });
});

describe("POST /api/users", () => {
  it("creates a user", async () => {
    const mockUser = { id: "1", name: "Bob", email: "bob@test.com" };
    vi.mocked(db.user.create).mockResolvedValue(mockUser);

    const request = createRequest("http://localhost:3000/api/users", {
      method: "POST",
      body: JSON.stringify({ name: "Bob", email: "bob@test.com" }),
    });

    const response = await POST(request);
    expect(response.status).toBe(201);
    const body = await response.json();
    expect(body).toEqual(mockUser);
  });

  it("returns 422 for invalid data", async () => {
    const request = createRequest("http://localhost:3000/api/users", {
      method: "POST",
      body: JSON.stringify({ name: "" }), // invalid
    });

    const response = await POST(request);
    expect(response.status).toBe(422);
  });
});
```

### DO: Playwright E2E test

```tsx
// e2e/posts.spec.ts
import { test, expect } from "@playwright/test";

test.describe("Posts page", () => {
  test("displays posts and navigates to detail", async ({ page }) => {
    await page.goto("/posts");

    // Wait for posts to load
    await expect(page.getByRole("list")).toBeVisible();
    const posts = page.getByRole("listitem");
    await expect(posts.first()).toBeVisible();

    // Click first post
    await posts.first().getByRole("link").click();

    // Should be on the detail page
    await expect(page).toHaveURL(/\/posts\/.+/);
    await expect(page.getByRole("heading").first()).toBeVisible();
  });

  test("shows loading state", async ({ page }) => {
    // Simulate slow network
    await page.route("**/api/posts**", async (route) => {
      await new Promise((r) => setTimeout(r, 2000));
      await route.continue();
    });

    await page.goto("/posts");
    // Check for loading skeleton
    await expect(page.locator(".animate-pulse")).toBeVisible();
  });
});
```

### DO NOT: Try to render a Server Component that uses `headers()` or `cookies()` without mock setup

```tsx
// ❌ Will fail — headers() not available in test environment
import { render } from "@testing-library/react";
import DashboardPage from "@/app/dashboard/page";

test("dashboard", async () => {
  const jsx = await DashboardPage(); // ❌ throws: headers() used outside request context
  render(jsx);
});
```

**The fix:** Extract the data-fetching call into a separate function that can be mocked:

```tsx
// app/dashboard/page.tsx
import { getDashboardData } from "@/lib/queries";

export default async function DashboardPage() {
  const data = await getDashboardData(); // ✅ mockable in tests
  return <div>{data.title}</div>;
}
```

---

## 8. Monorepo File Structure

### DO: Standard Next.js monorepo layout

```
project-root/
├── frontend/                          # Next.js App Router
│   ├── app/
│   │   ├── layout.tsx                 # Root layout
│   │   ├── page.tsx                   # Home page
│   │   ├── not-found.tsx              # Global 404
│   │   ├── error.tsx                  # Global error boundary
│   │   ├── loading.tsx                # Global loading state
│   │   ├── globals.css                # Tailwind / global styles
│   │   ├── (marketing)/               # Route group — public pages
│   │   │   ├── layout.tsx
│   │   │   ├── page.tsx               # /
│   │   │   ├── pricing/
│   │   │   │   └── page.tsx
│   │   │   └── blog/
│   │   │       └── [slug]/
│   │   │           └── page.tsx
│   │   ├── (dashboard)/               # Route group — authenticated pages
│   │   │   ├── layout.tsx             # auth-checking layout
│   │   │   ├── page.tsx               # /dashboard
│   │   │   ├── settings/
│   │   │   │   └── page.tsx
│   │   │   └── @modal/                # parallel route for modals
│   │   │       └── default.tsx
│   │   ├── api/                       # Route Handlers
│   │   │   ├── auth/
│   │   │   │   └── [...nextauth]/     # if using NextAuth/Auth.js
│   │   │   │       └── route.ts
│   │   │   ├── users/
│   │   │   │   ├── route.ts
│   │   │   │   └── [id]/
│   │   │   │       └── route.ts
│   │   │   └── webhooks/
│   │   │       └── stripe/
│   │   │           └── route.ts
│   │   └── actions/                   # Server Actions
│   │       ├── posts.ts
│   │       ├── users.ts
│   │       └── auth.ts
│   ├── components/                    # Shared UI components
│   │   ├── ui/                        # Primitive components (Button, Input, etc.)
│   │   │   ├── button.tsx
│   │   │   ├── input.tsx
│   │   │   └── index.ts              # barrel export
│   │   ├── layout/                    # Layout components (Header, Footer, Sidebar)
│   │   │   ├── header.tsx
│   │   │   ├── footer.tsx
│   │   │   └── sidebar.tsx
│   │   └── features/                  # Feature-specific components
│   │       ├── posts/
│   │       │   ├── post-card.tsx
│   │       │   ├── post-form.tsx
│   │       │   └── post-list.tsx
│   │       └── users/
│   │           ├── user-avatar.tsx
│   │           └── user-menu.tsx
│   ├── lib/                           # Shared utilities (server + client safe)
│   │   ├── db/                        # Database client & schema
│   │   │   ├── index.ts              # Drizzle / Prisma client instantiation
│   │   │   └── schema.ts             # Database schema
│   │   ├── auth.ts                    # Auth configuration
│   │   ├── utils.ts                   # cn(), formatDate(), etc.
│   │   ├── validators.ts              # Zod schemas (shared server + client)
│   │   └── constants.ts
│   ├── hooks/                         # React hooks (client-only)
│   │   ├── use-auth.ts
│   │   ├── use-media-query.ts
│   │   └── use-debounce.ts
│   ├── providers/                     # Context providers
│   │   ├── providers.tsx             # Combined providers component
│   │   ├── theme-provider.tsx
│   │   └── auth-provider.tsx
│   ├── types/                         # TypeScript type definitions
│   │   └── index.ts
│   ├── public/                        # Static assets
│   │   ├── images/
│   │   └── fonts/
│   ├── e2e/                           # Playwright E2E tests
│   │   └── posts.spec.ts
│   ├── __tests__/                     # Unit / integration tests (vitest)
│   │   ├── components/
│   │   └── api/
│   ├── next.config.ts
│   ├── tailwind.config.ts
│   ├── components.json                # shadcn/ui config
│   ├── package.json
│   ├── tsconfig.json
│   ├── vitest.config.ts
│   └── playwright.config.ts
├── backend/                           # Go backend (see go-backend-reference.md)
├── packages/                          # Shared packages (if using Turborepo)
│   ├── shared-types/
│   │   ├── src/
│   │   │   └── index.ts
│   │   └── package.json
│   └── eslint-config/
│       └── package.json
├── package.json                       # Workspace root (Turborepo)
├── turbo.json
├── pnpm-workspace.yaml
└── Makefile
```

### DO NOT: Common structure mistakes

| ❌ Wrong | ✅ Right |
|----------|----------|
| `frontend/pages/` directory | `frontend/app/` — App Router uses `app/` directory |
| `frontend/src/app/` without `src/` configured | Either use `src/app/` (with proper config) or `app/` directly |
| Mixing Page Router and App Router in same directory | Choose one; if migrating, use `app/` alongside `pages/` (Next.js handles this) |
| Server-only code in `components/` | Put server-only logic in `lib/` or inline in Server Components |
| `frontend/app/components/` | Components go in `frontend/components/` at root — not inside `app/` |
| Barrel exports from `"use client"` files | Only re-export client-safe code; a barrel file with `"use client"` contaminates all exports |

### DO: Barrel export with client/server awareness

```tsx
// components/ui/index.ts
// ❌ DEEPSEEK HALLUCINATION — adding "use client" to a barrel file
// "use client";  <-- This forces ALL imports from this barrel to be client-side!

export { Button } from "./button";      // Client Component
export { Input } from "./input";        // Client Component
export { Card } from "./card";          // Could be either
```

**The fix:** Only add `"use client"` to individual component files that need it. Do NOT add it to barrel exports.

### DO: Server-only vs client-safe code separation

```tsx
// ✅ lib/db/index.ts — SERVER-ONLY
import "server-only"; // ✅ Fails at compile-time if imported from a Client Component

import { PrismaClient } from "@prisma/client";

const globalForPrisma = globalThis as unknown as { prisma: PrismaClient };

export const db = globalForPrisma.prisma ?? new PrismaClient();

if (process.env.NODE_ENV !== "production") {
  globalForPrisma.prisma = db;
}
```

```tsx
// ✅ lib/utils.ts — CLIENT-SAFE
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
```

---

## 9. Common AI Hallucinations

This section catalogs specific Next.js APIs that DeepSeek invents, gets wrong, or mixes with Pages Router patterns.

### Hallucination: Non-existent imports

```tsx
// ❌ HALLUCINATION — these DO NOT EXIST
import { useRouter } from "next/router";          // ❌ Pages Router import — does not work in App Router
import { withRouter } from "next/router";         // ❌ Pages Router HOC — no App Router equivalent
import { NextPage, GetServerSideProps } from "next"; // ❌ Pages Router types
import { getServerSession } from "next-auth";     // ❌ v4 API; v5 uses `auth()` from `@/lib/auth`
import { ImageResponse } from "next/server";      // ❌ Actually from `next/og` not `next/server`
import { cookies, headers } from "next/headers";  // ✅ This IS correct, but DeepSeek often forgets to AWAIT them in Next.js 15
import { notFound, redirect } from "next/navigation"; // ✅ Correct for App Router
```

### Hallucination: Pages Router patterns in App Router

```tsx
// ❌ DEEPSEEK HALLUCINATION — Pages Router patterns in App Router
export default function Page({ posts }: { posts: Post[] }) { // ❌ No getServerSideProps
  return <ul>{posts.map(p => <li key={p.id}>{p.title}</li>)}</ul>;
}

export async function getServerSideProps() { // ❌ Pages Router — doesn't work in App Router
  const posts = await db.post.findMany();
  return { props: { posts } };
}
```

**The fix in App Router:**

```tsx
// ✅ App Router equivalent
export default async function Page() {
  const posts = await db.post.findMany();
  return <ul>{posts.map((p) => <li key={p.id}>{p.title}</li>)}</ul>;
}
```

### Hallucination: `ImageResponse` from wrong module

```tsx
// ❌ DEEPSEEK HALLUCINATION
import { ImageResponse } from "next/server"; // ❌ Wrong module

// ✅ Correct
import { ImageResponse } from "next/og";
```

### Hallucination: `useRouter` from wrong package

```tsx
// ❌ DEEPSEEK HALLUCINATION — still uses Pages Router import
import { useRouter } from "next/router";        // ❌ Pages Router

// ✅ Correct for App Router
import { useRouter } from "next/navigation";     // ✅ App Router

// Different APIs!
// App Router: router.push("/foo"), router.refresh(), router.back(), router.prefetch("/foo")
// Pages Router: router.push("/foo"), router.replace("/foo"), router.query.id
// App Router has NO `router.query` — use useSearchParams() instead
```

### Hallucination: `router.query` in App Router

```tsx
// ❌ DEEPSEEK HALLUCINATION
"use client";

import { useRouter } from "next/navigation";

export function Component() {
  const router = useRouter();
  const { id } = router.query; // ❌ router.query does NOT exist in App Router
}

// ✅ Use useSearchParams or useParams
"use client";

import { useSearchParams, useParams } from "next/navigation";

export function Component() {
  const searchParams = useSearchParams();
  const params = useParams();

  const id = params.id as string;           // dynamic route param
  const q = searchParams.get("q");          // ?q=search
}
```

### Hallucination: "use server" on components

```tsx
// ❌ DEEPSEEK HALLUCINATION — "use server" on a component file
"use server";

export default function MyComponent() {  // ❌ "use server" is for Server Actions, not components
  return <div>Hello</div>;
}

// ✅ "use server" is for marking exports as Server Actions (in a separate file)
"use server";

import { revalidatePath } from "next/cache";

export async function deleteUser(id: string) {
  await db.user.delete({ where: { id } });
  revalidatePath("/users");
}
```

### Hallucination: Using `async` in Client Components

```tsx
// ❌ DEEPSEEK HALLUCINATION
"use client";

export default async function ClientPage() {  // ❌ Client Components can't be async
  const data = await fetch("/api/data");
  return <div>{JSON.stringify(data)}</div>;
}
```

### Hallucination: `Link` with wrong props

```tsx
// ✅ Correct
import Link from "next/link";

<Link href="/posts/123">Post 123</Link>
<Link href={{ pathname: "/posts/[id]", query: { id: "123" } }}> {/* ❌ `query` doesn't exist */}
<Link href="/posts/123" prefetch={false}>...</Link>      {/* ✅ */}
<Link href="/posts/123" scroll={false}>...</Link>         {/* ✅ */}
<Link href="/posts/123" replace>...</Link>                {/* ✅ */}
```

### Hallucination: Middleware edge-runtime APIs

```tsx
// ❌ DEEPSEEK HALLUCINATION — using Node APIs in middleware
import fs from "fs";                         // ❌ Middleware runs on Edge — no fs
import { PrismaClient } from "@prisma/client"; // ❌ Prisma doesn't run on Edge by default
import { createConnection } from "mysql2";    // ❌ No raw DB connections in middleware

// ✅ Middleware is limited to:
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
```

### Hallucination: `next/headers` not awaited in Next.js 15

```tsx
// ❌ DEEPSEEK HALLUCINATION — headers() is synchronous (Next.js 13/14 pattern)
import { headers } from "next/headers";

export default async function Page() {
  const heads = headers();              // ❌ Next.js 15: Must await
  const host = heads.get("host");       // ❌ TypeScript error
}

// ✅ Next.js 15 — cookies() and headers() MUST be awaited
import { headers, cookies } from "next/headers";

export default async function Page() {
  const headersList = await headers();  // ✅
  const cookieStore = await cookies();  // ✅
  const host = headersList.get("host");
  const token = cookieStore.get("token")?.value;
}
```

### Hallucination: Non-existent Next.js 15 config options

```tsx
// ❌ HALLUCINATION — these config options don't exist or were removed
// next.config.ts
const nextConfig = {
  experimental: {
    appDir: true,            // ❌ Removed — App Router is default in 13.4+
    serverComponents: true,  // ❌ Never existed
    runtime: "edge",         // ❌ Use `export const runtime = "edge"` per-page
  },
  fetchCache: "default-cache", // ❌ Removed in Next.js 15
};
```

### Hallucination: `revalidate` in page-level config

```tsx
// ❌ DEEPSEEK HALLUCINATION
export const revalidate = 60;    // ✅ This IS valid (ISR-like)
export const dynamic = "static"; // ✅ This IS valid (but rarely needed in 15)
export const runtime = "edge";   // ✅ This IS valid

// ❌ BUT these are commonly hallucinated:
export const fetchCache = "force-cache"; // ❌ Removed in Next.js 15
export const dynamicParams = false;       // ✅ This IS valid (return 404 for unknown dynamic params)
```

---

## 10. Middleware

### DO: Middleware file conventions

```tsx
// middleware.ts — at the PROJECT ROOT (same level as app/), or inside src/ if using src/
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Protected routes
  if (pathname.startsWith("/dashboard")) {
    const token = request.cookies.get("session-token")?.value;

    if (!token) {
      const loginUrl = new URL("/login", request.url);
      loginUrl.searchParams.set("callbackUrl", pathname);
      return NextResponse.redirect(loginUrl);
    }
  }

  return NextResponse.next();
}

// ✅ Matcher config — middleware only runs on matching paths
export const config = {
  matcher: [
    "/dashboard/:path*",  // match /dashboard and all sub-paths
    "/api/:path*",         // match all API routes
    "/((?!_next|api/auth|favicon.ico).*)", // everything except static files and auth
  ],
};
```

### DO: Middleware patterns

```tsx
// middleware.ts
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
  const response = NextResponse.next();

  // Pattern 1: Set headers
  response.headers.set("x-pathname", request.nextUrl.pathname);
  response.headers.set("x-request-id", crypto.randomUUID());

  // Pattern 2: Geo-based redirect
  const country = request.geo?.country ?? "US";
  if (country === "EU" && request.nextUrl.pathname === "/pricing") {
    return NextResponse.redirect(new URL("/pricing-eu", request.url));
  }

  // Pattern 3: Bot protection
  const ua = request.headers.get("user-agent") ?? "";
  if (ua.includes("BadBot")) {
    return new NextResponse("Access denied", { status: 403 });
  }

  // Pattern 4: IP-based rate limiting (simple)
  const ip = request.ip ?? "unknown";
  // In production, use Upstash or similar for distributed rate limiting

  // Pattern 5: A/B testing via cookie
  const variant = request.cookies.get("ab-test")?.value;
  if (!variant && request.nextUrl.pathname === "/") {
    const newVariant = Math.random() > 0.5 ? "a" : "b";
    response.cookies.set("ab-test", newVariant, { maxAge: 60 * 60 * 24 * 30 });
  }

  return response;
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
```

### DO NOT: Edge runtime limitations in middleware

```tsx
// ❌ DEEPSEEK HALLUCINATION — these WON'T work in middleware
import fs from "node:fs";                        // ❌ No filesystem
import path from "node:path";                    // ❌ Limited Node APIs
import { db } from "@/lib/db";                   // ❌ Connect to DB from edge? Only with edge-compatible clients
import { Redis } from "@upstash/redis";          // ✅ This works (Upstash has edge SDK)
import { PrismaClient } from "@prisma/client";    // ❌ Prisma needs Node.js runtime (use Prisma Accelerate for edge)
import jwt from "jsonwebtoken";                  // ❌ jsonwebtoken uses Node crypto
import { SignJWT, jwtVerify } from "jose";       // ✅ jose works on Edge runtime
```

### DO: Auth middleware with `jose` (Edge-compatible JWT)

```tsx
// lib/auth-edge.ts — Edge-compatible auth utilities
import { jwtVerify } from "jose";

const JWT_SECRET = new TextEncoder().encode(process.env.JWT_SECRET!);

export async function verifyEdgeToken(token: string) {
  try {
    const { payload } = await jwtVerify(token, JWT_SECRET);
    return payload as { sub: string; email: string };
  } catch {
    return null;
  }
}
```

```tsx
// middleware.ts
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { verifyEdgeToken } from "@/lib/auth-edge";

export async function middleware(request: NextRequest) {
  const token = request.cookies.get("session-token")?.value;

  if (token) {
    const payload = await verifyEdgeToken(token);
    if (payload) {
      const response = NextResponse.next();
      response.headers.set("x-user-id", payload.sub);
      return response;
    }
  }

  // Redirect to login if on a protected route
  if (request.nextUrl.pathname.startsWith("/dashboard")) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  return NextResponse.next();
}
```

### DO NOT: DeepSeek hallucinates `request.page` or `request.query`

```tsx
// ❌ HALLUCINATION — Pages Router middleware API
export function middleware(request: NextRequest) {
  const { page } = request;      // ❌ request.page doesn't exist
  const query = request.query;   // ❌ request.query doesn't exist
}

// ✅ Use request.nextUrl
export function middleware(request: NextRequest) {
  const { pathname, searchParams } = request.nextUrl;
  const q = searchParams.get("q");
}
```

### DO NOT: Use `rewrite` as a replacement for `redirect` without understanding the difference

```tsx
// redirect — client URL changes, new request
return NextResponse.redirect(new URL("/login", request.url)); // 307 temporary
return NextResponse.redirect(new URL("/login", request.url), 301); // permanent

// rewrite — client URL unchanged, internal routing
return NextResponse.rewrite(new URL("/maintenance", request.url));
```

### DO: Chainable middleware pattern (if using multiple concerns)

```tsx
// middleware.ts
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { authMiddleware } from "@/middleware/auth";
import { loggingMiddleware } from "@/middleware/logging";
import { securityHeadersMiddleware } from "@/middleware/security-headers";

export function middleware(request: NextRequest) {
  // Chain middleware
  const authResult = authMiddleware(request);
  if (authResult instanceof NextResponse) return authResult;

  const loggingResult = loggingMiddleware(request);
  if (loggingResult instanceof NextResponse) return loggingResult;

  return securityHeadersMiddleware(request);
}
```

```tsx
// middleware/security-headers.ts
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function securityHeadersMiddleware(request: NextRequest) {
  const response = NextResponse.next();

  response.headers.set("X-Frame-Options", "DENY");
  response.headers.set("X-Content-Type-Options", "nosniff");
  response.headers.set("Referrer-Policy", "strict-origin-when-cross-origin");
  response.headers.set(
    "Permissions-Policy",
    "camera=(), microphone=(), geolocation=()"
  );

  // CSP (Content Security Policy)
  response.headers.set(
    "Content-Security-Policy",
    "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline';"
  );

  return response;
}
```

---

## Quick Reference: Final Checklist

Before deploying AI-generated Next.js code, verify:

- [ ] Every file with `useState`/`useEffect`/`onClick`/`useContext` has `"use client"` at the top
- [ ] No file has BOTH `"use client"` AND `async function Component()`
- [ ] `params` and `searchParams` are typed as `Promise<>` and `await`ed (Next.js 15)
- [ ] `headers()` and `cookies()` from `next/headers` are ALWAYS `await`ed (Next.js 15)
- [ ] Server Actions are in files with `"use server"` directive (NOT on components)
- [ ] `useRouter` imported from `next/navigation` (NOT `next/router`)
- [ ] No `getServerSideProps`/`getStaticProps`/`getStaticPaths` — those are Pages Router only
- [ ] No `fetch()` assumption of automatic caching — it's opt-in in Next.js 15
- [ ] `error.tsx` has `"use client"` directive
- [ ] Middleware uses Edge-compatible libraries (no `fs`, `prisma`, `jsonwebtoken`)
- [ ] Route handler params are awaited: `const { id } = await params`
- [ ] `notFound()` is called when a query returns null (not silently rendering)
- [ ] `generateMetadata` properly awaits params and searchParams
- [ ] No `"use client"` on barrel export files
- [ ] `import "server-only"` in server-only modules (db client, auth config)
- [ ] Zod validation on all user input (route handlers and server actions)
- [ ] `redirect()` calls are in try/catch if placed after other logic (they throw internally)

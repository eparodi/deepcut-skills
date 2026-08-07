# Expo (Managed Workflow) React Native Reference — Traps, Mistakes & Best Practices

> **Audience:** AI coding agents (DeepSeek) and developers writing production Expo React Native apps.
> **Versions:** Expo SDK 52+, Expo Router v4, React Native 0.76+, TypeScript 5.x, React Query v5, Zustand v5

---

## Table of Contents

1. [Managed Workflow Boundaries](#1-managed-workflow-boundaries)
2. [Monorepo Project Layout (mobile/)](#2-monorepo-project-layout)
3. [Expo Router (File-Based Routing)](#3-expo-router-file-based-routing)
4. [Navigation Patterns](#4-navigation-patterns)
5. [Data Fetching](#5-data-fetching)
6. [State Management](#6-state-management)
7. [Styling](#7-styling)
8. [Testing](#8-testing)
9. [Build & Deploy](#9-build--deploy)
10. [Common AI Hallucinations](#10-common-ai-hallucinations)
11. [Platform-Specific Code](#11-platform-specific-code)

---

## 1. Managed Workflow Boundaries

### Understanding the "Managed" vs "Bare" split

Expo's managed workflow runs your JS in a pre-built native shell. You do NOT have direct access to the `ios/` and `android/` directories. This is the single biggest source of AI hallucinations — DeepSeek frequently generates code that requires native modules not available in managed workflow.

### DO NOT: Things you CANNOT do in managed workflow (without `expo prebuild` / development build)

| ❌ Cannot Do in Pure Managed (Expo Go) | ✅ Alternative |
|----------------------------------------|---------------|
| Install arbitrary native modules (`react-native-*` packages NOT in Expo SDK) | Use an Expo SDK equivalent (`expo-camera`, `expo-notifications`, etc.) |
| Run background services (long-running tasks when app is killed) | `expo-task-manager` + `expo-background-fetch` (limited to periodic fetches, ~15 min interval minimum on iOS) |
| Bluetooth Classic (EEPROM/SPP) — only BLE is available | `expo-bluetooth` is BLE-only. For Classic, you MUST use a development build + `react-native-bluetooth-classic` |
| In-app purchases with native StoreKit/BillingClient code | `expo-in-app-purchases` (deprecated) or `react-native-iap` in a development build via EAS |
| Custom native UI components (Mapbox with custom overlays, custom video player, custom camera UI) | Expo's built-in components (`expo-camera`, `expo-av`, `expo-map`) cover most cases. For custom: development build |
| iOS App Clips, Widgets, Watch apps, iMessage extensions | Requires writing native code in Xcode — must use `expo prebuild` and work in bare workflow |
| Firebase Cloud Messaging (FCM) with custom native handling | Use `expo-notifications` which wraps FCM/APNs automatically |
| Custom keyboard extensions | Must write native Swift/Kotlin — bare workflow only |
| Accessing `react-native link` or CocoaPods directly | Use Expo plugins to configure native projects via `app.json`/`app.config.js` |

### DO: Everything available IN managed workflow

```tsx
// ✅ These expo-* packages work WITHOUT prebuild in Expo Go:
"expo-camera"           // Camera access + barcode scanning
"expo-notifications"    // Local + push notifications
"expo-location"         // Foreground + background location
"expo-secure-store"     // Keychain/Keystore storage for tokens
"expo-haptics"          // Haptic feedback
"expo-av"               // Audio/video playback
"expo-file-system"      // Read/write files in app sandbox
"expo-image-picker"     // Access photo library / camera roll
"expo-sharing"          // Share files to other apps
"expo-splash-screen"    // Control native splash screen
"expo-status-bar"       // Control the status bar
"expo-screen-orientation" // Lock/unlock screen orientation
"expo-linking"          // Deep linking + URL handling
"expo-web-browser"      // Open in-app browser (SFSafariViewController / Chrome Custom Tabs)
"expo-updates"          // Over-the-air (OTA) updates via EAS
"expo-constants"        // App version, build number, manifest, etc.
"expo-device"           // Device model, OS version, etc.
"expo-localization"     // Device locale, region, timezone
"expo-clipboard"        // Read/write clipboard
"expo-crypto"           // Cryptographic hashing (SHA, HMAC)
"expo-keep-awake"       // Prevent screen sleep
"expo-linear-gradient"  // Gradient views
"expo-blur"             // Blur effect views
```

### DO: EAS Build allows custom native code WITHOUT local eject

This is the most misunderstood concept. You can add ANY native module to your managed project, configure it with Expo plugins in `app.json`, and let EAS Build compile the native code in the cloud. You never run `expo eject` or touch `ios/`/`android/` locally.

```jsonc
// app.json — EAS Build with custom native modules (NO local prebuild needed)
{
  "expo": {
    "plugins": [
      // ✅ Expo plugins configure native code in the cloud during EAS Build
      [
        "expo-camera",
        { "cameraPermission": "Allow $(PRODUCT_NAME) to access your camera" }
      ],
      [
        "react-native-iap",  // ✅ third-party native module, configured via plugin
        { "receiptValidation": true }
      ]
    ]
  }
}
```

```bash
# ✅ Build with custom native code — NO local eject required
eas build --platform ios --profile production
eas build --platform android --profile production
```

### DO: Development builds (expo-dev-client) for local native module testing

When you need to test custom native modules locally (including your own native code), use a development build instead of Expo Go:

```bash
# 1. Install expo-dev-client
npx expo install expo-dev-client

# 2. Create a development build locally
npx expo prebuild        # generates ios/ and android/ directories locally
npx expo run:ios         # builds and runs on simulator
npx expo run:android     # builds and runs on emulator

# 3. Or use EAS Build for the dev client (no local prebuild)
eas build --profile development --platform ios

# 4. Start dev server — connects to development build, NOT Expo Go
npx expo start --dev-client
```

**Key distinction:**
- **Expo Go** = pre-built shell, only expo-* modules that were bundled into Go
- **Development build** = custom shell you build via `eas build --profile development`, includes your native modules
- **EAS Build production** = release build, includes your native modules, submits to stores

---

## 2. Monorepo Project Layout (mobile/)

### DO: Standard Expo monorepo layout

```
project-root/
├── mobile/                              # Expo managed React Native app
│   ├── app/                             # Expo Router v4 file-based routing (ALL screens here)
│   │   ├── _layout.tsx                  # Root layout (providers, splash screen, auth gate)
│   │   ├── (auth)/                       # Auth route group (no tab bar, no drawer)
│   │   │   ├── _layout.tsx              # Auth stack layout
│   │   │   ├── login.tsx
│   │   │   ├── register.tsx
│   │   │   └── forgot-password.tsx
│   │   ├── (tabs)/                       # Tab route group (main app with bottom tabs)
│   │   │   ├── _layout.tsx              # Tab navigator layout
│   │   │   ├── index.tsx                # Home / first tab
│   │   │   ├── search.tsx               # Search tab
│   │   │   ├── create.tsx               # Create tab (post/new item)
│   │   │   ├── notifications.tsx        # Notifications tab
│   │   │   └── profile.tsx              # Profile tab (could be a link to (profile) group)
│   │   ├── (profile)/                    # Profile route group (nested stack within)
│   │   │   ├── _layout.tsx              # Profile stack layout
│   │   │   ├── [userId].tsx             # Dynamic user profile
│   │   │   ├── settings.tsx
│   │   │   └── edit-profile.tsx
│   │   ├── post/
│   │   │   ├── [id].tsx                 # Post detail screen
│   │   │   └── comments.tsx             # Comments for a post (nested route)
│   │   ├── modal/                        # Modally presented screens
│   │   │   ├── create-post.tsx
│   │   │   └── image-picker.tsx
│   │   └── +not-found.tsx              # Custom 404 screen
│   ├── src/
│   │   ├── components/                   # Reusable UI components
│   │   │   ├── ui/                       # Primitive UI kit (Button, Input, Card, Avatar, etc.)
│   │   │   ├── PostCard.tsx
│   │   │   ├── UserAvatar.tsx
│   │   │   └── EmptyState.tsx
│   │   ├── hooks/                        # Custom hooks (NO UI — logic only)
│   │   │   ├── useAuth.ts
│   │   │   ├── usePosts.ts
│   │   │   ├── useDebounce.ts
│   │   │   └── useRefreshOnFocus.ts
│   │   ├── utils/                        # Pure utility functions
│   │   │   ├── api.ts                    # API client (axios/fetch wrapper)
│   │   │   ├── storage.ts                # expo-secure-store wrappers
│   │   │   ├── formatDate.ts
│   │   │   ├── cn.ts                     # className merge (for NativeWind)
│   │   │   └── constants.ts              # API URLs, feature flags, etc.
│   │   ├── stores/                       # Zustand stores
│   │   │   ├── authStore.ts
│   │   │   ├── themeStore.ts
│   │   │   └── appStore.ts
│   │   ├── providers/                    # React Context providers + wrapper components
│   │   │   ├── QueryProvider.tsx          # TanStack Query provider
│   │   │   ├── AuthProvider.tsx           # Auth context provider
│   │   │   └── ThemeProvider.tsx          # Theme context provider
│   │   ├── lib/                          # Library re-exports / configuration
│   │   │   ├── queryClient.ts           # TanStack Query client config
│   │   │   └── axios.ts                 # Axios instance with interceptors
│   │   └── types/                        # Shared TypeScript types
│   │       ├── user.ts
│   │       ├── post.ts
│   │       └── navigation.ts            # Expo Router typed routes (autogenerated)
│   ├── assets/                           # Static assets (images, fonts, animations)
│   │   ├── images/
│   │   ├── fonts/
│   │   └── animations/                  # Lottie JSON files
│   ├── app.json                          # Expo configuration (app name, permissions, plugins)
│   ├── app.config.ts                     # Dynamic config (env vars, versioning) — imports from app.json
│   ├── eas.json                          # EAS Build profiles, submit profiles
│   ├── tsconfig.json
│   ├── babel.config.js                   # Babel config (NativeWind, reanimated plugin, etc.)
│   ├── metro.config.js                   # Metro bundler config (SVG transformer, etc.)
│   ├── tailwind.config.js               # NativeWind Tailwind config (if using)
│   ├── nativewind-env.d.ts              # NativeWind type declarations (if using)
│   ├── jest.config.ts                    # Jest configuration
│   ├── jest.setup.ts                     # Jest setup (MSW, mocks)
│   ├── package.json
│   └── .env                              # Environment variables (DO NOT commit secrets)
├── backend/                              # Go backend (separate module)
├── web/                                  # Next.js web app (separate module)
└── package.json                          # Root workspace config (if monorepo tooling)
```

### DO NOT: Common layout mistakes

| ❌ Wrong | ✅ Right |
|----------|----------|
| `mobile/screens/` directory | `mobile/app/` — Expo Router uses file-based routing; `screens/` is bare React Navigation |
| `navigation/` folder with `createStackNavigator()` | `app/_layout.tsx` with `<Stack>` — Expo Router wraps React Navigation automatically |
| Mixing `app/` files with manual `NavigationContainer` | NEVER manually create a `NavigationContainer` when using Expo Router |
| `src/` at the root without `app/` | `app/` is MANDATORY for Expo Router — it's the file-system router root |
| `components/UserAvatar.tsx` at root | `src/components/UserAvatar.tsx` — keep `app/` for routes only |
| `app.json` without `expo` key wrapper | All Expo config MUST be nested under `"expo": {}` in `app.json` |

### DO: Root `_layout.tsx` skeleton (the entry point for everything)

```tsx
// app/_layout.tsx
import "react-native-reanimated"; // MUST be first import for reanimated
import { useEffect } from "react";
import { Stack } from "expo-router";
import { StatusBar } from "expo-status-bar";
import * as SplashScreen from "expo-splash-screen";
import { GestureHandlerRootView } from "react-native-gesture-handler";
import { useFonts } from "expo-font";
import { QueryProvider } from "@/src/providers/QueryProvider";
import { AuthProvider } from "@/src/providers/AuthProvider";
import { ThemeProvider } from "@/src/providers/ThemeProvider";

// Prevent splash screen from auto-hiding
SplashScreen.preventAutoHideAsync();

export default function RootLayout() {
  const [loaded, error] = useFonts({
    "Inter-Regular": require("../assets/fonts/Inter-Regular.ttf"),
    "Inter-Bold": require("../assets/fonts/Inter-Bold.ttf"),
  });

  useEffect(() => {
    if (loaded || error) {
      // Hide splash screen once fonts are loaded
      SplashScreen.hideAsync();
    }
  }, [loaded, error]);

  if (!loaded && !error) {
    return null; // Keep splash screen visible while loading
  }

  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
      <QueryProvider>
        <ThemeProvider>
          <AuthProvider>
            <StatusBar style="auto" />
            {/* Root stack — conditionally renders auth vs app groups */}
            <Stack screenOptions={{ headerShown: false }}>
              <Stack.Screen name="(auth)" options={{ headerShown: false }} />
              <Stack.Screen name="(tabs)" options={{ headerShown: false }} />
              <Stack.Screen
                name="modal/create-post"
                options={{ presentation: "modal", headerShown: true, title: "New Post" }}
              />
            </Stack>
          </AuthProvider>
        </ThemeProvider>
      </QueryProvider>
    </GestureHandlerRootView>
  );
}
```

### DO: `app.json` configuration

```jsonc
// app.json
{
  "expo": {
    "name": "MyApp",
    "slug": "my-app",
    "version": "1.0.0",
    "orientation": "portrait",
    "icon": "./assets/images/icon.png",
    "scheme": "myapp",                     // deep linking scheme: myapp://
    "userInterfaceStyle": "automatic",     // light/dark based on device setting
    "newArchEnabled": true,                // Enable React Native New Architecture
    "splash": {
      "image": "./assets/images/splash.png",
      "resizeMode": "contain",
      "backgroundColor": "#ffffff"
    },
    "ios": {
      "supportsTablet": true,
      "bundleIdentifier": "com.mycompany.myapp",
      "infoPlist": {
        "NSCameraUsageDescription": "We need camera access to take photos",
        "NSPhotoLibraryUsageDescription": "We need photo library access to select images"
      }
    },
    "android": {
      "adaptiveIcon": {
        "foregroundImage": "./assets/images/adaptive-icon.png",
        "backgroundColor": "#ffffff"
      },
      "package": "com.mycompany.myapp"
    },
    "plugins": [
      "expo-router",
      "expo-secure-store",
      [
        "expo-camera",
        {
          "cameraPermission": "Allow $(PRODUCT_NAME) to access your camera"
        }
      ],
      [
        "expo-notifications",
        {
          "icon": "./assets/images/notification-icon.png",
          "color": "#ffffff"
        }
      ]
    ],
    "experiments": {
      "typedRoutes": true                   // Expo Router v4 typed routes
    }
  }
}
```

### DO: Dynamic `app.config.ts` (for environment variables)

```ts
// app.config.ts
import { ExpoConfig, ConfigContext } from "@expo/config";

// NEVER hardcode secrets in app.json — use .env or EAS secrets
const IS_DEV = process.env.APP_VARIANT === "development";
const IS_PREVIEW = process.env.APP_VARIANT === "preview";

export default ({ config }: ConfigContext): ExpoConfig => ({
  ...config,
  name: IS_DEV ? "MyApp (Dev)" : IS_PREVIEW ? "MyApp (Preview)" : "MyApp",
  slug: "my-app",
  // Use .env values set via EAS secrets or local .env
  extra: {
    apiUrl: process.env.EXPO_PUBLIC_API_URL || "https://api.myapp.com",
    sentryDsn: process.env.EXPO_PUBLIC_SENTRY_DSN,
    eas: {
      projectId: "your-eas-project-id-here",
    },
  },
  updates: {
    url: "https://u.expo.dev/your-eas-project-id-here",
  },
  runtimeVersion: {
    policy: "appVersion", // or "sdkVersion" or a custom string
  },
});
```

### DO NOT: Common `app.json` / config mistakes

| ❌ Wrong | ✅ Right |
|----------|----------|
| `"orientation": "landscape"` on a portrait-first app | `"orientation": "portrait"` — this locks the default; users can still rotate |
| Forgetting `"scheme"` for deep linking | Always set `"scheme": "myapp"` for `myapp://` deep links |
| `"newArchEnabled": false` or omitted | Set to `true` — New Architecture is stable in SDK 52+ |
| Hardcoding API URLs in `app.json` | Use `app.config.ts` with `EXPO_PUBLIC_*` env vars or EAS secrets |
| Missing `expo-router` in `plugins` array | `"expo-router"` MUST be in plugins for expo-router to work |
| Forgetting `"typedRoutes": true` | Expo Router v4+ supports typed routes — enable it |

---

## 3. Expo Router (File-Based Routing)

Expo Router v4 is built on React Navigation but replaces manual navigator construction with file-system conventions. DeepSeek frequently generates bare React Navigation code or mixes the two — a critical mistake.

### The core rules

1. **Every file in `app/` is a route.** `app/hello.tsx` → `/hello`
2. **`_layout.tsx` wraps child routes** with a navigator (Stack, Tabs, Drawer)
3. **`index.tsx`** is the default/root route for a directory
4. **`[param].tsx`** defines a dynamic route segment
5. **`(group)` parentheses** define route groups (affect layout, NOT URL)
6. **Directories create nested routes.** `app/post/[id].tsx` → `/post/123`

### DO: Root layout with Stack navigator

```tsx
// app/_layout.tsx — Root layout wraps everything
import { Stack } from "expo-router";

export default function RootLayout() {
  return (
    <Stack screenOptions={{ headerShown: false }}>
      {/* Screens defined as files in app/ are automatically available */}
    </Stack>
  );
}
```

### DO: Tab navigator with `(tabs)/_layout.tsx`

```tsx
// app/(tabs)/_layout.tsx
import { Tabs } from "expo-router";
import { Ionicons } from "@expo/vector-icons";

export default function TabLayout() {
  return (
    <Tabs
      screenOptions={{
        headerShown: false,
        tabBarActiveTintColor: "#007AFF",
        tabBarInactiveTintColor: "#8E8E93",
        tabBarStyle: {
          backgroundColor: "#FFFFFF",
          borderTopColor: "#E5E5EA",
        },
      }}
    >
      <Tabs.Screen
        name="index"
        options={{
          title: "Home",
          tabBarIcon: ({ color, size }) => (
            <Ionicons name="home" size={size} color={color} />
          ),
        }}
      />
      <Tabs.Screen
        name="search"
        options={{
          title: "Search",
          tabBarIcon: ({ color, size }) => (
            <Ionicons name="search" size={size} color={color} />
          ),
        }}
      />
      <Tabs.Screen
        name="profile"
        options={{
          title: "Profile",
          tabBarIcon: ({ color, size }) => (
            <Ionicons name="person" size={size} color={color} />
          ),
        }}
      />
    </Tabs>
  );
}
```

### DO: Dynamic routes `[id].tsx` with `useLocalSearchParams`

```tsx
// app/post/[id].tsx
import { View, Text } from "react-native";
import { useLocalSearchParams, Stack } from "expo-router";

export default function PostScreen() {
  // ✅ useLocalSearchParams for dynamic route params
  const { id } = useLocalSearchParams<{ id: string }>();

  return (
    <View style={{ flex: 1, padding: 16 }}>
      <Stack.Screen options={{ title: `Post #${id}` }} />
      <Text>Post ID: {id}</Text>
    </View>
  );
}
```

### DO: `Link` component for navigation

```tsx
import { View } from "react-native";
import { Link } from "expo-router";

export default function HomeScreen() {
  return (
    <View>
      {/* ✅ Simple string path */}
      <Link href="/post/123">Post 123</Link>

      {/* ✅ Object with params — TYPED with typedRoutes: true */}
      <Link href={{ pathname: "/post/[id]", params: { id: "456" } }}>
        Post 456
      </Link>

      {/* ✅ Link can be styled */}
      <Link href="/profile/settings" style={{ color: "blue", fontWeight: "bold" }}>
        Settings
      </Link>
    </View>
  );
}
```

### DO: `useRouter` for imperative navigation

```tsx
// src/components/PostCard.tsx
import { Pressable, Text } from "react-native";
import { useRouter } from "expo-router";

export function PostCard({ postId }: { postId: string }) {
  const router = useRouter();

  return (
    <Pressable
      onPress={() => {
        // ✅ Imperative navigation
        router.push(`/post/${postId}`);
        // router.replace(`/post/${postId}`);  // replace current screen
        // router.back();                       // go back
        // router.dismiss();                    // dismiss modal
        // router.dismissAll();                 // dismiss all modals
      }}
    >
      <Text>Go to post {postId}</Text>
    </Pressable>
  );
}
```

### DO: Route groups `(group)` for layout grouping (does NOT affect URL)

```
app/
├── (auth)/
│   ├── _layout.tsx     # Stack navigator for auth — NO tabs visible
│   ├── login.tsx        # URL: /login  (NOT /(auth)/login)
│   └── register.tsx     # URL: /register
├── (tabs)/
│   ├── _layout.tsx     # Tab navigator
│   ├── index.tsx        # URL: /  (home tab)
│   └── profile.tsx      # URL: /profile
└── _layout.tsx    # Root: conditionally shows auth OR tabs
```

### DO: Auth gate pattern in root `_layout.tsx`

```tsx
// app/_layout.tsx
import { Stack, useSegments, useRouter } from "expo-router";
import { useAuth } from "@/src/hooks/useAuth";
import { useEffect } from "react";

function AuthGate({ children }: { children: React.ReactNode }) {
  const { session, isLoading } = useAuth();
  const segments = useSegments();
  const router = useRouter();

  useEffect(() => {
    if (isLoading) return;

    const inAuthGroup = segments[0] === "(auth)";

    if (!session && !inAuthGroup) {
      // Redirect to login if not authenticated
      router.replace("/(auth)/login");
    } else if (session && inAuthGroup) {
      // Redirect to home if already authenticated
      router.replace("/(tabs)");
    }
  }, [session, isLoading, segments]);

  return <>{children}</>;
}

export default function RootLayout() {
  return (
    <AuthGate>
      <Stack screenOptions={{ headerShown: false }} />
    </AuthGate>
  );
}
```

### DO NOT: Common Expo Router mistakes

| ❌ Wrong | ✅ Right |
|----------|----------|
| `import { useRouter } from "next/navigation"` | `import { useRouter } from "expo-router"` — Expo Router NOT Next.js |
| `import { Link } from "next/link"` | `import { Link } from "expo-router"` |
| Creating `NavigationContainer` manually | Expo Router wraps it automatically in `_layout.tsx` — NEVER manually create one |
| `router.push("/home")` (Next.js style) | `router.push("/(tabs)")` or `router.push("/home")` (both work, but use typed routes) |
| `navigation.navigate("Post", { id: "123" })` | `router.push({ pathname: "/post/[id]", params: { id: "123" } })` — Expo Router uses file paths, not screen names |
| `useNavigation()` or `useRoute()` from `@react-navigation/native` | Use Expo Router's `useRouter()`, `useLocalSearchParams()`, `useSegments()` instead |
| `<Stack.Navigator>` (React Navigation bare API) | `<Stack>` — Expo Router's simplified component (no `.Navigator` suffix) |
| `(group)/_layout.tsx` with `Stack` AND `Tabs` nested | One navigator per `_layout.tsx`. Nest: stack → tabs → stack for complex layouts |
| Forgetting `import "react-native-reanimated"` at the top of `_layout.tsx` | MUST be the first import if using reanimated (drawer, shared transitions) |

### DO: Modal presentation

```tsx
// app/_layout.tsx
import { Stack } from "expo-router";

export default function RootLayout() {
  return (
    <Stack>
      <Stack.Screen name="(tabs)" options={{ headerShown: false }} />
      {/* ✅ Modal screens — group them under a modal/ directory */}
      <Stack.Screen
        name="modal/create-post"
        options={{
          presentation: "modal",
          headerShown: true,
          title: "New Post",
        }}
      />
      <Stack.Screen
        name="modal/image-picker"
        options={{
          presentation: "fullScreenModal", // iOS full-screen modal
        }}
      />
    </Stack>
  );
}
```

### DO NOT: Incorrect segment/group syntax

```tsx
// ❌ WRONG — DeepSeek frequently invents these:
// [userId]     → missing .tsx extension — Expo Router needs the file extension in the filename
// (tabs)       → missing /_layout.tsx inside — groups need a layout to work
// _layout.tsx  → NOT inside a directory — wraps nothing

// ✅ CORRECT directory structure:
// app/
//   (tabs)/
//     _layout.tsx    ← Tab navigator wraps index.tsx and profile.tsx
//     index.tsx      ← / (home)
//     profile.tsx    ← /profile
//   post/
//     _layout.tsx    ← Optional: stack navigator for post/*
//     [id].tsx       ← /post/123
//     comments.tsx   ← /post/comments  (if no [id] wrapping needed)
```

---

## 4. Navigation Patterns

### DO: Stack vs Tabs vs Drawer — when to use each

| Navigator | Use Case | Expo Router Syntax |
|-----------|----------|-------------------|
| **Stack** | Push/pop screens — settings, detail views, edit forms. Default for most flows. | `<Stack>` in `_layout.tsx` |
| **Tabs** | Primary app sections always visible — home, search, profile. Bottom tabs. | `<Tabs>` in `(tabs)/_layout.tsx` |
| **Drawer** | Side menu navigation — rarely used on mobile. Consider tabs instead. | `<Drawer>` in `_layout.tsx` (requires `react-native-gesture-handler`) |

### DO: Nested navigators — Stack wrapping Tabs wrapping Stack

```tsx
// app/_layout.tsx — Root: Stack
import { Stack } from "expo-router";

export default function RootLayout() {
  return (
    <Stack screenOptions={{ headerShown: false }}>
      <Stack.Screen name="(auth)" />
      <Stack.Screen name="(tabs)" />
      <Stack.Screen name="post/[id]" />    {/* pushed ON TOP of tabs */}
      <Stack.Screen name="modal/create-post" options={{ presentation: "modal" }} />
    </Stack>
  );
}
```

```tsx
// app/(tabs)/_layout.tsx — Tabs: Bottom tab navigator
import { Tabs } from "expo-router";

export default function TabLayout() {
  return (
    <Tabs screenOptions={{ headerShown: false }}>
      <Tabs.Screen name="index" options={{ title: "Home" }} />
      <Tabs.Screen name="search" options={{ title: "Search" }} />
      <Tabs.Screen name="profile" options={{ title: "Profile" }} />
    </Tabs>
  );
}
```

```tsx
// app/(tabs)/profile/ — Profile can have its own nested stack
// Create: app/(tabs)/profile/_layout.tsx
import { Stack } from "expo-router";

export default function ProfileStackLayout() {
  return (
    <Stack>
      <Stack.Screen name="index" options={{ title: "Profile" }} />
      <Stack.Screen name="settings" options={{ title: "Settings" }} />
      <Stack.Screen name="edit-profile" options={{ title: "Edit Profile" }} />
    </Stack>
  );
}
// Files: app/(tabs)/profile/index.tsx  → Tab: Profile (index)
//        app/(tabs)/profile/settings.tsx → pushed on top of profile
```

### DO: Auth flow pattern (conditional rendering)

```tsx
// The pattern: _layout.tsx at root level checks auth state
// and conditionally shows (auth) group OR (tabs) group.

// app/_layout.tsx
import { Stack } from "expo-router";
import { useAuth } from "@/src/hooks/useAuth";
import { View, ActivityIndicator } from "react-native";

export default function RootLayout() {
  const { session, isLoading } = useAuth();

  // Show loading screen while checking auth
  if (isLoading) {
    return (
      <View style={{ flex: 1, justifyContent: "center", alignItems: "center" }}>
        <ActivityIndicator size="large" />
      </View>
    );
  }

  return (
    <Stack screenOptions={{ headerShown: false }}>
      {session ? (
        // ✅ Authenticated: show main app
        <>
          <Stack.Screen name="(tabs)" />
          <Stack.Screen name="post/[id]" />
          <Stack.Screen name="modal/create-post" options={{ presentation: "modal" }} />
        </>
      ) : (
        // ✅ Not authenticated: show auth screens only
        <Stack.Screen name="(auth)" />
      )}
    </Stack>
  );
}
```

### DO NOT: Auth mistakes

| ❌ Wrong | ✅ Right |
|----------|----------|
| Using navigation guard in every screen | Single auth gate in root `_layout.tsx` |
| `useEffect` redirect not wrapped with `isLoading` check | Always check loading state first — avoid flash-of-wrong-screen |
| Storing token in `AsyncStorage` | Use `expo-secure-store` for tokens (encrypted) |
| Exposing all screens when not authenticated | Conditionally render `<Stack.Screen>` based on auth state |

### DO: Deep linking setup

```jsonc
// app.json — the "scheme" defines your custom URL scheme
{
  "expo": {
    "scheme": "myapp"  // deep links: myapp://post/123
  }
}
```

```tsx
// app/_layout.tsx — No special setup needed! Expo Router handles deep linking
// automatically based on your file structure.
// myapp://post/123 → app/post/[id].tsx with { id: "123" }
// myapp://profile/settings → app/(tabs)/profile/settings.tsx

// For universal links (https://myapp.com/post/123):
// Set up an apple-app-site-association (iOS) and assetlinks.json (Android)
// in your web server. Then add associated domains in app.json.
```

```jsonc
// app.json — Associated domains for universal links
{
  "expo": {
    "ios": {
      "associatedDomains": ["applinks:myapp.com"]
    },
    "android": {
      "intentFilters": [
        {
          "action": "VIEW",
          "autoVerify": true,
          "data": [
            {
              "scheme": "https",
              "host": "myapp.com",
              "pathPrefix": "/"
            }
          ],
          "category": ["BROWSABLE", "DEFAULT"]
        }
      ]
    }
  }
}
```

---

## 5. Data Fetching

### DO: React Query (TanStack Query v5) setup

```tsx
// src/lib/queryClient.ts
import { QueryClient } from "@tanstack/react-query";
import { AppState, AppStateStatus } from "react-native";

// ✅ Global QueryClient config
function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        // ✅ Sensible defaults for mobile
        staleTime: 1000 * 60 * 2,        // 2 minutes before data is considered stale
        gcTime: 1000 * 60 * 10,           // 10 minutes garbage collection (formerly cacheTime)
        retry: 2,                          // Retry failed queries twice
        retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 10000),
        refetchOnWindowFocus: false,       // NOT on mobile — use AppState instead
        refetchOnReconnect: true,
      },
      mutations: {
        retry: 0,                          // Don't retry mutations by default
      },
    },
  });
}

export const queryClient = createQueryClient();
```

```tsx
// src/providers/QueryProvider.tsx
import { QueryClientProvider } from "@tanstack/react-query";
import { queryClient } from "@/src/lib/queryClient";
import { useOnlineManager } from "@/src/hooks/useOnlineManager";
import { useAppState } from "@/src/hooks/useAppState";
import type { ReactNode } from "react";

export function QueryProvider({ children }: { children: ReactNode }) {
  // ✅ Refetch on app foreground + network reconnect
  useOnlineManager();
  useAppState();

  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
}
```

```tsx
// src/hooks/useOnlineManager.ts
import { useEffect } from "react";
import { onlineManager } from "@tanstack/react-query";
import NetInfo from "@react-native-community/netinfo";

export function useOnlineManager() {
  useEffect(() => {
    const unsubscribe = NetInfo.addEventListener((state) => {
      onlineManager.setOnline(
        state.isConnected != null && state.isConnected && Boolean(state.isInternetReachable)
      );
    });

    return () => unsubscribe();
  }, []);
}
```

```tsx
// src/hooks/useAppState.ts
import { useEffect, useRef } from "react";
import { AppState, AppStateStatus } from "react-native";
import { focusManager } from "@tanstack/react-query";

export function useAppState() {
  const appState = useRef(AppState.currentState);

  useEffect(() => {
    const subscription = AppState.addEventListener("change", (nextAppState) => {
      // ✅ Only refetch when coming from background to active
      if (appState.current.match(/inactive|background/) && nextAppState === "active") {
        focusManager.setFocused(true);
      } else {
        focusManager.setFocused(false);
      }
      appState.current = nextAppState;
    });

    return () => subscription.remove();
  }, []);
}
```

### DO: API client with axios + interceptors

```tsx
// src/lib/axios.ts
import axios, { AxiosError, InternalAxiosRequestConfig } from "axios";
import Constants from "expo-constants";
import { getAccessToken, getRefreshToken, setTokens, clearTokens } from "@/src/utils/storage";

const API_URL = Constants.expoConfig?.extra?.apiUrl || "https://api.myapp.com";

export const apiClient = axios.create({
  baseURL: API_URL,
  timeout: 15000,
  headers: { "Content-Type": "application/json" },
});

// ✅ Request interceptor — attach access token
apiClient.interceptors.request.use(
  async (config: InternalAxiosRequestConfig) => {
    const token = await getAccessToken();
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// ✅ Response interceptor — handle 401, refresh token, retry
let isRefreshing = false;
let failedQueue: Array<{
  resolve: (token: string) => void;
  reject: (error: unknown) => void;
}> = [];

const processQueue = (error: unknown, token: string | null) => {
  failedQueue.forEach(({ resolve, reject }) => {
    if (error) {
      reject(error);
    } else {
      resolve(token!);
    }
  });
  failedQueue = [];
};

apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };
    if (!originalRequest) return Promise.reject(error);

    // ✅ Only attempt refresh on 401 and not already retried
    if (error.response?.status === 401 && !originalRequest._retry) {
      // If already refreshing, queue this request
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({
            resolve: (token: string) => {
              if (originalRequest.headers) {
                originalRequest.headers.Authorization = `Bearer ${token}`;
              }
              resolve(apiClient(originalRequest));
            },
            reject,
          });
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        const refreshToken = await getRefreshToken();
        if (!refreshToken) throw new Error("No refresh token");

        const { data } = await axios.post(`${API_URL}/auth/refresh`, {
          refreshToken,
        });

        const { accessToken, refreshToken: newRefreshToken } = data;
        await setTokens(accessToken, newRefreshToken);

        processQueue(null, accessToken);

        if (originalRequest.headers) {
          originalRequest.headers.Authorization = `Bearer ${accessToken}`;
        }
        return apiClient(originalRequest);
      } catch (refreshError) {
        processQueue(refreshError, null);
        await clearTokens();
        // ✅ Navigate to login — use a global event emitter or zustand store
        // Avoid importing router directly in axios config (circular dependency)
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    return Promise.reject(error);
  }
);
```

### DO: Token storage with `expo-secure-store`

```tsx
// src/utils/storage.ts
import * as SecureStore from "expo-secure-store";

const ACCESS_TOKEN_KEY = "access_token";
const REFRESH_TOKEN_KEY = "refresh_token";

// ✅ Use expo-secure-store for sensitive data (encrypted)
// NEVER use AsyncStorage for tokens — it's unencrypted plain text

export async function getAccessToken(): Promise<string | null> {
  try {
    return await SecureStore.getItemAsync(ACCESS_TOKEN_KEY);
  } catch {
    return null;
  }
}

export async function getRefreshToken(): Promise<string | null> {
  try {
    return await SecureStore.getItemAsync(REFRESH_TOKEN_KEY);
  } catch {
    return null;
  }
}

export async function setTokens(accessToken: string, refreshToken: string): Promise<void> {
  await SecureStore.setItemAsync(ACCESS_TOKEN_KEY, accessToken);
  await SecureStore.setItemAsync(REFRESH_TOKEN_KEY, refreshToken);
}

export async function clearTokens(): Promise<void> {
  await SecureStore.deleteItemAsync(ACCESS_TOKEN_KEY);
  await SecureStore.deleteItemAsync(REFRESH_TOKEN_KEY);
}
```

### DO: Custom hook pattern with React Query

```tsx
// src/hooks/usePosts.ts
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/src/lib/axios";
import type { Post } from "@/src/types/post";

// ✅ Query keys as constants (avoids typos, enables cache invalidation)
export const postKeys = {
  all: ["posts"] as const,
  lists: () => [...postKeys.all, "list"] as const,
  list: (filters: Record<string, unknown>) => [...postKeys.lists(), filters] as const,
  details: () => [...postKeys.all, "detail"] as const,
  detail: (id: string) => [...postKeys.details(), id] as const,
};

// ✅ Fetch posts list
export function usePosts() {
  return useQuery({
    queryKey: postKeys.lists(),
    queryFn: async () => {
      const { data } = await apiClient.get<Post[]>("/posts");
      return data;
    },
  });
}

// ✅ Fetch single post
export function usePost(id: string) {
  return useQuery({
    queryKey: postKeys.detail(id),
    queryFn: async () => {
      const { data } = await apiClient.get<Post>(`/posts/${id}`);
      return data;
    },
    enabled: !!id,  // ✅ Don't fetch if id is undefined/null
  });
}

// ✅ Create post mutation
export function useCreatePost() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (newPost: { title: string; content: string }) => {
      const { data } = await apiClient.post<Post>("/posts", newPost);
      return data;
    },
    onSuccess: () => {
      // ✅ Invalidate cache to refetch list
      queryClient.invalidateQueries({ queryKey: postKeys.lists() });
    },
  });
}
```

### DO NOT: Data fetching mistakes

| ❌ Wrong | ✅ Right |
|----------|----------|
| `useEffect` + `useState` for every API call | React Query — handles loading, error, caching, retry automatically |
| `fetch()` directly without error handling | Use the configured `apiClient` with interceptors |
| `AsyncStorage.setItem("token", token)` | `SecureStore.setItemAsync("token", token)` — encrypted storage |
| Storing refresh token in a Zustand store only | Persist tokens in SecureStore, Zustand can hold runtime session state |
| Not handling 401 globally (each screen checks auth) | Axios response interceptor handles refresh token flow globally |
| `cacheTime` (TanStack Query v4) | `gcTime` (TanStack Query v5) — renamed |
| Forgetting `enabled` on dependent queries | `enabled: !!someId` prevents fetching with undefined params |
| `refetchOnWindowFocus: true` on mobile | Does nothing on native — use AppState listener with `focusManager.setFocused()` |

### DO: Offline support with NetInfo

```tsx
// src/hooks/useNetworkStatus.ts
import { useEffect, useState } from "react";
import NetInfo, { NetInfoState } from "@react-native-community/netinfo";

export function useNetworkStatus() {
  const [isConnected, setIsConnected] = useState<boolean | null>(true);

  useEffect(() => {
    const unsubscribe = NetInfo.addEventListener((state: NetInfoState) => {
      setIsConnected(state.isConnected);
    });

    return () => unsubscribe();
  }, []);

  return { isConnected };
}
```

```tsx
// src/providers/QueryProvider.tsx — offline persistence setup
// For offline-first, use @tanstack/react-query-persist-client
// with an AsyncStorage adapter:
//
// import { persistQueryClient } from "@tanstack/react-query-persist-client";
// import { createAsyncStoragePersister } from "@tanstack/query-async-storage-persister";
//
// const asyncStoragePersister = createAsyncStoragePersister({
//   storage: AsyncStorage,
//   key: "REACT_QUERY_OFFLINE_CACHE",
// });
//
// persistQueryClient({
//   queryClient,
//   persister: asyncStoragePersister,
//   maxAge: 1000 * 60 * 60 * 24, // 24 hours
// });

// NOTE: @tanstack/react-query-persist-client and @tanstack/query-async-storage-persister
// are separate packages. Only use if you need offline persistence.
```

---

## 6. State Management

### Decision tree: What to use when

```
Is the state needed by many components across different screens?
├── YES → Is it server state (fetched from API)?
│   ├── YES → React Query (TanStack Query)   ← already handled in §5
│   └── NO  → Zustand                         ← global client state
└── NO  → React Context (or component state)   ← localized state
```

### DO: Zustand for global client state

```tsx
// src/stores/authStore.ts
import { create } from "zustand";

interface AuthState {
  session: { userId: string; email: string } | null;
  isLoading: boolean;
  setSession: (session: AuthState["session"]) => void;
  clearSession: () => void;
  setLoading: (loading: boolean) => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  session: null,
  isLoading: true, // Start loading — check stored token on mount
  setSession: (session) => set({ session, isLoading: false }),
  clearSession: () => set({ session: null, isLoading: false }),
  setLoading: (isLoading) => set({ isLoading }),
}));
```

```tsx
// src/stores/themeStore.ts
import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import AsyncStorage from "@react-native-async-storage/async-storage";

type Theme = "light" | "dark" | "system";

interface ThemeState {
  theme: Theme;
  setTheme: (theme: Theme) => void;
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set) => ({
      theme: "system",
      setTheme: (theme) => set({ theme }),
    }),
    {
      name: "theme-storage",
      storage: createJSONStorage(() => AsyncStorage), // ✅ OK for non-sensitive data like theme
    }
  )
);
```

### DO: React Context for providers and scoped state

```tsx
// src/providers/AuthProvider.tsx
import { createContext, useContext, useEffect, type ReactNode } from "react";
import { useAuthStore } from "@/src/stores/authStore";
import { getAccessToken } from "@/src/utils/storage";
import { apiClient } from "@/src/lib/axios";

interface AuthContextValue {
  session: { userId: string; email: string } | null;
  isLoading: boolean;
  signIn: (email: string, password: string) => Promise<void>;
  signOut: () => Promise<void>;
  signUp: (email: string, password: string, name: string) => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const { session, isLoading, setSession, clearSession, setLoading } = useAuthStore();

  // ✅ Check for existing token on app launch
  useEffect(() => {
    async function loadSession() {
      try {
        const token = await getAccessToken();
        if (token) {
          const { data } = await apiClient.get("/auth/me");
          setSession({ userId: data.id, email: data.email });
        } else {
          clearSession();
        }
      } catch {
        clearSession();
      }
    }
    loadSession();
  }, []);

  const signIn = async (email: string, password: string) => {
    setLoading(true);
    try {
      const { data } = await apiClient.post("/auth/login", { email, password });
      setSession({ userId: data.user.id, email: data.user.email });
    } finally {
      setLoading(false);
    }
  };

  const signOut = async () => {
    try {
      await apiClient.post("/auth/logout");
    } finally {
      clearSession();
    }
  };

  const signUp = async (email: string, password: string, name: string) => {
    setLoading(true);
    try {
      const { data } = await apiClient.post("/auth/register", { email, password, name });
      setSession({ userId: data.user.id, email: data.user.email });
    } finally {
      setLoading(false);
    }
  };

  return (
    <AuthContext.Provider value={{ session, isLoading, signIn, signOut, signUp }}>
      {children}
    </AuthContext.Provider>
  );
}

// ✅ Custom hook — throws if used outside provider
export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
```

### DO NOT: State management anti-patterns

| ❌ Wrong | ✅ Right |
|----------|----------|
| Redux for a new Expo app | Zustand — simpler, less boilerplate, TypeScript-friendly, smaller bundle |
| `useState` for API data | React Query — handles cache, loading/error states, deduplication |
| `useContext` for frequently-updating state | Zustand — Context re-renders ALL consumers on ANY change |
| Prop drilling beyond 2-3 levels | Zustand for shared global state; composition for local state |
| Storing auth tokens in Zustand | SecureStore for persisting tokens; Zustand only for runtime session |
| `persist` middleware on auth store | Sensitive data should NOT be persisted in AsyncStorage — use SecureStore |
| Global state for everything | Ask: "Does this need to survive screen transitions?" If no, keep it local |

---

## 7. Styling

### DO: `StyleSheet.create` — ALWAYS

```tsx
import { StyleSheet, View, Text } from "react-native";

export function PostCard({ title, body }: { title: string; body: string }) {
  return (
    <View style={styles.container}>
      <Text style={styles.title}>{title}</Text>
      <Text style={styles.body} numberOfLines={3}>{body}</Text>
    </View>
  );
}

// ✅ ALWAYS use StyleSheet.create — validates styles, sends them to native side
const styles = StyleSheet.create({
  container: {
    padding: 16,
    backgroundColor: "#FFFFFF",
    borderRadius: 12,
    marginHorizontal: 16,
    marginVertical: 8,
    // ✅ Shadow for iOS
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    // ✅ Shadow for Android (must use elevation)
    elevation: 3,
  },
  title: {
    fontSize: 18,
    fontWeight: "600",
    color: "#1A1A1A",
    marginBottom: 8,
  },
  body: {
    fontSize: 14,
    color: "#666666",
    lineHeight: 20,
  },
});
```

### DO NOT: Inline styles

```tsx
// ❌ WRONG — DeepSeek frequently generates this
export function BadComponent() {
  return (
    <View style={{ padding: 16, backgroundColor: "white" }}>
      <Text style={{ fontSize: 18, fontWeight: "bold" }}>Hello</Text>
    </View>
  );
}
// Problem: New style object EVERY render → RN sends new style over bridge every render.
// StyleSheet.create memoizes the style on the native side.

// ✅ RIGHT
export function GoodComponent() {
  return (
    <View style={styles.container}>
      <Text style={styles.title}>Hello</Text>
    </View>
  );
}
```

### DO: Theme system (light/dark)

```tsx
// src/theme/colors.ts
export const colors = {
  light: {
    background: "#FFFFFF",
    surface: "#F2F2F7",
    text: "#1A1A1A",
    textSecondary: "#8E8E93",
    primary: "#007AFF",
    border: "#E5E5EA",
    error: "#FF3B30",
    success: "#34C759",
  },
  dark: {
    background: "#000000",
    surface: "#1C1C1E",
    text: "#FFFFFF",
    textSecondary: "#8E8E93",
    primary: "#0A84FF",
    border: "#38383A",
    error: "#FF453A",
    success: "#30D158",
  },
} as const;

export type ColorScheme = keyof typeof colors;
```

```tsx
// src/hooks/useTheme.ts
import { useColorScheme } from "react-native";
import { useThemeStore } from "@/src/stores/themeStore";
import { colors, type ColorScheme } from "@/src/theme/colors";

export function useAppTheme() {
  const systemColorScheme = useColorScheme(); // "light" | "dark" | null
  const { theme } = useThemeStore();           // "light" | "dark" | "system"

  const resolvedScheme: ColorScheme =
    theme === "system"
      ? (systemColorScheme as ColorScheme) || "light"
      : theme;

  return {
    colors: colors[resolvedScheme],
    colorScheme: resolvedScheme,
    isDark: resolvedScheme === "dark",
  };
}
```

```tsx
// src/components/PostCard.tsx — Using the theme
import { StyleSheet, View, Text } from "react-native";
import { useAppTheme } from "@/src/hooks/useTheme";

export function PostCard({ title }: { title: string }) {
  const { colors } = useAppTheme();

  return (
    <View style={[styles.container, { backgroundColor: colors.surface }]}>
      <Text style={[styles.title, { color: colors.text }]}>{title}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { padding: 16, borderRadius: 12 },
  title: { fontSize: 18, fontWeight: "600" },
});
```

### DO: Platform-specific styles

```tsx
import { Platform, StyleSheet } from "react-native";

const styles = StyleSheet.create({
  container: {
    padding: 16,
    // ✅ Platform.select for platform-specific values
    ...Platform.select({
      ios: {
        shadowColor: "#000",
        shadowOffset: { width: 0, height: 2 },
        shadowOpacity: 0.1,
        shadowRadius: 4,
      },
      android: {
        elevation: 3,
      },
    }),
    // ✅ Platform.OS for conditional logic
    backgroundColor: Platform.OS === "ios" ? "#FFFFFF" : "#FAFAFA",
  },
  // ✅ Platform.select for fonts (San Francisco on iOS, Roboto on Android)
  header: {
    ...Platform.select({
      ios: { fontFamily: "System" },
      android: { fontFamily: "Roboto", fontWeight: "500" },
    }),
  },
});
```

### DO: NativeWind (Tailwind on React Native) — if desired

If you choose NativeWind v4+ for utility-first styling, here's the setup:

```bash
npx expo install nativewind tailwindcss react-native-reanimated
npx tailwindcss init
```

```tsx
// tailwind.config.js
/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./app/**/*.{js,jsx,ts,tsx}",
    "./src/**/*.{js,jsx,ts,tsx}",
  ],
  presets: [require("nativewind/preset")],
  theme: {
    extend: {
      colors: {
        primary: "#007AFF",
        surface: "#F2F2F7",
      },
    },
  },
  plugins: [],
};
```

```tsx
// babel.config.js
module.exports = function (api) {
  api.cache(true);
  return {
    presets: [
      ["babel-preset-expo", { jsxImportSource: "nativewind" }],
      "nativewind/babel",
    ],
  };
};
```

```tsx
// src/components/PostCard.nativewind.tsx — Using NativeWind
import { View, Text } from "react-native";

export function PostCard({ title }: { title: string }) {
  return (
    <View className="p-4 bg-white dark:bg-gray-900 rounded-xl shadow-sm mx-4 my-2">
      <Text className="text-lg font-semibold text-gray-900 dark:text-white">
        {title}
      </Text>
    </View>
  );
}
```

### DO NOT: Styling mistakes

| ❌ Wrong | ✅ Right |
|----------|----------|
| `style={{ padding: 16 }}` inline every render | `StyleSheet.create({ ... })` — always |
| `"100%"` as width string | `"100%"` works but `Dimensions.get("window").width` or `flex: 1` is better |
| `gap: 8` on old RN | `gap` requires RN 0.71+ (Expo SDK 48+). Use `margin` for older versions |
| `boxShadow: "0 2px 4px rgba(...)"` | This is CSS — doesn't exist in RN. Use `shadowColor`, `shadowOffset`, `shadowOpacity`, `shadowRadius` on iOS; `elevation` on Android |
| `overflow: "visible"` on Android with shadow | `overflow: "hidden"` clips shadows on both platforms |
| `fontWeight: "bold"` with custom font not loaded | Custom fonts need `fontWeight` to match the font file's weight |
| Not importing `StyleSheet` | `import { StyleSheet } from "react-native";` — DeepSeek often forgets this |

---

## 8. Testing

### DO: Jest configuration for Expo

```ts
// jest.config.ts
import type { Config } from "jest";

const config: Config = {
  preset: "jest-expo", // ✅ Expo's Jest preset — handles all transforms
  transformIgnorePatterns: [
    // ✅ Essential: transform ESM node_modules that aren't pre-compiled
    "node_modules/(?!(" +
      "expo-router|" +
      "expo-secure-store|" +
      "@react-native|" +
      "@react-navigation|" +
      "react-native|" +
      "react-native-.*|" +
      "@expo/.*|" +
      "expo-.*|" +
      "expo|" +
      "@?react-native-community/.*" +
      ")/)",
  ],
  setupFilesAfterSetup: ["<rootDir>/jest.setup.ts"],
  moduleNameMapper: {
    "^@/(.*)$": "<rootDir>/$1", // ✅ Match tsconfig paths
  },
  collectCoverageFrom: [
    "src/**/*.{ts,tsx}",
    "app/**/*.{ts,tsx}",
    "!**/*.d.ts",
    "!**/node_modules/**",
  ],
};

export default config;
```

```ts
// jest.setup.ts
import "@testing-library/react-native/extend-expect";
import { server } from "@/src/mocks/server"; // MSW server

// ✅ Start MSW before all tests
beforeAll(() => server.listen({ onUnhandledRequest: "warn" }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

// ✅ Mock expo-secure-store (native module — not available in Jest)
jest.mock("expo-secure-store", () => ({
  getItemAsync: jest.fn().mockResolvedValue(null),
  setItemAsync: jest.fn().mockResolvedValue(undefined),
  deleteItemAsync: jest.fn().mockResolvedValue(undefined),
}));

// ✅ Mock expo-font
jest.mock("expo-font", () => ({
  useFonts: () => [true, null],
}));

// ✅ Mock expo-splash-screen
jest.mock("expo-splash-screen", () => ({
  preventAutoHideAsync: jest.fn(),
  hideAsync: jest.fn(),
}));

// ✅ Mock SafeArea provider
jest.mock("react-native-safe-area-context", () => {
  const { Metrics } = jest.requireActual("react-native-safe-area-context");
  return {
    ...jest.requireActual("react-native-safe-area-context"),
    useSafeAreaInsets: () => ({
      top: Metrics.frame.x === 0 ? 47 : 0,
      bottom: Metrics.frame.x === 0 ? 34 : 0,
      left: 0,
      right: 0,
    }),
    useSafeAreaFrame: () => ({
      x: 0, y: 0, width: 390, height: 844,
    }),
  };
});

// ✅ Mock react-native-reanimated (needed for layout animations)
jest.mock("react-native-reanimated", () => {
  const Reanimated = jest.requireActual("react-native-reanimated/mock");
  return { ...Reanimated, default: Reanimated };
});
```

### DO: Component testing with React Native Testing Library

```tsx
// src/components/__tests__/PostCard.test.tsx
import { render, screen, fireEvent } from "@testing-library/react-native";
import { PostCard } from "@/src/components/PostCard";

// ✅ Mock useRouter
jest.mock("expo-router", () => ({
  useRouter: () => ({
    push: jest.fn(),
    replace: jest.fn(),
    back: jest.fn(),
  }),
}));

describe("PostCard", () => {
  it("renders title and truncated body", () => {
    render(
      <PostCard
        postId="123"
        title="Test Post"
        body="This is a long body that should be truncated"
        authorName="Jane"
      />
    );

    expect(screen.getByText("Test Post")).toBeTruthy();
    expect(screen.getByText(/This is a long body/)).toBeTruthy();
    expect(screen.getByText("Jane")).toBeTruthy();
  });

  it("navigates to post detail on press", () => {
    const mockPush = jest.fn();
    jest.spyOn(require("expo-router"), "useRouter").mockReturnValue({ push: mockPush });

    render(
      <PostCard postId="123" title="Test" body="Body" authorName="Jane" />
    );

    fireEvent.press(screen.getByText("Test"));
    expect(mockPush).toHaveBeenCalledWith("/post/123");
  });
});
```

### DO: MSW (Mock Service Worker) for API mocking

```tsx
// src/mocks/server.ts
import { setupServer } from "msw/native"; // ✅ "msw/native" for React Native
import { handlers } from "./handlers";

export const server = setupServer(...handlers);
```

```tsx
// src/mocks/handlers.ts
import { http, HttpResponse } from "msw";

const API_URL = "https://api.myapp.com";

export const handlers = [
  // ✅ Mock GET /posts
  http.get(`${API_URL}/posts`, () => {
    return HttpResponse.json([
      {
        id: "1",
        title: "Mock Post 1",
        body: "This is a mock post",
        author: { id: "u1", name: "Jane" },
        createdAt: "2024-01-01T00:00:00Z",
      },
      {
        id: "2",
        title: "Mock Post 2",
        body: "Another mock post",
        author: { id: "u2", name: "John" },
        createdAt: "2024-01-02T00:00:00Z",
      },
    ]);
  }),

  // ✅ Mock GET /posts/:id
  http.get(`${API_URL}/posts/:id`, ({ params }) => {
    const { id } = params;
    return HttpResponse.json({
      id,
      title: `Post ${id}`,
      body: "Post body content",
      author: { id: "u1", name: "Jane" },
      createdAt: "2024-01-01T00:00:00Z",
    });
  }),

  // ✅ Mock POST /posts — return created entity
  http.post(`${API_URL}/posts`, async ({ request }) => {
    const body = await request.json();
    return HttpResponse.json(
      {
        id: "new-123",
        ...(body as Record<string, unknown>),
        author: { id: "u1", name: "Jane" },
        createdAt: new Date().toISOString(),
      },
      { status: 201 }
    );
  }),

  // ✅ Mock 401 for auth testing
  http.get(`${API_URL}/auth/me`, () => {
    return new HttpResponse(null, { status: 401 });
  }),
];
```

```tsx
// src/hooks/__tests__/usePosts.test.tsx
import { renderHook, waitFor } from "@testing-library/react-native";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { usePosts } from "@/src/hooks/usePosts";
import { server } from "@/src/mocks/server";
import { http, HttpResponse } from "msw";

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false }, // ✅ Disable retries in tests
    },
  });
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
  };
}

describe("usePosts", () => {
  it("returns posts on success", async () => {
    const { result } = renderHook(() => usePosts(), {
      wrapper: createWrapper(),
    });

    expect(result.current.isLoading).toBe(true);

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toHaveLength(2);
    expect(result.current.data![0].title).toBe("Mock Post 1");
  });

  it("handles error state", async () => {
    // ✅ Override handler for this specific test
    server.use(
      http.get("https://api.myapp.com/posts", () => {
        return new HttpResponse(null, { status: 500 });
      })
    );

    const { result } = renderHook(() => usePosts(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
  });
});
```

### DO NOT: Testing mistakes

| ❌ Wrong | ✅ Right |
|----------|----------|
| `@testing-library/react` (web version) | `@testing-library/react-native` |
| `msw/node` (Node.js MSW) | `msw/native` — React Native needs the native MSW integration |
| Forgetting `transformIgnorePatterns` in jest.config.ts | Node modules like `expo-router` are ESM and MUST be transformed |
| Testing implementation details | Test behavior — what the user sees and interacts with |
| Mocking React Query at the query level | Use MSW to mock the API — tests the full data flow |
| Not wrapping hooks with `QueryClientProvider` | Hooks using `useQuery` need the provider in tests |
| `jest.useFakeTimers()` without RN adjustments | React Native timer-based APIs (Animations) break with fake timers |

---

## 9. Build & Deploy

### DO: `eas.json` configuration

```jsonc
// eas.json
{
  "cli": {
    "version": ">= 14.0.0"
  },
  "build": {
    // ✅ Development build — for expo-dev-client testing
    "development": {
      "developmentClient": true,
      "distribution": "internal",
      "env": {
        "APP_VARIANT": "development",
        "EXPO_PUBLIC_API_URL": "https://dev-api.myapp.com"
      }
    },
    // ✅ Internal preview — testers via internal distribution
    "preview": {
      "distribution": "internal",
      "env": {
        "APP_VARIANT": "preview",
        "EXPO_PUBLIC_API_URL": "https://staging-api.myapp.com"
      },
      "android": {
        "buildType": "apk"  // APK for easy distribution
      }
    },
    // ✅ Production build — App Store / Play Store
    "production": {
      "autoIncrement": true,  // ✅ Auto-increment build number
      "channel": "production",
      "env": {
        "APP_VARIANT": "production",
        "EXPO_PUBLIC_API_URL": "https://api.myapp.com"
      }
    }
  },
  "submit": {
    "production": {
      "ios": {
        "appleId": "developer@mycompany.com",
        "ascAppId": "1234567890",
        "appleTeamId": "ABCDEF1234"
      },
      "android": {
        "serviceAccountKeyPath": "./android-service-account.json",
        "track": "production"
      }
    }
  }
}
```

### DO: Environment variables

```bash
# .env (local development — DO NOT commit secrets)
EXPO_PUBLIC_API_URL=https://dev-api.myapp.com

# ✅ EXPO_PUBLIC_ prefix makes it accessible in JS code
# ❌ Variables WITHOUT EXPO_PUBLIC_ are only available in app.config.ts
```

```bash
# EAS secrets (for builds)
# Secrets are encrypted and injected at build time
eas secret:create --name EXPO_PUBLIC_API_URL --value "https://api.myapp.com" --scope project
eas secret:create --name SENTRY_AUTH_TOKEN --value "sntrys_..." --scope project
```

```tsx
// ✅ Access env vars in code
import Constants from "expo-constants";

const API_URL = Constants.expoConfig?.extra?.apiUrl;
// OR directly from process.env (EXPO_PUBLIC_ prefix):
const PUBLIC_API_URL = process.env.EXPO_PUBLIC_API_URL;
```

### DO: App versioning

```jsonc
// app.json — version and build numbers
{
  "expo": {
    "version": "1.2.0",       // User-facing version (semver)
    "ios": {
      "buildNumber": "42"      // iOS CFBundleVersion — must increment for each store submission
    },
    "android": {
      "versionCode": 42        // Android versionCode — must increment for each store submission
    }
  }
}
```

```jsonc
// eas.json — auto-increment in production
{
  "build": {
    "production": {
      "autoIncrement": true  // ✅ Automatically bumps buildNumber / versionCode
    }
  }
}
```

### DO: Over-the-air updates (expo-updates)

```jsonc
// app.json
{
  "expo": {
    "updates": {
      "enabled": true,
      "checkAutomatically": "ON_LOAD",
      "fallbackToCacheTimeout": 0,
      "url": "https://u.expo.dev/YOUR_PROJECT_ID"
    },
    "runtimeVersion": {
      "policy": "appVersion"   // ✅ Updates only delivered to matching appVersion
    }
  }
}
```

```tsx
// app/_layout.tsx — Handle OTA updates
import { useEffect } from "react";
import * as Updates from "expo-updates";

export default function RootLayout() {
  useEffect(() => {
    async function checkForUpdate() {
      try {
        const update = await Updates.checkForUpdateAsync();
        if (update.isAvailable) {
          await Updates.fetchUpdateAsync();
          // ✅ Tell user an update is ready, then reload:
          // Updates.reloadAsync();
          // Or show a banner saying "Update available. Restart to apply."
        }
      } catch (error) {
        // Silently fail — don't block the app for update checks
        console.log("Error checking for updates:", error);
      }
    }

    if (!__DEV__) {
      checkForUpdate();
    }
  }, []);

  // ... rest of layout
}
```

```bash
# Publish an OTA update via EAS
eas update --branch production --message "Fixed login crash"
```

### DO: EAS Build and Submit workflow

```bash
# 1. Build for internal distribution (testers)
eas build --platform ios --profile preview
eas build --platform android --profile preview

# 2. Submit to stores
eas submit --platform ios --profile production
eas submit --platform android --profile production

# 3. Combined: build + submit in one command
eas build --platform all --profile production --auto-submit

# 4. Run a build locally (development build)
npx expo run:ios      # Build and run on iOS Simulator
npx expo run:android  # Build and run on Android Emulator
```

### DO NOT: Build & deploy mistakes

| ❌ Wrong | ✅ Right |
|----------|----------|
| Committing `.env` with secrets | Use EAS secrets for sensitive values; `.env` only for public defaults |
| Using `expo publish` (classic build) | `eas update` — classic builds are deprecated |
| Forgetting `runtimeVersion` with `expo-updates` | Without `runtimeVersion`, OTA updates fail silently |
| `expo build:ios` or `expo build:android` | Classic build is deprecated — use `eas build` |
| Manual versionCode/bump buildNumber increments | `"autoIncrement": true` in eas.json |
| Building development clients for production | Use separate profiles: `development` (dev-client), `preview` (internal), `production` (store) |
| Forgetting App Store review screenshots | EAS Submit doesn't handle screenshots — use `fastlane deliver` or App Store Connect API |

---

## 10. Common AI Hallucinations

This section catalogs specific Expo/React Native APIs that DeepSeek invents, gets wrong, or mixes with web/Node.js patterns.

### Hallucination: Importing from non-existent React Native packages

```tsx
// ❌ HALLUCINATION — these imports DO NOT EXIST
import { FlatButton } from "react-native";           // ❌ No such component — use Pressable + Text
import { ScrollViewHorizontal } from "react-native"; // ❌ Try <ScrollView horizontal>
import { TextField } from "react-native";            // ❌ It's <TextInput>
import { NavigationBar } from "react-native";        // ❌ Doesn't exist
import { Dialog } from "react-native";               // ❌ Use <Modal> or third-party
import { Toast } from "react-native";                // ❌ Doesn't exist — use react-native-toast-message
import { Divider } from "react-native";              // ❌ Use <View style={{ height: 1, backgroundColor: "#eee" }} />
import { Badge } from "react-native";               // ❌ Doesn't exist natively
import { Card } from "react-native";                 // ❌ Doesn't exist — build with <View>
import { Icon } from "react-native";                 // ❌ Use @expo/vector-icons or react-native-vector-icons

// ❌ HALLUCINATION — Web APIs that don't exist in React Native
import { Image } from "next/image";                  // ❌ Next.js — use <Image> from "expo-image" or "react-native"
localStorage.setItem("key", "value");                // ❌ No localStorage — use AsyncStorage or SecureStore
document.querySelector("#root");                      // ❌ No DOM — React Native has no document
window.innerWidth;                                    // ❌ Use Dimensions.get("window").width
navigator.geolocation.getCurrentPosition(...);        // ❌ Use expo-location instead
```

### Hallucination: Mixing bare React Native APIs with Expo managed workflow

```tsx
// ❌ HALLUCINATION — Uses bare RN APIs that need native modules
import PushNotification from "react-native-push-notification";  // ❌ Use expo-notifications
import Camera from "react-native-camera";                        // ❌ Use expo-camera
import MapView from "react-native-maps";                         // ❌ Use expo-map or require dev build
import Video from "react-native-video";                          // ❌ Use expo-av
import AsyncStorage from "@react-native-async-storage/async-storage"; // ✅ Actually this DOES work in managed workflow
                                                                  // But use expo-secure-store for sensitive data

// ❌ HALLUCINATION — Linking from wrong package
import { Linking } from "react-native";  // ✅ This works
// But DeepSeek often generates:
import { Linking } from "expo-linking";  // ❌ expo-linking does NOT export Linking! It exports:
// - useURL()          — get the URL that opened the app
// - createURL()       — create deep link URLs
// - parse()           — parse a URL

// ✅ Correct for opening external URLs:
import { Linking } from "react-native";
Linking.openURL("https://example.com");

// ✅ Correct for handling deep links:
import { useURL } from "expo-linking";
const url = useURL(); // myapp://post/123
```

### Hallucination: expo-camera vs react-native-camera

```tsx
// ❌ HALLUCINATION — DeepSeek frequently invents react-native-camera import path
import { RNCamera } from "react-native-camera";           // ❌ Deprecated package, not in Expo SDK
import { Camera } from "react-native-camera";             // ❌ Doesn't exist

// ✅ CORRECT — expo-camera API
import { CameraView, useCameraPermissions } from "expo-camera";

export function CameraScreen() {
  const [permission, requestPermission] = useCameraPermissions();

  if (!permission) return null; // Still loading
  if (!permission.granted) {
    // Show permission request UI
    return <PermissionScreen onRequest={requestPermission} />;
  }

  return (
    <CameraView style={{ flex: 1 }} facing="back">
      {/* Camera UI here */}
    </CameraView>
  );
}

// ❌ DeepSeek also invents these non-existent expo-camera APIs:
// <Camera type={Camera.Constants.Type.back}>   → ❌ Use <CameraView facing="back">
// <Camera ref={ref => this.camera = ref}>       → ❌ Use useCameraPermissions hook
// Camera.Constants.FlashMode.on                  → ❌ Doesn't exist in expo-camera v7+
```

### Hallucination: Incorrect expo-router imports

```tsx
// ❌ HALLUCINATION — wrong import sources
import { useRouter } from "next/navigation";       // ❌ Next.js, not Expo Router
import { Link } from "next/link";                  // ❌ Next.js
import { useSearchParams } from "next/navigation"; // ❌ Next.js
import { useParams } from "next/navigation";       // ❌ Next.js
import { Stack, Tabs } from "@react-navigation/native"; // ❌ Bare React Navigation

// ✅ CORRECT imports
import { useRouter } from "expo-router";
import { Link } from "expo-router";
import { useLocalSearchParams } from "expo-router";     // ✅ NOT useSearchParams (that's Next.js)
import { useGlobalSearchParams } from "expo-router";    // ✅ For params available globally
import { Stack, Tabs, Drawer } from "expo-router";      // ✅ From expo-router, NOT react-navigation

// ❌ HALLUCINATION — navigation.navigate with wrong route names
// DeepSeek often generates React Navigation patterns:
navigation.navigate("PostDetail", { id: "123" });   // ❌ Bare RN style
navigation.navigate("Post", { params: { id } });    // ❌ DeepSeek mix of patterns

// ✅ CORRECT — Expo Router uses file paths, not screen names:
router.push({ pathname: "/post/[id]", params: { id: "123" } });
router.push("/post/123");
```

### Hallucination: StyleSheet not imported or used incorrectly

```tsx
// ❌ HALLUCINATION — DeepSeek forgets to import StyleSheet
export function MyComponent() {
  return (
    <View style={styles.container}>  {/* ❌ Compile error: styles is not defined */}
      <Text>Hello</Text>
    </View>
  );
}
const styles = StyleSheet.create({   // ❌ StyleSheet is not imported
  container: { padding: 16 },
});

// ❌ HALLUCINATION — CSS syntax in React Native
const styles = StyleSheet.create({
  container: {
    display: "flex",           // ❌ RN doesn't use CSS `display`. Layout is always flex.
    boxSizing: "border-box",   // ❌ CSS property — doesn't exist in RN
    ":hover": { opacity: 0.8 }, // ❌ Pseudo-classes don't exist in RN
    "& > *": { margin: 4 },    // ❌ CSS nesting doesn't exist in RN
  },
});

// ✅ CORRECT
import { StyleSheet } from "react-native";  // ✅ MUST import

const styles = StyleSheet.create({
  container: {
    flexDirection: "column",   // ✅ RN uses this, not display:flex
    padding: 16,
    backgroundColor: "#fff",
  },
});
```

### Hallucination: Node.js APIs in React Native

```tsx
// ❌ HALLUCINATION — Node.js APIs that DON'T work in React Native
import fs from "fs";                          // ❌ No filesystem module — use expo-file-system
import path from "path";                      // ❌ No path module — use string manipulation
import crypto from "crypto";                  // ❌ Use expo-crypto
import os from "os";                          // ❌ Use expo-device
import { Buffer } from "buffer";             // ❌ Buffer is limited in RN — use expo-crypto for hashing
import { createServer } from "net";          // ❌ Cannot create TCP servers in React Native
import child_process from "child_process";   // ❌ Cannot spawn child processes
import cluster from "cluster";               // ❌ Cannot fork/cluster in RN

// ✅ CORRECT alternatives
import * as FileSystem from "expo-file-system";
// Replace path.join:
const filePath = `${FileSystem.documentDirectory}images/${filename}`;
// Hashing:
import * as Crypto from "expo-crypto";
const digest = await Crypto.digestStringAsync(
  Crypto.CryptoDigestAlgorithm.SHA256,
  "my string"
);
```

### Hallucination: React Native components used incorrectly

```tsx
// ❌ HALLUCINATION — Web-style event handlers
<TextInput onChange={(e) => setName(e.target.value)} />
// ❌ e.target.value is Web API — React Native uses:
<TextInput onChangeText={(text) => setName(text)} />

// ❌ HALLUCINATION — onClick (Web event)
<Pressable onClick={() => {}} />
// ✅ onPress (React Native event)
<Pressable onPress={() => {}} />

// ❌ HALLUCINATION — <img> tag
<img src="https://..." alt="photo" />
// ✅ <Image> component
<Image source={{ uri: "https://..." }} style={{ width: 100, height: 100 }} />

// ❌ HALLUCINATION — <button> tag
<button type="submit">Submit</button>
// ✅ <Pressable> or <TouchableOpacity>
<Pressable onPress={handleSubmit}><Text>Submit</Text></Pressable>

// ❌ HALLUCINATION — <input type="checkbox" />
<input type="checkbox" checked={done} />
// ✅ <Switch> component
<Switch value={done} onValueChange={setDone} />

// ❌ HALLUCINATION — <select><option>
<select><option value="a">A</option></select>
// ✅ Use @react-native-picker/picker or a custom bottom sheet
```

### Hallucination: Expo-specific import/API mistakes

```tsx
// ❌ HALLUCINATION — Wrong expo-linking API
import { useDeepLink } from "expo-linking";      // ❌ Doesn't exist
import { onDeepLink } from "expo-linking";       // ❌ Doesn't exist
import { parseDeepLink } from "expo-linking";    // ❌ Doesn't exist

// ✅ Correct expo-linking API:
import * as Linking from "expo-linking";          // ✅
import { useURL } from "expo-linking";            // ✅ Get URL that opened the app
import { createURL } from "expo-linking";         // ✅ Create deep links: myapp://path
import { parse } from "expo-linking";             // ✅ Parse a deep link URL

// ❌ HALLUCINATION — Wrong expo-notifications API
import { registerForPushNotifications } from "expo-notifications";  // ❌ Doesn't exist
import { getPushToken } from "expo-notifications";                  // ❌ Wrong function name

// ✅ Correct expo-notifications API:
import * as Notifications from "expo-notifications";
Notifications.getExpoPushTokenAsync();   // ✅ Get push token
Notifications.scheduleNotificationAsync({ // ✅ Schedule local notification
  content: { title: "Hello", body: "World" },
  trigger: { seconds: 5 },
});

// ❌ HALLUCINATION — Wrong expo-secure-store API
import { setItem, getItem } from "expo-secure-store";  // ❌ Wrong — it's AsyncStorage API
// ✅ Correct:
import * as SecureStore from "expo-secure-store";
await SecureStore.setItemAsync("key", "value");
await SecureStore.getItemAsync("key");
await SecureStore.deleteItemAsync("key");
```

### Hallucination: Confusing `navigation.navigate` with Expo Router push

```tsx
// ❌ HALLUCINATION — Bare React Navigation patterns with Expo Router
import { useNavigation } from "@react-navigation/native";

export function MyScreen() {
  const navigation = useNavigation();

  // ❌ These don't work with Expo Router
  navigation.navigate("PostDetail", { id: "123" });  // Screen name not file path
  navigation.push("Profile");                          // Expo Router uses router.push()
  navigation.goBack();                                // ✅ This DOES work, but...
  navigation.dangerouslyGetParent();                   // ❌ Don't do this with Expo Router
}

// ✅ CORRECT — Use Expo Router's useRouter()
import { useRouter } from "expo-router";

export function MyScreen() {
  const router = useRouter();

  router.push({ pathname: "/post/[id]", params: { id: "123" } });
  router.replace("/profile");
  router.back();
  router.dismiss(); // dismiss modal
}
```

### Hallucination: Direct state mutation patterns

```tsx
// ❌ HALLUCINATION — DeepSeek often generates mutable patterns
const [posts, setPosts] = useState<Post[]>([]);

// ❌ Mutating state directly:
posts.push(newPost);
setPosts(posts); // ❌ Same reference — won't trigger re-render

// ❌ Mutating nested object:
const [user, setUser] = useState({ name: "Jane", profile: { bio: "" } });
user.profile.bio = "New bio"; // ❌ Mutation
setUser(user);

// ✅ CORRECT — Immutable updates
setPosts(prev => [...prev, newPost]);
setUser(prev => ({ ...prev, profile: { ...prev.profile, bio: "New bio" } }));
```

### Comprehensive table of hallucinated APIs vs correct APIs

| ❌ Hallucinated API | ✅ Correct API |
|---------------------|---------------|
| `import { useRouter } from "next/navigation"` | `import { useRouter } from "expo-router"` |
| `import { Link } from "next/link"` | `import { Link } from "expo-router"` |
| `import { Stack } from "@react-navigation/native-stack"` | `import { Stack } from "expo-router"` |
| `navigation.navigate("ScreenName")` | `router.push("/file-path")` |
| `useSearchParams()` | `useLocalSearchParams()` — Expo Router |
| `useParams()` | `useLocalSearchParams()` — Expo Router |
| `import { ImageResponse } from "next/og"` | Not applicable — React Native doesn't have OG images |
| `cookies().get("token")` | `SecureStore.getItemAsync("token")` |
| `headers()` | No equivalent — use `expo-constants` for app metadata |
| `revalidatePath("/posts")` | `queryClient.invalidateQueries({ queryKey: postKeys.lists() })` |
| `fetch("https://api...", { next: { revalidate: 60 } })` | React Query `staleTime` + `gcTime` |
| `"use client"` directive | All Expo Router components are already client-side |
| `"use server"` directive | Server Actions don't exist in React Native |
| `router.refresh()` | React Query `refetch()` or `queryClient.invalidateQueries()` |
| `useFormState()` / `useActionState()` | `useMutation` from React Query |
| `generateMetadata()` | Set `<Stack.Screen options={{ title: "..." }} />` |
| `notFound()` from `next/navigation` | `<Stack.Screen name="+not-found" />` in file routing |
| `redirect("/login")` from `next/navigation` | `router.replace("/(auth)/login")` |
| `import { Image } from "next/image"` | `import { Image } from "expo-image"` (or `react-native`) |
| `getServerSideProps` / `getStaticProps` | React Query `useQuery` — all data is fetched client-side |

---

## 11. Platform-Specific Code

### DO: `Platform.OS` for conditional logic

```tsx
import { Platform, StyleSheet } from "react-native";

const styles = StyleSheet.create({
  // ✅ Platform.select for style values
  shadow: Platform.select({
    ios: {
      shadowColor: "#000",
      shadowOffset: { width: 0, height: 2 },
      shadowOpacity: 0.15,
      shadowRadius: 6,
    },
    android: {
      elevation: 4,
    },
    default: {
      // Web or other platforms
    },
  }),

  // ✅ Platform.OS for conditional logic outside styles
  headerHeight: Platform.OS === "ios" ? 88 : 64,
});

// ✅ Runtime platform checks
export function PlatformAwareComponent() {
  if (Platform.OS === "ios") {
    return <IOSSpecificUI />;
  }
  if (Platform.OS === "android") {
    return <AndroidSpecificUI />;
  }
  return null;
}

// ✅ Platform.Version for version-specific code
if (Platform.OS === "ios" && parseInt(Platform.Version as string, 10) >= 16) {
  // iOS 16+ specific behavior
}
if (Platform.OS === "android" && (Platform.Version as number) >= 33) {
  // Android 13+ (API 33)
}
```

### DO: Platform-specific file extensions

```
src/
├── components/
│   ├── DatePicker.tsx           # Shared interface / type exports
│   ├── DatePicker.ios.tsx       # ✅ iOS implementation — auto-selected on iOS
│   ├── DatePicker.android.tsx   # ✅ Android implementation — auto-selected on Android
│   └── DatePicker.native.tsx    # ✅ Both iOS and Android (but not web)
└── utils/
    ├── fileHandler.ts           # Fallback / default
    ├── fileHandler.ios.ts       # Uses iOS document picker
    └── fileHandler.android.ts   # Uses Android SAF (Storage Access Framework)
```

```tsx
// src/components/DatePicker.tsx — Shared type
export interface DatePickerProps {
  date: Date;
  onDateChange: (date: Date) => void;
  minimumDate?: Date;
  maximumDate?: Date;
}

// Consumers just import from "DatePicker" and React Native resolves the platform file:
import { DatePicker } from "@/src/components/DatePicker";
```

### DO: SafeAreaView vs SafeAreaProvider vs SafeAreaView

```tsx
// ❌ HALLUCINATION — DeepSeek frequently mixes these up

// ❌ Option 1: react-native's SafeAreaView (ONLY works on iOS)
import { SafeAreaView } from "react-native";
// Problem: Does NOTHING on Android. Not recommended.

// ❌ Option 2: Both nested — unnecessary
import { SafeAreaView } from "react-native";
import { SafeAreaProvider } from "react-native-safe-area-context";
// <SafeAreaProvider>
//   <SafeAreaView> ...   ← Don't nest SafeAreaView inside SafeAreaProvider

// ✅ CORRECT: Use react-native-safe-area-context for cross-platform
import { SafeAreaProvider, SafeAreaView } from "react-native-safe-area-context";

export default function App() {
  return (
    <SafeAreaProvider>
      <SafeAreaView style={{ flex: 1 }} edges={["top", "bottom"]}>
        {/* Content that respects safe areas on BOTH platforms */}
      </SafeAreaView>
    </SafeAreaProvider>
  );
}

// ✅ OR use useSafeAreaInsets for manual insets
import { useSafeAreaInsets } from "react-native-safe-area-context";

export function CustomHeader() {
  const insets = useSafeAreaInsets();
  return (
    <View style={{ paddingTop: insets.top, height: 44 + insets.top }}>
      <Text>Header</Text>
    </View>
  );
}
```

### DO: Keyboard avoiding

```tsx
// ✅ KeyboardAvoidingView — simplest approach
import { KeyboardAvoidingView, Platform, TextInput } from "react-native";

export function FormScreen() {
  return (
    <KeyboardAvoidingView
      style={{ flex: 1 }}
      behavior={Platform.OS === "ios" ? "padding" : "height"}
      keyboardVerticalOffset={Platform.OS === "ios" ? 88 : 0}
    >
      {/* Form fields */}
      <TextInput placeholder="Email" />
      <TextInput placeholder="Password" />
    </KeyboardAvoidingView>
  );
}

// ✅ react-native-keyboard-aware-scroll-view — better for scrollable forms
// npx expo install react-native-keyboard-aware-scroll-view
import { KeyboardAwareScrollView } from "react-native-keyboard-aware-scroll-view";

export function LongFormScreen() {
  return (
    <KeyboardAwareScrollView
      style={{ flex: 1 }}
      contentContainerStyle={{ padding: 16 }}
      enableOnAndroid
      extraScrollHeight={20}
    >
      {/* Long form content — auto-scrolls to focused input */}
    </KeyboardAwareScrollView>
  );
}
```

### DO: StatusBar

```tsx
// ❌ HALLUCINATION — react-native's StatusBar used directly
import { StatusBar } from "react-native";

// ✅ Use expo-status-bar for better Expo integration
import { StatusBar } from "expo-status-bar";
import { useAppTheme } from "@/src/hooks/useTheme";

export function MyScreen() {
  const { isDark } = useAppTheme();

  return (
    <View style={{ flex: 1 }}>
      {/* ✅ expo-status-bar auto-manages padding, background, etc. */}
      <StatusBar style={isDark ? "light" : "dark"} />

      {/* Options:
        style="auto"  — follows system
        style="light" — white text (for dark backgrounds)
        style="dark"  — dark text (for light backgrounds)
        hidden         — hides the status bar
        translucent    — transparent background (Android)
      */}
    </View>
  );
}

// ✅ Set StatusBar globally in root _layout.tsx
// app/_layout.tsx
import { StatusBar } from "expo-status-bar";

export default function RootLayout() {
  return (
    <>
      <StatusBar style="auto" />
      <Stack screenOptions={{ headerShown: false }} />
    </>
  );
}
```

### DO: Platform-specific API availability checks

```tsx
// ✅ Check if a native module is available before using it
import { Platform } from "react-native";

async function useBiometrics() {
  // expo-local-authentication works on both platforms
  // But some features are platform-specific:

  const hasHardware = await LocalAuthentication.hasHardwareAsync();
  if (!hasHardware) {
    // Show fallback (password/pin)
    return;
  }

  // ✅ Specific to platform
  if (Platform.OS === "android") {
    // Android-specific biometric types
    const enrolled = await LocalAuthentication.isEnrolledAsync();
    // Android APIs like BiometricPrompt specifics
  }

  if (Platform.OS === "ios") {
    // iOS-specific: check biometryType (FaceID vs TouchID)
    const types = await LocalAuthentication.supportedAuthenticationTypesAsync();
    const hasFaceID = types.includes(
      LocalAuthentication.AuthenticationType.FACIAL_RECOGNITION
    );
  }
}
```

### DO NOT: Platform-specific mistakes

| ❌ Wrong | ✅ Right |
|----------|----------|
| `import { SafeAreaView } from "react-native"` | `import { SafeAreaProvider, SafeAreaView } from "react-native-safe-area-context"` |
| `overflow: "visible"` shadow on Android | Android shadows use `elevation` and clip `overflow: "hidden"` — use separate views |
| `import { StatusBar } from "react-native"` | `import { StatusBar } from "expo-status-bar"` — Expo-compatible, handles Expo Go correctly |
| Assuming all devices have a notch | Always wrap in `SafeAreaView` with `edges={["top", "bottom"]}` |
| Hardcoding keyboard height offset | Use `KeyboardAvoidingView` + `keyboardVerticalOffset` based on header height |
| `Platform.select({ web: {} })` in RN-only app | If no web target, don't include web — it's dead code |
| Not testing on both platforms | Behavior differences: shadows, fonts, pickers, haptics, permissions, back button |
| `Platform.OS === "ios" ? "SF Pro" : "Roboto"` | Use `Platform.select` with a `default` fallback — always provide defaults |

---

## Quick Reference: Final Checklist (Before Shipping AI-Generated Expo Code)

### Imports
- [ ] `useRouter` imported from `"expo-router"` (NOT `"next/navigation"` or `"@react-navigation/native"`)
- [ ] `Link` imported from `"expo-router"` (NOT `"next/link"`)
- [ ] `Stack`, `Tabs`, `Drawer` imported from `"expo-router"` (NOT `"@react-navigation/*"`)
- [ ] `StyleSheet` imported from `"react-native"` in EVERY file with styles
- [ ] `StatusBar` imported from `"expo-status-bar"` (NOT `"react-native"`)
- [ ] `SafeAreaView` imported from `"react-native-safe-area-context"` (NOT `"react-native"`)
- [ ] `Image` from `"react-native"` or `"expo-image"` — never `"next/image"`
- [ ] No Node.js imports: `fs`, `path`, `crypto`, `os`, `child_process`, `cluster`
- [ ] No web API usage: `localStorage`, `document`, `window`, `navigator.geolocation`

### Routing
- [ ] No manual `NavigationContainer` creation — Expo Router handles it
- [ ] Route groups use `(groupName)` with a `_layout.tsx` inside
- [ ] Dynamic routes use `[param].tsx` with `useLocalSearchParams()` to read params
- [ ] Navigation uses `router.push("/file-path")` not `navigation.navigate("ScreenName")`
- [ ] Root `_layout.tsx` imports `"react-native-reanimated"` FIRST if using drawer/shared transitions
- [ ] Auth gate checks `isLoading` before redirecting

### Data
- [ ] API calls use the configured `apiClient` (axios instance), not raw `fetch`
- [ ] Tokens stored in `expo-secure-store`, NEVER `AsyncStorage`
- [ ] React Query hooks use `queryKey` constants, not inline arrays
- [ ] Mutations call `queryClient.invalidateQueries` on success
- [ ] `enabled` prop used for dependent queries

### Styling
- [ ] All styles use `StyleSheet.create()` — no inline style objects
- [ ] Shadows use `shadowColor`/`shadowOffset`/`shadowOpacity`/`shadowRadius` (iOS) + `elevation` (Android)
- [ ] No CSS syntax: `display: flex`, `:hover`, `box-shadow`, `& > *`
- [ ] `SafeAreaView` wraps top-level screens with proper `edges`

### State
- [ ] Zustand for global client state, React Query for server state
- [ ] No Redux unless specific complex state machine requirement
- [ ] No `persist` middleware on auth-sensitive stores

### Platform
- [ ] `Platform.select()` for platform-specific styles
- [ ] Platform files (`.ios.tsx`, `.android.tsx`) for diverging implementations
- [ ] `KeyboardAvoidingView` with `behavior="padding"` on iOS for forms
- [ ] `StatusBar` style matches theme (light/dark)

### Testing
- [ ] Jest uses `jest-expo` preset
- [ ] `transformIgnorePatterns` configured for ESM node_modules
- [ ] MSW uses `msw/native` (NOT `msw/node`)
- [ ] `expo-secure-store`, `expo-font`, `expo-splash-screen` mocked in setup
- [ ] React Query `retry: false` in test wrapper

### Build
- [ ] `eas.json` has separate profiles: `development`, `preview`, `production`
- [ ] Secrets use EAS secrets (`eas secret:create`), never committed to repo
- [ ] `EXPO_PUBLIC_` prefix for env vars accessible in JS
- [ ] `autoIncrement: true` on production profile
- [ ] `runtimeVersion` set for OTA updates
- [ ] `expo-router` plugin in `app.json` plugins array

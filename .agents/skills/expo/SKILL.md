---
name: expo
description: Expo/React Native (managed workflow) development standards — project layout, Expo Router, platform-specific patterns, testing, and managed workflow boundaries. Load when writing React Native code in the mobile/ directory.
---

# Expo / React Native Standards

You are writing React Native code using Expo managed workflow with
Expo Router for navigation. Follow these conventions exactly.

## Project Layout (Monorepo)

```
mobile/
├── app/                          # Expo Router file-based routing
│   ├── _layout.tsx               # Root layout (providers, fonts, splash)
│   ├── index.tsx                 # Home screen
│   ├── (auth)/                   # Route group for auth screens
│   │   ├── _layout.tsx
│   │   ├── login.tsx
│   │   └── signup.tsx
│   ├── (tabs)/                   # Route group for tabs
│   │   ├── _layout.tsx           # Tab navigator layout
│   │   ├── index.tsx             # Tab 1
│   │   ├── profile.tsx           # Tab 2
│   │   └── [id].tsx              # Dynamic route
│   └── modal.tsx                 # Modal screen
├── src/
│   ├── components/               # Reusable components
│   │   ├── ui/                   # Generic UI components
│   │   └── features/             # Feature-specific components
│   ├── hooks/                    # Custom hooks
│   │   ├── useAuth.ts
│   │   ├── useTheme.ts
│   │   └── useUsers.ts           # TanStack Query hooks
│   ├── stores/                   # Zustand stores
│   │   ├── authStore.ts
│   │   └── themeStore.ts
│   ├── lib/                      # Utilities
│   │   ├── api.ts                # Axios client
│   │   └── constants.ts
│   └── types/                    # TypeScript types
│       └── index.ts
├── assets/                       # Static assets (images, fonts)
├── app.json                      # Expo config
├── eas.json                      # EAS Build config
├── tsconfig.json
├── package.json
└── babel.config.js
```

## Managed Workflow Boundaries

### DO — Use expo-* packages FIRST

Before considering any native module, check if Expo has it:

| Need | Expo Package |
|------|-------------|
| Camera | `expo-camera` |
| Image picker | `expo-image-picker` |
| Notifications | `expo-notifications` |
| Secure storage | `expo-secure-store` |
| Location | `expo-location` |
| Haptics | `expo-haptics` |
| File system | `expo-file-system` |
| Audio/Video | `expo-av` |
| Barcode scanning | `expo-barcode-scanner` |
| Splash screen | `expo-splash-screen` |
| Device info | `expo-device` |
| Network info | `expo-network` |

### DO NOT — These require ejecting (expo prebuild)

| ❌ Need | ✅ Alternatives |
|---|---|
| Bluetooth classic | `expo-bluetooth` (BLE only) or eject |
| Background audio playback (iOS) | Use `expo-av` with audio background mode |
| Custom native UI components (UIKit/Android Views) | Eject or use EAS Build with a config plugin |
| Native in-app purchases (not revenuecat) | Use `expo-in-app-purchases` or eject |
| Low-level Bluetooth LE (beyond expo-bluetooth) | Eject |

### When You Need Native Code

If you MUST use native code, use **EAS Build with a config plugin** rather
than local `expo prebuild`. This keeps the developer workflow managed while
the build handles native customization.

## Expo Router

### DO — Root layout with providers

```tsx
// app/_layout.tsx
import { Stack } from "expo-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useFonts } from "expo-font";
import * as SplashScreen from "expo-splash-screen";
import { useEffect } from "react";

SplashScreen.preventAutoHideAsync();

const queryClient = new QueryClient();

export default function RootLayout() {
  const [loaded] = useFonts({
    "Inter-Regular": require("../assets/fonts/Inter-Regular.ttf"),
    "Inter-Bold": require("../assets/fonts/Inter-Bold.ttf"),
  });

  useEffect(() => {
    if (loaded) {
      SplashScreen.hideAsync();
    }
  }, [loaded]);

  if (!loaded) return null;

  return (
    <QueryClientProvider client={queryClient}>
      <Stack screenOptions={{ headerShown: false }} />
    </QueryClientProvider>
  );
}
```

### DO — Tab layout

```tsx
// app/(tabs)/_layout.tsx
import { Tabs } from "expo-router";
import { Ionicons } from "@expo/vector-icons";

export default function TabLayout() {
  return (
    <Tabs>
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

### DO — Dynamic routes

```tsx
// app/(tabs)/[id].tsx
import { useLocalSearchParams } from "expo-router";
import { Text, View } from "react-native";

export default function ItemScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  return (
    <View>
      <Text>Item: {id}</Text>
    </View>
  );
}
```

### DO NOT — Expo Router mistakes

| ❌ Wrong | ✅ Right |
|---|---|
| `import { useLocalSearchParams } from "expo-router/build/..."`| `import { useLocalSearchParams } from "expo-router"` |
| `router.push({ pathname: "/profile", params: { id } })` | `router.push(\`/profile/${id}\`)` or `router.push({ pathname: "/profile/[id]", params: { id } })` |
| Mix `react-navigation` APIs directly | Use Expo Router's `router`, `Link`, `useLocalSearchParams` |

## Data Fetching

### DO — TanStack Query setup

```tsx
// src/hooks/useUsers.ts
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "../lib/api";

export function useUsers() {
  return useQuery({
    queryKey: ["users"],
    queryFn: () => api.get("/users").then((r) => r.data),
  });
}

export function useCreateUser() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateUserInput) =>
      api.post("/users", data).then((r) => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
    },
  });
}
```

### DO — Axios client with auth token

```tsx
// src/lib/api.ts
import axios from "axios";
import * as SecureStore from "expo-secure-store";

export const api = axios.create({
  baseURL: process.env.EXPO_PUBLIC_API_URL,
  timeout: 10000,
});

api.interceptors.request.use(async (config) => {
  const token = await SecureStore.getItemAsync("token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      // Handle token refresh or logout
      await SecureStore.deleteItemAsync("token");
    }
    return Promise.reject(error);
  }
);
```

### DO NOT — Data fetching mistakes

| ❌ Wrong | ✅ Right |
|---|---|
| Fetch in `useEffect` for data that doesn't change | Use TanStack Query |
| Store tokens in AsyncStorage | Use `expo-secure-store` |
| Hardcode API URL in components | Use `process.env.EXPO_PUBLIC_API_URL` |

## State Management

### Decision Tree

```
Shared state across many components?
├── Yes, simple global state → Zustand
├── Yes, auth state → React Context (AuthProvider)
└── No → local useState

Server state (from API)?
└── Always TanStack Query (never store server data in Zustand)
```

### DO — Zustand store

```tsx
// src/stores/authStore.ts
import { create } from "zustand";

interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  login: (user: User) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: false,
  user: null,
  login: (user) => set({ user, isAuthenticated: true }),
  logout: () => set({ user: null, isAuthenticated: false }),
}));
```

### DO NOT — State management mistakes

| ❌ Wrong | ✅ Right |
|---|---|
| Use Redux by default | Redux only if specifically needed; prefer Zustand |
| Store API response data in Zustand | Use TanStack Query for server state |
| Create Context for everything | Only Context for truly app-wide state (auth, theme) |

## Styling

### DO — Always use StyleSheet.create

```tsx
import { StyleSheet, View, Text } from "react-native";

export function UserCard({ user }: { user: User }) {
  return (
    <View style={styles.card}>
      <Text style={styles.name}>{user.name}</Text>
      <Text style={styles.email}>{user.email}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  card: {
    padding: 16,
    borderRadius: 8,
    backgroundColor: "#fff",
    shadowColor: "#000",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 2, // Android shadow
  },
  name: { fontSize: 18, fontWeight: "600" },
  email: { fontSize: 14, color: "#666" },
});
```

### DO — Platform-specific styles

```tsx
import { Platform, StyleSheet } from "react-native";

const styles = StyleSheet.create({
  shadow: {
    ...Platform.select({
      ios: {
        shadowColor: "#000",
        shadowOffset: { width: 0, height: 2 },
        shadowOpacity: 0.1,
        shadowRadius: 4,
      },
      android: {
        elevation: 4,
      },
    }),
  },
});
```

### DO NOT — Styling mistakes

| ❌ Wrong | ✅ Right |
|---|---|
| Inline styles as plain objects | `StyleSheet.create` (memoized, validated) |
| CSS `gap` without checking RN version | `gap` available in RN 0.71+, Expo SDK 48+ |
| `vw`/`vh` units | Use `Dimensions` or percentage-based flex |
| CSS `:hover`, `:focus` | These don't exist in React Native |

## Platform-Specific Code

### DO — Platform file extensions

```
button.tsx         # Shared implementation
button.ios.tsx     # iOS-specific (automatically used on iOS)
button.android.tsx # Android-specific (automatically used on Android)
```

### DO — SafeAreaView

```tsx
import { SafeAreaView } from "react-native-safe-area-context";

export default function Screen() {
  return (
    <SafeAreaView edges={["top"]} style={{ flex: 1 }}>
      {/* Content */}
    </SafeAreaView>
  );
}
```

### DO NOT — Platform mistakes

| ❌ Wrong | ✅ Right |
|---|---|
| `SafeAreaView` from `react-native` | Use `react-native-safe-area-context` |
| `StatusBar` on iOS with `backgroundColor` | `backgroundColor` is Android-only; use style for iOS |
| Keyboard avoiding without behavior | `<KeyboardAvoidingView behavior={Platform.OS === "ios" ? "padding" : "height"}>` |

## Testing

### DO — Component test with RNTL

```tsx
import { render, screen, fireEvent } from "@testing-library/react-native";
import { UserCard } from "../UserCard";

test("renders user name and email", () => {
  render(<UserCard user={{ name: "Alice", email: "alice@example.com" }} />);
  expect(screen.getByText("Alice")).toBeTruthy();
  expect(screen.getByText("alice@example.com")).toBeTruthy();
});
```

### DO — Jest config for Expo

```js
// jest.config.js
module.exports = {
  preset: "jest-expo",
  transformIgnorePatterns: [
    "node_modules/(?!((jest-)?react-native|@react-native(-community)?)|expo(nent)?|@expo(nent)?/.*|@expo-google-fonts/.*|react-navigation|@react-navigation/.*|@sentry/react-native|native-base|react-native-svg)",
  ],
};
```

## Build & Deploy

### DO — EAS Build

```bash
# Development build (for expo-dev-client)
eas build --profile development --platform ios

# Preview build (internal testing)
eas build --profile preview --platform all

# Production build
eas build --profile production --platform all

# Submit to stores
eas submit --platform ios
eas submit --platform android
```

### DO — Environment variables

```json
// eas.json
{
  "build": {
    "production": {
      "env": {
        "EXPO_PUBLIC_API_URL": "https://api.example.com"
      }
    },
    "preview": {
      "env": {
        "EXPO_PUBLIC_API_URL": "https://staging-api.example.com"
      }
    }
  }
}
```

Use `EXPO_PUBLIC_` prefix for values accessible in app code:
```tsx
const apiUrl = process.env.EXPO_PUBLIC_API_URL;
```

## Common AI Hallucinations — Complete Reference

### Non-Existent React Native APIs

| ❌ DeepSeek says | ✅ Reality |
|---|---|
| `<div>`, `<span>`, `<p>` | `<View>`, `<Text>` |
| `className` | `style={styles.xxx}` |
| `onClick` | `onPress` |
| `localStorage` | `expo-secure-store` or `AsyncStorage` |
| `navigator.geolocation` | `expo-location` |
| `fetch` with `file://` URLs | Not supported; use `expo-file-system` |
| CSS files (`.css`) | StyleSheet.create or NativeWind |

### Non-Existent Expo APIs

| ❌ DeepSeek says | ✅ Reality |
|---|---|
| `import { Camera } from "react-native-camera"` | Use `expo-camera` (managed workflow) |
| `import { Camera } from "expo-camera"` with `<Camera ref={...}>` | `expo-camera` v7+ uses `useCameraPermissions()` + `<CameraView>` |
| `Permissions.askAsync(Permissions.CAMERA)` | Use `useCameraPermissions()` hook (Expo SDK 50+) |
| `Notifications.presentLocalNotificationAsync({...})` (old API) | `expo-notifications` v0.28+ uses new API |
| `SecureStore.setItemAsync("key", "value", {})` | `SecureStore.setItemAsync("key", "value")` — no third argument |
| `expo-updates` with `Updates.reload()` direct | Must import: `import * as Updates from "expo-updates"` |

### Web/Frontend Patterns That Don't Work in RN

| ❌ Pattern | ✅ React Native Alternative |
|---|---|
| `useEffect` with cleanup watching `window` | No `window`; use `AppState` from `react-native` |
| `document.querySelector` | Use refs: `useRef<View>(null)` |
| CSS Grid | Use `Flexbox` |
| `form onSubmit` | `<TextInput onSubmitEditing={...}>` |
| `<img src={url}>` | `<Image source={{ uri: url }}>` |
| `fetch` with `credentials: "include"` | Not available; use token-based auth |
| WebSockets via `new WebSocket(url)` | Available in RN but subtle differences |

### Import Traps

| ❌ DeepSeek says | ✅ Reality |
|---|---|
| `import { useRouter } from "next/router"` | `import { router } from "expo-router"` |
| `import { Link } from "react-router-native"` | `import { Link } from "expo-router"` |
| `import { StyleSheet } from "react-native-web"` | `import { StyleSheet } from "react-native"` |
| `import fs from "fs"` in RN code | Node.js APIs don't exist in React Native |

### Pre-Ship Checklist

Before claiming a mobile task is complete:
- [ ] Works on both iOS and Android (tested or Platform.select used)
- [ ] SafeAreaView wraps every screen
- [ ] No hardcoded API URLs (use `EXPO_PUBLIC_` env vars)
- [ ] Tokens stored in `expo-secure-store`, not AsyncStorage
- [ ] All text in `<Text>` components (no raw strings in `<View>`)
- [ ] `StyleSheet.create` for all styles (no inline plain objects)
- [ ] Keyboard avoiding for forms
- [ ] `npx tsc --noEmit` passes
- [ ] `eas build` succeeds for preview profile

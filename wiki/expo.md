# expo

> **Generated** from `.agents/skills/expo/SKILL.md` — do not edit by hand.
> Source: [SKILL.md](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md)

Expo/React Native (managed workflow) development standards — project layout, Expo Router, platform-specific patterns, testing, and managed workflow boundaries. Load when writing React Native code in the mobile/ directory.

**Category:** stack · **Tags:** expo, mobile

## Sections

- [Project Layout (Monorepo)](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#project-layout-monorepo)
  - [`.nvmrc` — pin Node version](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#nvmrc--pin-node-version)
- [Managed Workflow Boundaries](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#managed-workflow-boundaries)
  - [DO — Use expo-* packages FIRST](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#do--use-expo--packages-first)
  - [DO NOT — These require ejecting (expo prebuild)](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#do-not--these-require-ejecting-expo-prebuild)
  - [When You Need Native Code](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#when-you-need-native-code)
- [Expo Router](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#expo-router)
  - [DO — Root layout with providers](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#do--root-layout-with-providers)
  - [DO — Tab layout](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#do--tab-layout)
  - [DO — Dynamic routes](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#do--dynamic-routes)
  - [DO NOT — Expo Router mistakes](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#do-not--expo-router-mistakes)
- [Data Fetching](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#data-fetching)
  - [DO — TanStack Query setup](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#do--tanstack-query-setup)
  - [DO — Axios client with auth token](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#do--axios-client-with-auth-token)
  - [DO NOT — Data fetching mistakes](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#do-not--data-fetching-mistakes)
- [State Management](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#state-management)
  - [Decision Tree](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#decision-tree)
  - [DO — Zustand store](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#do--zustand-store)
  - [DO NOT — State management mistakes](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#do-not--state-management-mistakes)
- [Styling](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#styling)
  - [DO — Always use StyleSheet.create](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#do--always-use-stylesheetcreate)
  - [DO — Platform-specific styles](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#do--platform-specific-styles)
  - [DO NOT — Styling mistakes](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#do-not--styling-mistakes)
- [Platform-Specific Code](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#platform-specific-code)
  - [DO — Platform file extensions](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#do--platform-file-extensions)
  - [DO — SafeAreaView](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#do--safeareaview)
  - [DO NOT — Platform mistakes](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#do-not--platform-mistakes)
- [Testing](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#testing)
  - [DO — Component test with RNTL](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#do--component-test-with-rntl)
  - [DO — Jest config for Expo](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#do--jest-config-for-expo)
- [Build & Deploy](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#build--deploy)
  - [DO — EAS Build](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#do--eas-build)
  - [DO — Environment variables](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#do--environment-variables)
- [Common AI Hallucinations — Complete Reference](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#common-ai-hallucinations--complete-reference)
  - [Non-Existent React Native APIs](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#non-existent-react-native-apis)
  - [Non-Existent Expo APIs](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#non-existent-expo-apis)
  - [Web/Frontend Patterns That Don't Work in RN](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#webfrontend-patterns-that-dont-work-in-rn)
  - [Import Traps](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#import-traps)
  - [Pre-Ship Checklist](https://github.com/eparodi/deepcut-skills/blob/main/.agents/skills/expo/SKILL.md#pre-ship-checklist)


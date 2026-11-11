# Personal Cursor Assistant Releases Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Maintain the approved provider-name feature in 2tdbbfgk2m-beep/cursor-byok and update the installed Apple Silicon app from personal releases.

**Architecture:** Keep upstream ancestry and personal changes on main. Publish tested arm64 bundles with SHA-256 manifests on main pushes. A daily synchronization workflow merges upstream only when conflict-free, runs regression checks, pushes the merge, and explicitly dispatches the release workflow because GITHUB_TOKEN pushes do not trigger other workflows.

**Tech Stack:** Go 1.25, Wails v3.0.0-alpha.74, Node 22, Yarn 1.22.22, Python stdlib, GitHub Actions.

**Spec:** User-approved request in this chat: maintain a personal fork, merge upstream changes, build automatically, and configure the installed app to update from that fork.

## Global Constraints

- Preserve all existing model settings, provider names, channel IDs, and remembered model selection.
- Keep credentials, local configuration, generated protocol files, and build outputs out of Git.
- Personal release source: https://github.com/2tdbbfgk2m-beep/cursor-byok/releases/latest/download/.
- Personal versions: upstream build/config.yml version followed by -personal.GITHUB_RUN_NUMBER.
- Publish only macos-arm64; preserve upstream default multi-platform manifest behavior.
- Do not push to upstream. Conflict or test failure must stop synchronization before pushing or publishing.
- Use ad-hoc signatures, matching the existing local app; do not weaken OS security settings.

## Task 1: Personal update source and version ordering

**Files:** internal/buildinfo/buildinfo.go; internal/updater/semver.go; internal/updater/semver_test.go.

- [x] Add regression cases: personal.10 > personal.9; personal.11 > personal.10; equal versions do not update; newer upstream base beats older personal base.
- [x] Run `go test ./internal/updater -count=1` and confirm the numeric prerelease case fails.
- [x] Compare dot-separated numeric prerelease identifiers numerically; retain stable-over-prerelease behavior and ignore build metadata.
- [x] Point ReleaseRepo, UpdateBaseURL, and ReleasePageURL to the personal fork.
- [x] Run updater and provider regression tests.

## Task 2: Single-platform manifest and automatic release

**Files:** scripts/release/main.go; scripts/release/main_test.go; .github/workflows/personal-release.yml.

- [x] Add an integration test invoking `go run . manifest -platform macos-arm64` with only the arm64 asset; assert the literal download URL, byte length, and SHA-256 checksum for the abc fixture.
- [x] Confirm it fails on the missing platform flag.
- [x] Add optional `-platform` filtering; blank keeps all four original platforms and an unsupported value fails.
- [x] Create a main-push/manual macOS workflow: checkout exact SHA; install pinned tooling; generate protocols and Wails bindings; install frozen Yarn dependencies; build frontend; run Go/Node regressions; build arm64 executable; copy bundle assets; set the release version in Info.plist; ad-hoc sign and verify; archive; generate manifest; create draft release with both assets, then publish latest only after upload succeeds.
- [x] Validate workflow syntax with actionlint and run the manifest integration test locally.

## Task 3: Upstream synchronization and end-to-end deployment

**Files:** .github/workflows/sync-upstream.yml; docs/PERSONAL-RELEASE.md.

- [x] Add daily/manual sync: full checkout main; fetch upstream main; stop with a no-op if already merged; abort on merge conflicts; run protocol generation, provider/routing/updater regressions and Node tests; push only after checks; explicitly dispatch personal-release.yml using the repository token.
- [x] Record maintenance, conflict handling, release URLs, ad-hoc signing, and rollback instructions in docs/PERSONAL-RELEASE.md.
- [x] Preserve upstream workflow disabled on the personal fork; enable only personal release and synchronization workflows.
- [x] Push personal main, wait for its Actions run to finish, download the published asset and verify manifest SHA-256 and signature.
- [x] Build/install a bootstrap personal.0 version pointing to the personal source, use the native in-app updater to install the published version, and verify VC/Fable labels and unchanged settings.
- [x] Manually run synchronization once and confirm it is a safe no-op while upstream is unchanged.

## Completion evidence (2026-11-12)

- Cloud release [v1.0.1-personal.6](https://github.com/2tdbbfgk2m-beep/cursor-byok/releases/tag/v1.0.1-personal.6) built from `781cb41e01415e451e18ee8ca58182235170c893`; [Actions run 43082856566](https://github.com/2tdbbfgk2m-beep/cursor-byok/actions/runs/43082856566) succeeded.
- Downloaded the public manifest and archive; validated size, SHA-256, arm64 architecture and ad-hoc signature.
- Native updater exercised personal.0 → personal.1, then personal.1 → cloud-built personal.6. Installed executable matches the cloud archive; macOS signature verification passes.
- Both provider groups (VC/Fable), all seven model adapters, and remembered model selection match the pre-migration configuration. Service and model UI were inspected after restart.
- Manual Sync Upstream run 43077769304 succeeded with no upstream changes.
- An initial cloud run exposed root schema version versus nested application version injection. Replaced inline metadata editing with a tested helper; regression checks preserve root schema version 3 and verify the application version via the release CLI.

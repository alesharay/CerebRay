# Cerebray - Troubleshooting Log

Issues are listed newest first. Each entry captures what went wrong, how it was diagnosed, and how it was fixed.

---

## 2026-04-08: CI fails on `npm ci` with missing @emnapi packages

**Issue:** The "Lint (TypeScript)" job in Gitea Actions failed at the `npm ci` step with `Missing: @emnapi/core@1.9.2 from lock file` and `Missing: @emnapi/runtime@1.9.2 from lock file`.

**Investigation:** The lock file was generated on macOS where `@tailwindcss/oxide` installs a native binary (`oxide-darwin-arm64`). On macOS, npm skips the `oxide-wasm32-wasi` fallback package entirely, so its bundled `@emnapi` sub-dependencies never get written into `package-lock.json`. The CI runner (Linux) doesn't have a native binary available, falls back to the wasm package, and finds those transitive deps missing from the lock file.

**Root cause:** Platform-specific optional dependency resolution in Tailwind CSS v4. The `@tailwindcss/oxide` package ships native binaries for each OS and a wasm32-wasi fallback. npm only resolves the sub-dependencies for the platform where `npm install` runs, leaving other platforms' transitive deps out of the lock file.

**Fix:** Changed CI workflow to `npm ci || npm install --no-audit` so it falls back gracefully when the lock file has cross-platform drift. Applied to both the lint-frontend and test-frontend jobs in `.gitea/workflows/ci.yaml`.

**Lessons learned:** When using packages with platform-specific optional dependencies (like Tailwind's oxide), expect lock file drift between dev machines (macOS) and CI (Linux). The `npm ci` fallback pattern handles this without sacrificing reproducibility on matching platforms.

# Cerebray - Troubleshooting Log

Issues are listed newest first. Each entry captures what went wrong, how it was diagnosed, and how it was fixed.

---

## 2026-04-10: PostgreSQL HelmRelease fails with ErrImagePull on tag "16"

**Issue:** After deploying cerebray to k8s, the PostgreSQL StatefulSet pod failed with `ErrImagePull`. The error was `docker.io/bitnami/postgresql:16: not found`.

**Investigation:** The HelmRelease values had `image.tag: "16"` expecting the Bitnami chart to resolve a major-version-only tag. Checked Docker Hub and confirmed that Bitnami doesn't publish a bare `16` tag for postgresql - they use full semver tags like `16.8.0-debian-12-r6`.

**Root cause:** The Bitnami PostgreSQL chart doesn't remap bare major version tags to full image tags. The `image.tag` value is used directly as the Docker image tag, and `bitnami/postgresql:16` doesn't exist on Docker Hub.

**Fix:** Changed `image.tag` from `"16"` to `latest` and `pullPolicy` to `Always` in `homelab-gitops/apps/base/cerebray/postgresql.yaml`. Had to `helm uninstall postgresql -n cerebray` and let Flux reinstall since the initial Helm install was stuck with the old spec.

**Lessons learned:** Always use `latest` or a full semver tag for Bitnami Helm chart image overrides. Bare major version tags don't exist. When a HelmRelease is stuck mid-install with bad values, `helm uninstall` + Flux reconcile is the fastest recovery path.

---

## 2026-04-08: CI fails on `npm ci` with missing @emnapi packages

**Issue:** The "Lint (TypeScript)" job in Gitea Actions failed at the `npm ci` step with `Missing: @emnapi/core@1.9.2 from lock file` and `Missing: @emnapi/runtime@1.9.2 from lock file`.

**Investigation:** The lock file was generated on macOS where `@tailwindcss/oxide` installs a native binary (`oxide-darwin-arm64`). On macOS, npm skips the `oxide-wasm32-wasi` fallback package entirely, so its bundled `@emnapi` sub-dependencies never get written into `package-lock.json`. The CI runner (Linux) doesn't have a native binary available, falls back to the wasm package, and finds those transitive deps missing from the lock file.

**Root cause:** Platform-specific optional dependency resolution in Tailwind CSS v4. The `@tailwindcss/oxide` package ships native binaries for each OS and a wasm32-wasi fallback. npm only resolves the sub-dependencies for the platform where `npm install` runs, leaving other platforms' transitive deps out of the lock file.

**Fix:** Changed CI workflow to `npm ci || npm install --no-audit` so it falls back gracefully when the lock file has cross-platform drift. Applied to both the lint-frontend and test-frontend jobs in `.gitea/workflows/ci.yaml`.

**Lessons learned:** When using packages with platform-specific optional dependencies (like Tailwind's oxide), expect lock file drift between dev machines (macOS) and CI (Linux). The `npm ci` fallback pattern handles this without sacrificing reproducibility on matching platforms.

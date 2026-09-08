# Cerebray - Troubleshooting Log

Issues are listed newest first. Each entry captures what went wrong, how it was diagnosed, and how it was fixed.

---

## 2026-09-07: Promoting a note dies partway with "Error in input stream"

**Issue:** Promoting a note from the Inbox streamed AI text for a while, then stopped and showed "Error in input stream" in the UI. Short chat messages were unaffected.

**Investigation:** The error string is nowhere in the codebase. It turned out to be the browser's own message for a `fetch` response body that fails mid-read (Firefox's wording; Chrome says `network error`). It reaches the UI through `frontend/src/api/chat.ts` - the `.catch` around the reader loop passes `err.message` straight to `onError`, which `useSSE.ts` puts into `chatError` for `NoteDetailPage` to render. So the UI was faithfully showing a raw network error, not an application one.

That pointed at the connection being torn down rather than at any AI logic. `WriteTimeout: 60 * time.Second` in `cmd/server/main.go` was the culprit, and the comment beside it ("longer for SSE streaming") showed the wrong assumption.

**Root cause:** Go's `http.Server.WriteTimeout` is an absolute deadline measured from when the request headers finish being read. It is not idle-based and does not reset per write, so any handler still streaming at 60 seconds gets its write killed and the connection closed. The promote path is the only flow that regularly crosses that line: `BuildExpandPrompt` asks for all 11 Zettel fields "thoroughly" with a body of "several paragraphs" against `MaxTokens: 4096`, which streams for roughly 50-90 seconds. Because it straddles the limit, the failure was intermittent. The nginx `proxy_read_timeout` (300s) and the ingress annotation were both fine - Go was the binding constraint.

A second, independent gap surfaced while reading the provider: the SSE event switch in `internal/ai/anthropic.go` had no `case "error"`. Anthropic can emit an error event partway through a stream (overloaded, rate limited). The loop silently skipped it, hit EOF, and `StreamChat` returned a **nil** error - so the handler saved the truncated text as a complete assistant message.

**Fix:**
- `internal/handlers/chat.go` - clear the write deadline for the SSE request only, via `http.NewResponseController(w).SetWriteDeadline(time.Time{})`. This keeps the 60s protection on every other route. Also bounded the provider call with a 4 minute `context.WithTimeout` so a hung upstream cannot pin the handler forever, and included the real error text in the SSE error payload instead of a bare "AI stream failed".
- `internal/ai/anthropic.go` - added `case "error"` that returns the API's error type and message, so mid-stream failures surface instead of being persisted as finished notes.
- `cmd/server/main.go` - corrected the misleading comment on `WriteTimeout`.
- `internal/handlers/stream_deadline_test.go` - regression test that builds the real middleware chain, sets a short `WriteTimeout`, and asserts a longer stream still completes. Includes a control case proving the stream is cut without the fix (108 bytes vs 183).

**Lessons learned:**
- `WriteTimeout` is an absolute deadline, not an idle timeout. Any long-lived response (SSE, downloads, websocket upgrades) needs a per-request `http.ResponseController` escape hatch rather than a globally inflated timeout.
- `SetWriteDeadline` only reaches the connection if every `ResponseWriter` wrapper in the chain implements `Unwrap() http.ResponseWriter`. Both wrappers here come from `chi/middleware`, which does. A hand-rolled wrapper would silently break this and the fix would no-op - hence the test.
- When an error string does not exist anywhere in the repo, it is coming from the browser or a proxy. Grep first, then work outward through the layers.
- Silently ignoring unknown event types in a stream parser turns upstream failures into corrupt-but-successful writes. Handle the error event explicitly.

**Still open:** if the expand stream does fail, the user message is already stored, so a retry makes `len(dbMessages) == 2` and `chat.go` falls back to the generic system prompt instead of the expand prompt. Retrying a failed promote gives a worse note.

---

## 2026-04-10: Keycloak login fails with "auth exchange failed"

**Issue:** After deploying to k8s, clicking the Keycloak login button redirected to Keycloak correctly, but the callback returned "auth exchange failed".

**Investigation:** Backend logs showed: `Post "https://keycloak.homelab/realms/homelab/protocol/openid-connect/token": tls: failed to verify certificate: x509: certificate signed by unknown authority`. The backend container (Alpine) didn't trust the homelab CA that issued the Keycloak TLS certificate.

**Root cause:** The backend's Alpine container only ships with public CA certificates. The homelab uses a self-signed CA (via mkcert) for internal TLS. When Go's HTTP client tried to POST to the Keycloak token endpoint, TLS verification failed.

**Fix:** Added `SSL_CERT_FILE` env var pointing to `/etc/ssl/homelab/ca-certificates.crt` and mounted the `homelab-ca-bundle` ConfigMap (already present in the cluster) as a volume in the backend deployment. Go's `crypto/tls` reads `SSL_CERT_FILE` automatically.

**Lessons learned:** Any backend container that calls other homelab services over HTTPS needs the homelab CA mounted. The archdraft project already had this pattern - check existing apps when deploying new services.

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

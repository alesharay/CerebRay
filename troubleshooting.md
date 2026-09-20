# Cerebray - Troubleshooting Log

Issues are listed newest first. Each entry captures what went wrong, how it was diagnosed, and how it was fixed.

---

## 2026-09-20: Every ExternalSecret in the cluster silently stopped syncing

**Issue:** Found during a drift audit, not because anything broke. `cerebray-secrets` reported `SecretSyncedError`, and both `ClusterSecretStore/homelab-path-vault` and `ClusterSecretStore/secret-path-vault` were `InvalidProviderConfig` with the unhelpful message `unable to create client`. Every app kept running normally, which is why nobody noticed.

**Investigation:** The store's message says nothing useful, so the first question was whether Vault was reachable at all. It was:

```
ping 100.114.88.8          -> 0% packet loss
curl /v1/sys/health        -> http_code=503
```

503 is not a failure here. Vault returns it specifically to mean sealed, and `/v1/sys/seal-status` confirmed it:

```json
{"type":"shamir","initialized":true,"sealed":true,"t":1,"n":1}
```

So the NAS had restarted at some point and Vault came back sealed, as it always does. The initial read of "Vault unreachable" was wrong - it was reachable the whole time and answering correctly.

**Root cause:** Vault was sealed. ExternalSecrets could not authenticate, so no secret could be refreshed anywhere in the cluster. Nothing broke because every ExternalSecret uses `deletionPolicy: Retain`, so the last successfully synced Secret stayed in place and the apps kept reading it.

**Fix:** Unsealed Vault with the single Shamir key (`t:1`). The stores cache their status, so they needed a nudge rather than waiting out the refresh interval:

```
kubectl annotate clustersecretstore <name> force-sync="$(date +%s)" --overwrite
```

Same annotation on each ExternalSecret. All seven went `SecretSynced/True` across cerebray, archdraft, do-a-doc, keycloak and cert-manager.

**Lessons learned:**
- A sealed Vault is invisible from the app side. `Retain` keeps everything running on stale credentials, so the failure only shows up when you go looking or when a secret needs to rotate. This had been broken for an unknown length of time.
- HTTP 503 from Vault means sealed, not down. Check `/v1/sys/seal-status` before concluding there is a network problem - `InvalidProviderConfig` on the ESO side gives no hint either way.
- ESO caches store validation. After fixing the backend, force a reconcile with an annotation instead of waiting for the refresh interval (1h on most of these).
- Worth adding an alert on `ClusterSecretStore` readiness, since the blast radius is every secret in the cluster and the symptom is silence.

---

## 2026-09-20: Backend stuck in CrashLoopBackOff after the k3d cluster restarted

**Issue:** The backend pod would not stay up. Logs showed it dying on the first line of work every time:

```
{"level":"info","message":"starting cerebray"}
{"level":"fatal","error":"failed to connect to `user=cerebray database=cerebray`: 10.43.233.8:5432 ... connect: connection refused","message":"pinging postgres"}
```

The frontend and Redis were both fine, which made it look like a backend problem. It was not.

**Investigation:** `kubectl get pods -n cerebray` showed `postgresql-0` sitting in `Pending`, so the backend was just the last domino. The scheduler said why:

```
0/3 nodes are available: 1 node(s) had untolerated taint {node.kubernetes.io/unreachable: },
2 node(s) didn't match PersistentVolume's node affinity.
```

`k3d-homelab-agent-2` was `NotReady` ("Kubelet stopped posting node status"), and the Postgres PV is a `local-path` volume with `nodeAffinity` pinning it to exactly that node. Redis survived only because its PV happened to land on agent-0.

The node container itself was `Up`, so this was not a dead container. Inside it, the k3s agent was stuck in a one-second loop: `Waiting for containerd startup: rpc error: code = Unimplemented desc = unknown service runtime.v1.RuntimeService`. Containerd's own log at `/var/lib/rancher/k3s/agent/containerd/containerd.log` had the real error:

```
failed to load plugin ... failed to create CRI service:
failed to create cni conf monitor for default: failed to create fsnotify watcher: too many open files
```

`fs.inotify.max_user_instances` in the Docker Desktop VM is 512, and that limit is per-uid across the *whole* VM, shared by all three k3d nodes and every pod on them. Per-container `ulimit -n` was identical on the healthy agent-0, which is what ruled out a node-specific misconfiguration.

One red herring worth recording: the Node object was advertising `InternalIP 172.18.0.2`, which is agent-0's address, while agent-2's actual Docker IP was `172.18.0.3`. That looked like an IP conflict but was just the last-known-good status cached from before the restart. It self-corrected the moment the kubelet posted a real update.

**Root cause:** A Docker Desktop restart brought all three k3d node containers back at once. On agent-2, containerd's CRI plugin failed to load because the shared inotify instance limit was exhausted, so the CRI runtime service never registered. The kubelet cannot start until containerd answers, so it never posted node status, the node went `NotReady` and picked up the `unreachable` taint, and the node-pinned `local-path` PV made `postgresql-0` unschedulable anywhere else. No Postgres endpoint meant the backend's startup ping failed and the process called `log.Fatal`.

**Fix:** Backed up the Postgres data directory first, since no backup of it existed anywhere:

```
docker cp k3d-homelab-agent-2:/var/lib/rancher/k3s/storage/pvc-b8e82f49-..._cerebray_data-postgresql-0/data \
  "/Volumes/WD 2TB/Dev/homelab/backups/cerebray-pg-20260920"
```

Verified it matched the source exactly (1346 files, 64M, `PG_VERSION` 18), then patched the PV off its auto-delete policy as a safety net:

```
kubectl patch pv pvc-b8e82f49-253b-4585-b1f6-225c728479bd \
  -p '{"spec":{"persistentVolumeReclaimPolicy":"Retain"}}'
```

`docker restart k3d-homelab-agent-2` failed with "tried to kill container, but did not receive an exit event" and left it `Exited (143)`. A plain `docker start` brought it back cleanly. Containerd loaded, the kubelet registered with the correct IP, the taint cleared, and `postgresql-0` scheduled and replayed WAL without incident (`redo done`, `database system is ready to accept connections`). Deleting the backend pod cleared its five-minute backoff and it came up connected. Data verified intact: 1 user, 14 notes, 28 conversations, 71 messages.

**Lessons learned:**
- Start at the scheduler, not the application logs. The backend's fatal error was accurate and completely useless - `kubectl get events` and the `Pending` pod named the problem immediately, and there is still no Taskfile task that surfaces either.
- `local-path` pins a PV to one node through `nodeAffinity`, and `WaitForFirstConsumer` means that pin is decided silently by wherever the pod first landed. `postgresql.yaml` never sets a `storageClass`, so it inherits the `local-path` default and the pin exists nowhere in the manifests. Adding a `nodeSelector` would not help - the PVC has already bound.
- The `nfs` StorageClass (NAS-backed, `reclaimPolicy: Retain`) is already installed and Keycloak uses it via `global.storageClass: nfs`. Cerebray's Postgres and Redis, and do-a-doc's MongoDB and Redis, all silently inherit `local-path` and carry this identical failure mode today.
- A `NotReady` node whose container is still `Up` is usually a container-runtime problem, not a dead node. Check for the `Waiting for containerd startup` loop and read containerd's own log before concluding anything is corrupt.
- `docker logs` on these node containers returns a stale buffer, so `--since` can come back empty while the container is actively failing. Read the unfiltered tail and the in-container log instead.
- A stale `InternalIP` on a `NotReady` node is a symptom of the node being stalled, not evidence of an address conflict.

**Resolved later the same day**, during the follow-up drift audit:
- All six PVs in the cluster are now `reclaimPolicy: Retain`, not just this one. The four that were still `Delete` were cerebray's own Redis, archdraft's Postgres, and do-a-doc's MongoDB and Redis.
- `task k8s:manifests` is safe again. All five template regressions are fixed (the three named above plus `KEYCLOAK_CLIENT_ID` and a 256Mi memory limit that would OOMKill SSE streams), the Postgres tag now matches live rather than reverting to `"16"`, and the task preserves the deployed image tag instead of resetting it to `:latest`. A new `k8s:manifests:guard` refuses to run if the gitops checkout is behind `origin/main` or has uncommitted changes under `apps/base/cerebray` - the clone was three months stale when this was found.
- `schema_migrations` did not exist in the live database at all, so `task db:migrate:up` would have replayed `000001` and failed on `relation "users" already exists`, leaving the tracker dirty. Baselined with `migrate force 8`; `up` is now a clean no-op.

**Still open:**
- **No backup exists for any cluster database.** There are no backup CronJobs and no `db:backup` task. The copy taken above is a one-off.
- **The Postgres image is still unpinned.** The HelmRelease runs `tag: latest` with `pullPolicy: Always` under a `>=16.0.0 <17.0.0` chart constraint, and the on-disk data directory is `PG_VERSION` 18. The template now carries a warning comment, but a real pin to an explicit 18.x tag needs a maintenance window. Redis, MongoDB and Keycloak carry the same unpinned risk in other namespaces.
- **`/health` checks nothing.** `internal/handlers/health.go` returns a static 200 and both the liveness and readiness probes point at it. Harmless while the process dies before listening, but any retry-on-startup change must split out a real `/ready` that pings Postgres and Redis first, or readiness will pass with a nil pool and every request will 500.
- **The Redis startup ping has no timeout.** `cmd/server/main.go:61-68` uses a bare `context.Background()` where the Postgres ping gets 10 seconds. If Redis hangs rather than refusing, the process blocks forever in `Running`/not-ready instead of crash-looping.

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

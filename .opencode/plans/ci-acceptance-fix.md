# Fix acceptance with automatic Elasticsearch HTTPS

## Goal, boundaries and evidence

Fix PR #1 acceptance on `/projects/elasticsearch`, module `github.com/disaster37/elasticsearch/v9`, branch `feat/auto-retry-multi-address`, base v9. Keep acceptance enabled and make the full existing CI pipeline green. This finalized HTTPS design replaces the previous HTTP proposal.

The user confirmed an empty `git status --porcelain=v1 --untracked-files=all` before investigation; the orchestrator reconfirmed a clean tree after the first capture. This planning session had file tools but no shell/orchestration tool; runtime results below are supplied orchestrator evidence, not commands executed by this agent. Only this plan file was edited.

### Confirmed facts

- CI run `38581490920`, pinned Dagger 0.21.10: lint, spec-check (`implemented=158, known-missing=354, ignored-params=21`) and unit tests passed. Acceptance failed in the eager curl sidecar, panicking at `ci/dagger/main.go:282` from `TestAcceptance:202`. Logs included curl exit 35 / TLS `wrong version number`; connection failures were also reported.
- First Docker reproduction, exact original ES 9.0.0 fixture settings including explicit `xpack.security.enabled=true`: plain HTTP on 9200 returned 401 unauthenticated and green authenticated health; HTTPS returned rc=35 / `wrong version number`, through both localhost and another container's network DNS. Effective HTTP and transport SSL settings were false. Logs showed security enabled, no certificates, and a transport-SSL warning. First HTTP response: 13,986 ms.
- Second reproduction, two independent fresh Docker runs with only explicit `xpack.security.enabled=true` removed: security auto-configuration wrote `xpack.security.enabled: true`, HTTP SSL enabled and transport SSL enabled to `config/elasticsearch.yml`. `config/certs/` contained `http.p12`, `http_ca.crt` and `transport.p12`. HTTPS returned 401 unauthenticated and green health with the existing fixed password. Plain HTTP returned rc=52 / empty reply. Auto-configuration was enabled, certificates were generated afresh, and no WARN/ERROR lines were reported. First HTTPS responses were reported at 19–25 seconds (individual measurements 19 and 23 seconds).
- Source inspection: all existing CI consumers expect `https://elasticsearch.svc:9200`. `api/integration_helper_test.go:27-33` already sets `TLSSkipVerify: true` and `AllowInsecureHTTP: true`. No integration-helper/client change is needed for self-signed HTTPS. Only `api/document_service_integration_test.go` currently carries integration tests, including `TestIntegrationDocumentCRUD`.
- Elastic documentation states that explicitly setting `xpack.security.enabled`, even to true, skips security auto-configuration. `discovery.type=single-node` is an allowed exception to discovery settings that would otherwise skip it. Pinned Docker entrypoint initializes `bootstrap.password` from `ELASTIC_PASSWORD` before starting ES.

**Root cause:** explicit security enablement suppresses automatic certificate/TLS setup, leaving authenticated plaintext HTTP while CI requests HTTPS. `-k` disables verification, not the TLS protocol, and cannot fix this. Separately, a fixed 10-second sleep and single curl provide inadequate readiness handling, especially with 19–25-second automatic TLS startup. The eager sidecar fails before acceptance can report useful setup/test errors and creates a separate service-consuming evaluation.

**Limits of the evidence:** the supplied captures establish Docker image behavior, not an explicit Dagger 0.21.10 reproduction of automatic setup. Dagger normally checks exposed service TCP ports before starting a bound consumer; the exact connection-refused path in CI, actual double service launches, and startup duration on constrained runners are not proven. Validate those boundaries with the commands below rather than claiming the Docker timing proves the Dagger race.

References inspected:
- https://www.elastic.co/docs/deploy-manage/security/self-auto-setup
- https://raw.githubusercontent.com/elastic/elasticsearch/v9.0.0/distribution/docker/src/docker/config/elasticsearch.yml
- https://raw.githubusercontent.com/elastic/elasticsearch/v9.0.0/distribution/docker/src/docker/bin/docker-entrypoint.sh
- Dagger service binding and exposed-port documentation via Context7 (`/dagger/dagger`).

## Approved design

1. **HTTPS via automatic security configuration**, not plaintext HTTP and not manual certificate generation. Remove the explicit security-enabled environment setting; keep the existing image entrypoint/output, fixed password, single-node discovery and ephemeral filesystem.
2. Remove the eager alpine/curl sidecar from `Elasticsearch()`. Define and return the service lazily. Each acceptance/debug consumer performs the same bounded authenticated health wait in its own bound container. Keep Dagger's exposed-port TCP health check as an additional startup check.
3. Keep existing skip-verification for isolated CI self-signed certificates: curl `-k`, integration helper `TLSSkipVerify: true`. This exercises TLS and basic authentication, not CA/hostname verification. Verified-CA tests and production deployment hardening are out of scope.
4. Preserve live ES stdout/stderr for Dagger diagnostics. **Do not install the previous proposal's stdout-redirection wrapper or set file logging:** output redirection can suppress auto-configuration. No new dependencies, public client changes, spec policy changes, workflow changes or generated SDK edits.

## Ordered implementation tasks

### 1. Establish baseline and change the fixture — `ci/dagger/main.go`

- Recheck branch/tree and record the current filtered acceptance failure before editing.
- In `Elasticsearch(ctx)` remove `.WithEnvVariable("xpack.security.enabled", "true")` only from security configuration. Do not replace it with false or add explicit HTTP/transport SSL settings. Do not set `xpack.security.autoconfiguration.enabled=false`, disable enrollment or overwrite `elasticsearch.yml`.
- Preserve `ELASTIC_PASSWORD`, `discovery.type=single-node`, cluster/node names, publish host, memory-lock setting, trial license, `path.repo`, backup mkdir/chown, user switching, 1 GiB heap, exposed port 9200, and the existing service cache-busting variable. Preserve the original image entrypoint/default command; no custom ES startup shell, output redirection, keystore manipulation or certificate commands.
- Delete the entire eager health-wait sidecar and its `panic(err)`. `Elasticsearch()` should return the definition without evaluating a consumer.
- Correct the service comment: HTTPS/security are provided by first-start automatic configuration; explicit security/SSL overrides can disable it. Certificates/configuration/data are ephemeral per service container; do not add persistent ES data/config mounts or rely on previous runs' certificates.

### 2. Share bounded readiness — `ci/dagger/main.go`

Add a private endpoint constant `elasticsearchURL = "https://elasticsearch.svc:9200"` and private readiness-script helper. Use the helper in both `TestAcceptance` and `DebugTest`. Reuse `helper.ForgeScript`, curl and shell already present rather than adding a retry library.

Both consumer containers must bind the exact service with `WithServiceBinding("elasticsearch.svc", esService)` and set:

```text
ELASTICSEARCH_URL=https://elasticsearch.svc:9200
ELASTICSEARCH_USERNAME=elastic
ELASTICSEARCH_PASSWORD=<existing fixed fixture password, via WithSecretVariable>
```

Acceptance already sets these; debug must set them too. Never embed the literal password in generated scripts or turn on `set -x`. Leave `api/integration_helper_test.go` unchanged; its existing TLS settings support generated self-signed certificates. No `AllowInsecureHTTP` change is needed.

**Readiness algorithm (implement these semantics exactly):**

1. Under `set -eu`, establish a 180-second deadline using `date +%s`; attempt immediately, then sleep 2 seconds between retries, clamping sleep to remaining budget. Remove every fixed startup sleep. No nested curl retries that multiply the deadline.
2. Each attempt uses:

   ```sh
   curl --silent --show-error --insecure --connect-timeout 2 \
     --max-time "$attempt_timeout" \
     --user "$ELASTICSEARCH_USERNAME:$ELASTICSEARCH_PASSWORD" \
     --output "$body_file" --write-out '%{http_code}' \
     "$ELASTICSEARCH_URL/_cluster/health?wait_for_status=yellow&timeout=2s"
   ```

   Set `attempt_timeout=min(5, remaining seconds)`; never issue a request once the deadline expires. Capture curl rc, HTTP status and stderr with temporary files, explicitly handling failure inside an `if` so `set -e` does not exit before retry/diagnostics. Clear response files before each attempt and clean them up on exit.
3. Success requires curl rc=0, HTTP 200, ES JSON `status` exactly yellow/green, and `timed_out=false`. HTTP 200 alone is not readiness. Match these exact JSON fields with whitespace-tolerant patterns against the simple ES response; do not search for an arbitrary word green. Unexpected/malformed successful responses fail with their body rather than count as ready.
4. Retry transient connection refusal/reset, DNS failures, timeouts, empty reply (rc=52), TLS connection failures (rc=35), HTTP 408/429/5xx, and red/timed-out HTTP 200 health until the deadline. This explicit loop provides the required all-errors retry coverage, including failures curl's default retries exclude. Repeated TLS/empty-reply errors print a hint that the listener may be serving HTTP because auto-configuration was skipped; **never fall back to HTTP**.
5. Fail immediately on HTTP 401/403 with an authentication-specific message, other unexpected 4xx/redirects, malformed health, and nontransient curl invocation/configuration errors. Do not follow redirects or hide wrong credentials behind 180 seconds of retries.
6. Print attempt/elapsed time, curl rc and HTTP status without authentication headers. On exhaustion print `Elasticsearch HTTPS readiness failed after 180s`, expected URL, last curl error/status and last response body, bounded to 8 KiB. Exit nonzero; no silent skip or final `|| true`.

Readiness deliberately omits `--fail`: it handles HTTP statuses explicitly and must not conflate authenticated HTTP errors with transport failures. Avoid the old `timeout=500s` single request. When using Go formatting for scripts, escape curl `%{http_code}` as needed or compose the readiness fragment outside `fmt.Sprintf`.

### 3. Acceptance/debug error boundaries — `ci/dagger/main.go`

- Acceptance script order: `set -eu`; readiness; existing index-template PUT; gotestsum installation; existing full integration-tagged tests and atomic coverage. Preserve template payload/priority and optional test regex. Add `-timeout=10m` to bound Go test hangs.
- Template PUT stays HTTPS with `--insecure`, environment-based credentials, connect timeout 2s and maximum request time 30s. Capture response/status explicitly and require HTTP 200 with `acknowledged=true`; failure prints bounded response/error and stops before tests. Do not blindly retry permanent template rejection. `--fail` is unnecessary when status is validated explicitly; never assume a 200 proves acknowledgement.
- Change `TestAcceptance` to return `(*dagger.File, error)`. Evaluate the consumer with `Sync(ctx)` before returning its coverage file, catching readiness, template, installation and test failures at this boundary. Wrap the original Dagger error with stage and expected HTTPS endpoint, preserving original exec stdout/stderr (including the last curl diagnostic). Do not panic or return a successful coverage artifact after failure.
- Update `Ci()` to handle the returned acceptance error before adding coverage. Preserve CLI function names and successful artifact shape. No product-client/API signature changes.
- Debug must use the shared readiness helper and all three env variables before delve. Add `-tags=integration` to `dlv test`, otherwise its current command does not select the integration-tagged tests. Remove `defer esService.Stop(ctx)` from the lazily returned debug-service function; it runs before consumption and binding manages service lifetime. Preserve debug port 4000 and interactive behavior; do not impose acceptance's test timeout on an attached debugger.

**Diagnostics:** preserve the original ES service stdout/stderr in Dagger's execution trace. Validation must run with plain progress so startup/configuration/certificate errors remain visible alongside consumer diagnostics. If Dagger TCP startup checking fails before the consumer runs, propagate that original service-start error with fixture context; the readiness script cannot produce a curl error in that case. Do not pretend `container.Stdout(ctx)` reads a running service's logs: it can evaluate another container/startup path. No unverified service-log SDK calls, restarts for diagnostics, Docker-socket mounting, or redirecting ES output. The guaranteed acceptance diagnostic is the original Dagger error plus bounded last curl stderr/body when its script runs; ES logs are retained in the Dagger service trace, not promised as a separately fetched dump. Check this in failure-injection validation. If the engine trace omits startup errors, treat that as a validation failure requiring compatible log retrieval—not permission to mask the failure or switch to HTTP.

### 4. Regression/failure tests

Add CI helper tests in `ci/dagger/main_test.go` if a suitable module test harness is available; execute in the Dagger-generated module environment, not root-module tests alone. Do not commit generated SDK files. Use disposable container script tests with stub curl/time/sleep to validate:

- connection refused, TLS rc=35 and empty reply rc=52 followed by success;
- 503/429 then healthy; red/timed-out health then green;
- permanent refusal/wrong scheme reaches deadline without fallback;
- 401 fails immediately; malformed/incorrect 200 does not pass;
- template 400 stops before gotestsum; bounded output and original exit propagation;
- both consumers use HTTPS, env credentials, shared wait; no eager sidecar or bare panic remains.

Then perform disposable Dagger failure injections: invalid ES setting / certificate-config failure must expose service startup errors; wrong password must expose authentication error; temporarily wrong HTTP scheme must fail, not pass or skip; slow startup must retry within budget. Remove all injected faults before normal checks. Preserve original entrypoint/output in every auto-configuration test.

## Exact verification commands — no publishing/pushing locally

```sh
cd /projects/elasticsearch
git status --porcelain=v1 --untracked-files=all
git branch --show-current
dagger version   # require CLI v0.21.10; dagger.json also pins engine v0.21.10

# Before implementation: capture the existing failure.
dagger --progress=plain call --src . test-acceptance --run '^TestIntegrationDocumentCRUD$'

# After implementation: filtered smoke, full acceptance, fresh-repeat validation.
dagger --progress=plain call --src . test-acceptance --run '^TestIntegrationDocumentCRUD$'
dagger --progress=plain call --src . test-acceptance
dagger --progress=plain call --src . test-acceptance
dagger call --src . lint
dagger call --src . spec-check
dagger call --src . test
dagger call --src . vet
git diff --check
git diff --stat
```

The plan is now an intentional new/changed file relative to the original clean-tree confirmation. Do not commit diagnostic output, coverage changes or generated SDK artifacts unintentionally.

For actual Dagger auto-configuration evidence, inspect the live service startup trace during the repeated acceptance runs and verify authenticated HTTPS health/CRUD succeeds. In a disposable diagnostic consumer bound to the same fixture, also probe HTTPS root without auth (expect 401), authenticated health (expect green/yellow), and plain HTTP (expect empty reply/failure); request `/_nodes/settings?include_defaults=true&flat_settings=true` using fixed fixture credentials to confirm HTTP/transport TLS enabled. Do not add a permanent diagnostic endpoint/function solely for this experiment. Fresh repeated runs must use the fixture's existing cache-busting variable and no persisted ES data/config.

**Do not run `dagger call --src . ci` locally.** The source gates uploads/pushes on `ci=true`, but local verification intentionally avoids that orchestration entrypoint and uses only safe individual functions above.

The existing `.github/workflows/workflow.yaml` will run Dagger 0.21.10 on the PR with:

```sh
dagger call --src . ci --ci --code-cove-token env:CODECOV_TOKEN \
  --git-token env:GITHUB_TOKEN --git-branch "${BRANCH_NAME}" ${IS_TAG}
```

Confirm the new run completes build/vet/lint/format/spec-check/unit/**acceptance**, produces acceptance coverage, then succeeds at CodeCov and git push-back before declaring the full pipeline green. Preserve all workflow triggers and spec-check policy; no acceptance suppression is authorized.

## Edge cases and residual risks

- **Slow constrained CI:** automatic TLS already took 19–25s; tolerate 60s+ startup with a 180s application deadline. Dagger TCP startup checking, pulls and gotestsum download occur outside that deadline. Allow sufficient job time for those plus up to 10m tests; inspect actual stage timing before increasing budgets. Never reinstate a fixed sleep.
- **Auto-configuration/cert generation failure:** output/permissions, explicit settings, nonempty data or nonwritable config may suppress automatic setup. Keep the original entrypoint/output and writable image config; retain no data/config caches. Persistent rc=35/52 is an actionable auto-configuration/scheme diagnostic, not a reason to fall back to HTTP. Self-signed verification is intentionally skipped only in this fixture.
- **Crash/restart:** binding lifecycle may release/restart services; removing the eager evaluation avoids the unnecessary separate consumer. One acceptance consumer runs readiness/template/tests in sequence. Unexpected restarts can regenerate certificates and erase indices; investigate ES/Dagger logs rather than assuming persistence or retrying CRUD indiscriminately. Do not build a restart loop.
- **Resources/bootstrap:** preserve 1 GiB heap initially; inspect OOM/runner memory, disk and kernel/bootstrap logs if startup fails. Do not disable checks or reduce resource requirements speculatively. Successful automatic transport TLS removes the earlier warning; do not add manual transport TLS overrides.
- **DNS/Docker-in-Docker:** use `elasticsearch.svc` via service binding, never localhost inside consumer containers. Retry transient DNS/socket failures within budget; persistent failures include endpoint/status/error and service trace. Docker runtime captures do not prove Dagger networking, so repeated pinned-Dagger checks are mandatory.
- **CI/local parity and security:** same image, password, trial license and HTTPS URL; generated certificates/config/state are local to each fresh fixture. Password works through image bootstrap entrypoint. No CA export or keystore passwords need to be managed by the Go client. Do not expose the fixed-password service publicly or describe skip-verification as production-safe.
- **Other downstream failures:** download/spec clone outages, client CRUD regressions, fork secret availability, CodeCov permissions and push-back races may appear after acceptance is fixed. Diagnose separately; this plan does not preemptively change those stages.

## Rollout and rollback

1. Record baseline; remove the explicit security setting/eager sidecar and preserve automatic startup behavior.
2. Add shared HTTPS readiness/environment wiring, template validation and acceptance error propagation; update debug and `Ci()` callers.
3. Validate helper failure cases, service-start diagnostics and repeated fresh Dagger HTTPS acceptance; run safe pipeline checks and review scoped diff.
4. Submit the CI-module fix on the existing PR; existing GitHub CI proves every stage green. No data migration, public API migration or feature flag is needed.
5. Roll back by reverting only this fix's commit(s), preserving unrelated branch work. That restores the known failing baseline and is containment, not a workaround. Do not skip acceptance, disable authentication, silently switch to HTTP or introduce manual certificates. If Dagger prevents automatic setup despite the Docker evidence, stop and obtain a new design decision with its startup logs.

No unresolved design choice: automatic-security HTTPS is approved. Actual Dagger startup/diagnostics and full-pipeline success remain required validation obligations, not facts claimed by this planning agent.

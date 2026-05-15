# Changelog

All notable changes to this provider are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this provider adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.4.1] — Unreleased

### Changed

- **Lowered the rate-limited HTTP transport's minimum interval from 1s to 100ms.** The 1s spacing introduced in 1.3.0 was defensive against Shodan-style 1 req/sec limits, but Pingdom paid plans allow significantly higher throughput. With 50+ checks in a workspace the 1s floor added ~50 seconds per refresh; 100ms still smooths bursts while letting plan/apply complete promptly. 429 retry with exponential backoff remains in place as the actual safety net.
- `newRateLimitedTransport` minimum-interval clamp relaxed from 1s to 1ms; passing 0 now effectively disables spacing (still gated by the 1ms floor).

## [1.4.0] — 2026-05-15

### Changed

- **Replaced the archived `github.com/russellcardullo/go-pingdom` upstream library with an internal REST client (`pingdom/client.go`).** Upstream was archived in February 2023 and has not received updates since November 2020. The internal client now owns: HTTP transport (via the existing `rateLimitedTransport`), request encoding (form-encoded query params for `/checks` endpoints, JSON body for `/alerting/contacts` and `/alerting/teams`), response parsing including the `check.type` polymorphic shape (bare string for `"ping"`, single-key object for `{"http":{...}}`/`{"tcp":{...}}`), and the `*PingdomError` envelope returned for non-2xx responses. Public type names mirror the legacy library so end-user HCL is unaffected.
- **`go.mod` direct dependencies dropped to one (`terraform-plugin-sdk/v2`).** Removes the unmaintained transitive surface that came with go-pingdom (`io/ioutil` callsites, deprecated `encoding/json` patterns, plus its own transitive deps).
- **Module Go version bumped to `1.25.8`** to match the CI test runner's Go release.

### Added

- **httptest-based integration tests covering the wire protocol** (`pingdom/client_test.go`, 12 tests):
  - `Checks.Read`: 200 OK happy path with team backfill, 404 surfaced as `*PingdomError`, HTTP type-detail decode, malformed-error-body fallback.
  - `Checks.Create`: HTTP variant query-param serialisation including `integrationids`/`tags`/`verify_certificate`/`ssl_down_days_before`, empty-value stripping for Ping/TCP variants.
  - `Contacts.Create`: JSON body shape + `Content-Type: application/json` header.
  - `Teams`: full Create → Read → Update → Delete lifecycle against a mock server, asserting the exact method/path sequence.
  - `HttpCheck`: header serialisation order (`requestheaderN` sorted by key), inline auth (`auth=user:pass`), validation rejections (`TCPCheck.Port < 1`, `HttpCheck.ShouldContain` + `ShouldNotContain` mutual-exclusion).
- **Test coverage now spans 33 test functions across 5 files** — payload helpers, HTTP transport, and now the REST client itself. The schema-to-API contract and the API-to-Pingdom contract are both locked in golden form.

### Removed

- `github.com/russellcardullo/go-pingdom` (no longer in `go.mod` or `go.sum`).

### Upgrade notes

End-user HCL is unchanged. The provider's exposed schema (resource attributes, data sources, importer behaviour) is identical to 1.3.1. The internal types (`pingdom.HttpCheck`, `pingdom.Contact`, `pingdom.PingdomError`, etc.) keep their names but now resolve to the internal package — anyone vendoring this provider as a Go module dependency should re-run `go mod tidy` to drop the transitive go-pingdom.

## [1.3.1] — 2026-05-15

### Added

- **Unit tests for payload construction and HTTP transport.** Covers `checkForResource` (http/ping/tcp variants + integration/team/user IDs + request headers + unknown-type error path), `contactForResource` + `getNotificationMethods` (severity validation, provider allow-list, paused passthrough), `teamForResource` (members + empty case), `sortString`, and `rateLimitedTransport` (request spacing, 429 retry, max-retry give-up, request-body replay across retries, min-interval clamp). 20 tests total, none require live API access — golden behaviour now locks the schema-to-API mapping before any future internal-client refactor.

## [1.3.0] — 2026-05-15

### Changed

- **Migrated from `terraform-plugin-sdk v1` to `v2`.** The `v1` SDK has been deprecated by HashiCorp for years and is missing from current `terraform validate` / acceptance-testing tooling. All resources and data sources now use the context-aware `CreateContext` / `ReadContext` / `UpdateContext` / `DeleteContext` lifecycle, return `diag.Diagnostics` instead of plain `error`, and pass a `context.Context` through that future cancellation/timeout logic can hook into. End-user HCL is unchanged — there is no schema-level breaking change.

### Added

- **Rate-limited HTTP transport with 429 retry.** The provider now wraps the go-pingdom `HTTPClient` with `pingdom.rateLimitedTransport`: every outbound API call is spaced at least 1 second apart, and a `429 Too Many Requests` response triggers exponential backoff (`1s, 2s, 4s`, up to 3 retries) before propagating. Large workspaces hitting Pingdom's burst protection are now handled inside the provider instead of failing the plan.
- 60-second HTTP timeout on the underlying `http.Client` — prevents hung connections from deadlocking refresh.

### Removed

- `github.com/mitchellh/mapstructure` direct dependency. The provider only needed it to decode a single attribute (`api_token`); replaced with a direct `d.Get` call in `providerConfigure`.

### Notes for future work

- `github.com/russellcardullo/go-pingdom` upstream is **archived** (Feb 2023, no future maintenance). The next major release should replace it with an internal REST client to remove the dead-dep risk and to support Pingdom API features the library does not cover. Tracked as v2.0.0 work.

## [1.2.4] — 2026-05-15

### Fixed

- **`pingdom_check.paused` could trigger a perpetual plan diff.** Read populated the `paused` attribute only when the API reported the check as paused; the false branch was silently skipped, so toggling `paused = false` in config kept the state pinned at `true` and plan kept proposing the same no-op change. Read now writes the actual value unconditionally.
- **`pingdom_check` and `pingdom_team` Read each made two API calls per refresh.** The old flow listed every check/team and then re-read by ID to confirm existence — for a workspace with N resources this scaled at 2×N requests per plan and made Pingdom rate-limit far easier to trip. Read now calls the ID-scoped endpoint directly and treats a `*pingdom.PingdomError` with `StatusCode == 404` as resource-gone (`d.SetId("")`), idiomatic Terraform SDK behaviour.
- **`pingdom_contact` Read hard-errored when a contact was deleted out-of-band.** The provider returned `Error retrieving contact: ...` instead of removing the resource from state. Read now mirrors the check/team 404 pattern.
- **`PINGDOM_API_TOKEN` environment variable silently overrode explicit provider config.** A leftover token in a shell environment could shadow `api_token = var.pingdom_api_token` in TF code and obscure auth issues. The env var is now a fallback for the unset case — explicit config wins. If neither is set the provider fails configuration with `pingdom: api_token is required`.
- **`go-pingdom` client constructor error was discarded.** `pingdom.NewClientWithConfig` returned `(client, _)`; any non-nil error surfaced later as a confusing nil-pointer or auth failure. Errors are now surfaced at Configure time.
- **Duplicate `responsetime_threshold` set** in `pingdom_check` Read (set once at the top, then again inside the HTTP branch with the same source) — collapsed to a single write.
- **Stale `hostname` reference in debug log** of `pingdom_check` Create/Update (`d.Get("hostname")` always returned `<nil>` since the schema field is `host`).

### Changed

- Dropped unused `terraform-plugin-sdk/v2` direct dependency from `go.mod`. The provider still targets SDK v1 (`terraform-plugin-sdk v1.17.2`); v2 was listed but not imported anywhere.

### Upgrade notes

If your CI environment exports `PINGDOM_API_TOKEN` for the provider, behaviour is unchanged when provider config has no `api_token`. If you set `api_token` explicitly in provider config AND have `PINGDOM_API_TOKEN` in env with a different value, the new release uses the explicit config value (the old release used the env value).

## Earlier versions

See git history.

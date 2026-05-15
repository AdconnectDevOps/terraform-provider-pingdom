# Changelog

All notable changes to this provider are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this provider adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.4] — Unreleased

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

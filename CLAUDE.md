# CLAUDE.md

Notes for contributors (and AI assistants) working on this provider.

## What this is

A Terraform provider for the Pingdom API (uptime checks, contacts, teams). Built on `terraform-plugin-sdk/v2`. Wraps `github.com/russellcardullo/go-pingdom` for transport (note: upstream is archived — see "Known limitations"). Published to the Terraform Registry as `AdconnectDevOps/pingdom`.

## Layout

```
.
├── main.go                                # plugin.Serve entrypoint (SDK v2)
├── pingdom/
│   ├── provider.go                        # schema, ResourcesMap, ConfigureContextFunc
│   ├── config.go                          # API client wiring + token resolution
│   ├── transport.go                       # rate-limited http.RoundTripper with 429 retry
│   ├── resource_pingdom_check.go          # pingdom_check resource (http/ping/tcp)
│   ├── resource_pingdom_contact.go        # pingdom_contact resource
│   ├── resource_pingdom_team.go           # pingdom_team resource
│   ├── data_source_pingdom_contact.go     # data source: lookup contact by name
│   └── data_source_pingdom_team.go        # data source: lookup team by name
├── docs/                                  # user-facing docs (TF registry)
├── examples/                              # runnable HCL examples
├── .goreleaser.yml                        # build matrix
├── .github/workflows/                     # release on tag push, test on PR
└── CHANGELOG.md                           # release notes
```

## Common commands

```bash
make build      # build for local arch
go vet ./...
gofmt -s -l .   # CI fails on any non-empty output
go test ./...   # currently main_test.go is empty — no real tests
```

## Release flow

Same pattern as the other AdconnectDevOps providers: push a tag matching `v*` (e.g. `v1.2.4`) to trigger `.github/workflows/release.yml`. GoReleaser builds linux/darwin × amd64/arm64, signs the checksum file with the GPG key from `secrets.GPG_PRIVATE_KEY`, publishes a GitHub Release. The public Terraform Registry picks up new releases automatically.

Before tagging:
1. Move the `Unreleased` section in `CHANGELOG.md` under the new version header with today's date.
2. `go vet ./... && gofmt -s -l .` locally — CI is strict on these.
3. `go mod tidy` to keep `go.mod` clean.

Tags must be annotated (signed): `git tag -m "Release vX.Y.Z" vX.Y.Z` — a lightweight tag without `-m` fails when `tag.gpgsign=true` is set in git config.

## Conventions used here

### SDK v2 lifecycle (`CreateContext`, `ReadContext`, `UpdateContext`, `DeleteContext`)

Resources implement the SDK v2 context-aware callbacks via `*schema.Resource`. Each callback takes `context.Context` + `*schema.ResourceData` + `meta interface{}` and returns `diag.Diagnostics`. Wrap plain `error` returns via `diag.FromErr(...)`. Patterns:

- **`CreateContext`** — build the API payload with `<resource>ForResource(d)`, call the go-pingdom client, `d.SetId(strconv.Itoa(result.ID))`, then `return resourcePingdomXRead(ctx, d, meta)` to populate computed fields.
- **`ReadContext`** — `strconv.Atoi(d.Id())`, call the ID-scoped client method (`Checks.Read(id)`, `Teams.Read(id)`, `Contacts.Read(id)`). On `*pingdom.PingdomError` with `StatusCode == 404`, call `d.SetId("")` and `return nil` — Terraform will recreate on next plan.
- **`UpdateContext`** — call the ID-scoped Update method, then delegate to Read (when populating computed fields back to state matters; check does, team/contact don't).
- **`DeleteContext`** — call the ID-scoped Delete method. The go-pingdom client treats 404 as an error; if you want idempotent deletes, surface the 404 the same way Read does.

### Importer uses context-aware passthrough

```go
Importer: &schema.ResourceImporter{
    StateContext: schema.ImportStatePassthroughContext,
},
```

`State: schema.ImportStatePassthrough` from SDK v1 is replaced by `StateContext` in v2.

### Always reflect actual state in `Read`

Setting a boolean only on its `true` branch (`if ck.Status == "paused" { d.Set("paused", true) }`) causes the false→true drift loop: Terraform never writes `false` back to state, so a config change to `false` keeps proposing the same diff forever. **Always write the value unconditionally** (`d.Set("paused", ck.Status == "paused")`). Same rule for any computed/optional bool, int, or string.

### 404 detection via `*pingdom.PingdomError`

```go
import "github.com/russellcardullo/go-pingdom/pingdom"

res, err := client.Checks.Read(id)
if err != nil {
    if perr, ok := err.(*pingdom.PingdomError); ok && perr.StatusCode == 404 {
        d.SetId("")
        return nil
    }
    return fmt.Errorf("Error retrieving check: %s", err)
}
```

This is the only safe way to distinguish "deleted out-of-band" from "transient network error" with this client library.

### Token resolution order

`config.go` `Client()` prefers explicit provider config over the `PINGDOM_API_TOKEN` env var (env is a fallback for the unset case). The provider fails configuration with a clear error if neither is set — silent empty-token failures further down the stack are worse than an immediate fail-fast.

### Don't call `List()` per `Read`

Pingdom's list endpoints return every resource of a type for the account. Calling `Checks.List()` inside `Read` to check existence multiplies the request count by N (where N is the number of state entries) on every plan — Pingdom's rate limit trips fast. Use the ID-scoped Read + 404 detection above instead.

## Known limitations

- **`russellcardullo/go-pingdom` upstream is archived** (Feb 2023, last release v1.3.0). No future bug fixes, no support for new Pingdom API endpoints. Next major release should replace the library with an internal REST client — see the existing `terraform-provider-shodan` repo for the pattern (REST client + types + CRUD methods in a single `pingdom/client.go`). Until then, anything in Pingdom's API that go-pingdom does not expose is unreachable from this provider.
- **No acceptance tests.** `main_test.go` is an empty `package main` stub. Adding even a minimal unit test for `checkForResource` / `contactForResource` payload building would catch a meaningful class of regressions. Aim to add real tests before the v2.0.0 rewrite so behaviour is locked in.

## Adding a new resource

1. Create `pingdom/resource_pingdom_<thing>.go` modelled on the existing resources.
2. Register it in the `ResourcesMap` of `Provider()` in `provider.go`.
3. Add user docs at `docs/resources/pingdom_<thing>.md` (and to `docs/index.md` table).
4. Add an example at `examples/<thing>.tf`.
5. Update `CHANGELOG.md` under the unreleased section.

## Style

- `go fmt` everything — CI runs `gofmt -s -l .` and fails on any output.
- Comments explain *why*, not *what*. The code already shows what.
- Surface errors via `fmt.Errorf("doing X: %w", err)`; let the SDK wrap them into diagnostics.
- No `panic` in non-fatal paths.

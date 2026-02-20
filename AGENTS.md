# AGENTS.md - Terraform Provider Contentful

Terraform provider for the Contentful Content Management API, built with the
Terraform Plugin Framework (Protocol 6). The SDK client is auto-generated from
an OpenAPI spec via oapi-codegen.

## Build / Test / Lint Commands

This project uses [Task](https://taskfile.dev/) (Taskfile.yml) instead of Make.
Environment variables are loaded from `.env` via the `dotenv` directive.

```bash
# Format code
task format            # runs: go fmt ./...

# Run all unit tests
task test              # runs: go test -v ./...

# Run a single test (by name regex)
go test -v -run TestSpaceResource_Create ./internal/resources/space/

# Run a single test (acceptance, hits real API)
TF_ACC=1 go test -v -run TestSpaceResource_Create ./internal/resources/space/

# Run all acceptance tests (requires CONTENTFUL_MANAGEMENT_TOKEN,
# CONTENTFUL_ORGANIZATION_ID, CONTENTFUL_SPACE_ID env vars)
task test-acc          # runs: TF_ACC=1 go test -v ./...

# Test coverage
task coverage          # generates coverage.txt

# Build
task build             # goreleaser snapshot build
task build-local       # builds and installs to ~/.terraform.d/plugins/

# Regenerate SDK client and docs after modifying openapi.yaml
task generate          # runs oapi-codegen + go generate ./...
```

## Code Generation

- `openapi.yaml` is the Contentful CMA OpenAPI 3 spec
- `oapi-config.yaml` configures oapi-codegen
- `internal/sdk/main.gen.go` is auto-generated -- never edit manually
- After modifying `openapi.yaml`, always run `task generate`
- Terraform registry docs in `docs/` are auto-generated via `go generate`

## Changelog (Changie)

Every feature, bug fix, or change **must** include a changie entry:

```bash
go run github.com/miniscruff/changie@latest new --kind <KIND> --body "<description>"
```

Kinds: `Added` (minor), `Changed` (major), `Deprecated` (minor), `Removed`
(major), `Fixed` (patch), `Security` (patch), `Dependency` (patch).

## Project Structure

```
internal/
  provider/           # Provider definition, schema, configure
  resources/          # One subdirectory per resource (14 total)
    <resource>/
      resource.go     # CRUD implementation
      model.go        # Terraform state structs + SDK conversion
      resource_test.go
      test_resources/  # .tf files for acceptance tests (some resources)
  datasource/         # Data sources (currently: space)
  sdk/                # Auto-generated API client (do not edit)
  utils/              # Shared helpers (client, types, HCL, HTTP)
  acctest/            # Acceptance test helpers (PreCheck, GetClient)
  custommodifier/     # Custom Terraform plan modifiers
  customvalidator/    # Custom Terraform validators
```

## Code Style

### Imports

Group imports in three blocks separated by blank lines:
1. Standard library
2. External packages (hashicorp, third-party)
3. Internal packages (`github.com/labd/terraform-provider-contentful/internal/...`)

```go
import (
    "context"
    "fmt"

    "github.com/hashicorp/terraform-plugin-framework/resource"

    "github.com/labd/terraform-provider-contentful/internal/sdk"
    "github.com/labd/terraform-provider-contentful/internal/utils"
)
```

### Formatting

- Use `go fmt ./...` before committing (no additional formatter)
- No golangci-lint config file; CI runs it with `--issues-exit-code=0`

### Naming Conventions

- **Resource structs**: unexported camelCase (`spaceResource`, `webhookResource`)
- **Model/state structs**: exported PascalCase (`Space`, `ContentType`, `Webhook`)
- **Constructor functions**: `New<Name>Resource()` / `New<Name>DataSource()`
- **Receiver name**: always `e` for resource methods, `s` or similar for models
- **File names**: `resource.go`, `model.go` (or `models.go`), `resource_test.go`
- **Directory names**: snake_case (`api_key/`, `editor_interface/`)
- **Test functions**: `Test<Resource>Resource_<Operation>` (e.g., `TestSpaceResource_Create`)

### Resource Implementation Pattern

Every resource must:

1. Declare compile-time interface checks:
```go
var (
    _ resource.Resource                = &fooResource{}
    _ resource.ResourceWithConfigure   = &fooResource{}
    _ resource.ResourceWithImportState = &fooResource{}
)
```

2. Implement methods in this order:
   `Metadata` -> `Schema` -> `Configure` -> `Create` -> `Read` -> `Update` -> `Delete` -> `ImportState`

3. Use `Configure` to extract the client from provider data:
```go
func (e *fooResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }
    data := req.ProviderData.(utils.ProviderData)
    e.client = data.Client
}
```

### Model Pattern

Model structs live in `model.go` with `tfsdk` struct tags. Each model has:
- `Import(*sdk.Type)` -- populates from SDK response
- `DraftForCreate()` -- returns SDK create payload
- `DraftForUpdate()` -- returns SDK update payload

### Error Handling

- Use `response.Diagnostics.AddError(title, detail)` in CRUD methods
- Title: `"Error <operation> <resource>"` (e.g., `"Error creating space"`)
- Detail: `"Could not <operation> <resource>: " + err.Error()`
- For HTTP status checks: compare `resp.StatusCode()` against expected codes
- Delete accepts both 204 and 404 as success (idempotent)
- Use `utils.CheckClientResponse(resp, err, expectedStatus)` in newer resources
- Wrap errors with `fmt.Errorf("message: %w", err)` in non-Terraform code

### Types and Utilities

- `utils.Pointer[T](v T) *T` -- generic pointer helper, use instead of `&v`
- `utils.ProviderData` -- container for SDK client + organization ID
- `utils.HCLTemplate()` / `utils.HCLTemplateFromPath()` -- test HCL generation
- `pie/v2` for functional slice operations (`pie.Map`, `pie.FilterNot`)
- `cenkalti/backoff/v5` for retry logic

### Test Conventions

- Acceptance tests use external test package (`package foo_test`)
- Unit tests use internal package (`package foo`)
- Provider factory:
```go
ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
    "contentful": providerserver.NewProtocol6WithError(provider.New("test", true)()),
},
```
- Each test file defines `type assertFunc func(*testing.T, *sdk.Type)`
- Standard helpers: `testAccCheck<Resource>Exists()`, `testAccCheck<Resource>Destroy()`
- Use `acctest.TestAccPreCheck(t)` as `PreCheck` function
- HCL configs via `fmt.Sprintf` or `utils.HCLTemplateFromPath("test_resources/file.tf", params)`
- Acceptance tests require env vars: `CONTENTFUL_MANAGEMENT_TOKEN`, `CONTENTFUL_ORGANIZATION_ID`, `CONTENTFUL_SPACE_ID`

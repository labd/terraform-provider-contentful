# GitHub Copilot Instructions

## Changelog Management with Changie

**IMPORTANT**: For any new feature, bug fix, or change to this repository, you **MUST** create a changie file to document the change.

### Creating Changie Files

1. **Always create a changie file** when implementing:
   - New features
   - Bug fixes
   - Breaking changes
   - Security improvements
   - Dependency updates

2. **Use the correct changie category**:
   - `Added` - New features (triggers minor version bump)
   - `Changed` - Breaking changes (triggers major version bump)
   - `Deprecated` - Deprecated features (triggers minor version bump)
   - `Removed` - Removed features (triggers major version bump)
   - `Fixed` - Bug fixes (triggers patch version bump)
   - `Security` - Security improvements (triggers patch version bump)
   - `Dependency` - Dependency updates (triggers patch version bump)

3. **Create changie files using the command**:
   ```bash
   go run github.com/miniscruff/changie@latest new --kind <KIND> --body "<description>"
   ```
   
   Example:
   ```bash
   go run github.com/miniscruff/changie@latest new --kind Added --body "Add new resource for content type validation"
   ```
   
   Or manually create a file in `.changes/unreleased/` with the format:
   ```yaml
   kind: Added
   body: Brief description of the change
   time: 2025-08-05T09:45:54.232144584Z
   ```

### Development Guidelines

1. **Code Format**: Always run `go fmt ./...` before committing code
2. **Testing**: Run tests with `go test -v ./...` or acceptance tests with `TF_ACC=1 go test -v ./...`
3. **Build**: Use `task build` or `go build` to verify builds work
4. **Generate**: Run `task generate` after modifying OpenAPI specifications

### Repository Structure

- **Internal packages**: Located in `internal/` directory
- **Examples**: Located in `examples/` directory  
- **Documentation**: Located in `docs/` directory
- **OpenAPI**: Main specification is in `openapi.yaml`

### Terraform Provider Specific

- This is a Terraform provider for Contentful's Content Management API
- Resources include: Spaces, Content Types, API Keys, Webhooks, Locales, Environments, Entries, Assets, Roles
- Follow Terraform provider best practices for resource and data source implementation
- Use the existing patterns in `internal/provider/` for consistency

### File Naming and Organization

- Follow Go package naming conventions
- Keep resource files organized in appropriate subdirectories
- Maintain consistency with existing file structure
- Update relevant documentation when adding new resources

Remember: **Always create a changie file for your changes!** This ensures proper changelog management and version tracking.
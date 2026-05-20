# Mendix PVM — Development Guidelines

You're an expert Go developer who:

- Consistently follows coding best practices and the DRY principle
- Keeps solutions simple, maintainable, and easy for others to understand
- Always updates SPEC.md to ensure code and documentation remain consistent

Before starting your work, read @SPEC.md to get to know the current implementation.

## Architecture

This project is a CLI tool (Cobra-based) for managing Mendix Studio Pro versions and projects. See SPEC.md for the complete technical specification.

**Core packages:**

- `config/` — Persistent configuration management (~/.mendix-pvm.json)
- `version/` — Studio Pro version discovery and launching
- `project/` — Mendix project discovery and opening
- `platform/` — Mendix Platform API integration and syncing
- `branch/` — Git branch operations (clone, create, push)
- `convert/` — Project version conversion via mx.exe
- `search/` — Unified search and filtering logic
- `utils/` — Cross-platform file/app launching
- `ui/` — Output formatting

**Design principles:**

- Modular packages with clear separation of concerns
- Platform detection via `runtime.GOOS` for Windows/macOS/Linux compatibility
- Normalized search: lowercase, alphanumeric-only substring matching
- Non-blocking error handling: warn and continue on individual failures
- Direct streaming of tool output (git, mx.exe) to user for debugging

## When Making Changes

1. **Keep packages focused** — each handles one concern (config, version, project, etc.)
2. **Update SPEC.md immediately** — changes to architecture, functions, or flows must be reflected in the spec
3. **Test cross-platform** — verify behavior on Windows, macOS, Linux where applicable
4. **Use existing patterns** — search for similar logic before introducing new approaches
5. **Validate at boundaries** — only validate user input; trust internal code and framework guarantees

# mx (Mendix Project and Version Manager)

`mx` is a command-line tool for quickly working with Mendix Studio Pro:

- Search local Studio Pro versions and Mendix projects
- Open projects or Studio Pro directly from the terminal
- Convert projects between Mendix versions
- Inspect and edit configuration

The CLI supports flexible searching using name fragments. All terms must match somewhere in the name after normalization (case-insensitive; letters and digits only).

## Quick start

1. Build the binary (see Build and install).
2. Run `mx list` to see discovered projects and versions.
3. Run `mx open <name>` to open a project or Studio Pro.
4. Run `mx path <name>` to print a single resolved path (useful for cd).

## Build and install (local)

Go version is defined in go.mod: Go 1.25.3.

Build the binary in the repo root:

```powershell
# From the repository root
$env:GOOS = "windows"  # optional, if cross-building
$env:GOARCH = "amd64"  # optional, if cross-building

go build -o mx.exe .
```

For non-Windows shells, you can build `mx`:

```bash
go build -o mx .
```

The intended binary name is `mx` (or `mx.exe` on Windows).

## OS support

- Windows is the primary supported OS.
- Linux has partial support for opening files using `xdg-open`, but Studio Pro executables referenced by `mx` are Windows-specific.
- macOS is not supported; config creation returns an explicit error.

## Configuration

On first run, `mx` creates a config file if it does not exist:

- Path: `$HOME/.mendix-pvm.json` (Windows example: `C:\Users\<you>\.mendix-pvm.json`)

Configuration keys:

```json
{
    "VersionDirectory": "C:\\Program Files\\Mendix",
    "ProjectDirectory": "C:\\Users\\<you>\\Mendix"
}
```

Defaults and behavior:

- `VersionDirectory`
  - Windows default: `%ProgramFiles%\Mendix`
  - Linux default: empty string
  - macOS: config creation fails
  - Must be an existing directory for subsequent runs; otherwise `mx` exits with a validation error.
- `ProjectDirectory`
  - Default: `$HOME\Mendix`
  - Must be an existing directory for subsequent runs.

To edit the config, run:

```powershell
mx config
```

This opens the config in your default editor using OS-specific file open helpers.

## Search and matching rules

- Searches only the direct subdirectories of `ProjectDirectory` and `VersionDirectory`.
- Each query token is normalized to letters and digits only and matched case-insensitively.
- All tokens must match somewhere in the directory name.
- Projects are recognized only if a `.mpr` file exists at the project root.
- Versions are recognized only if the directory contains a `modeler` subdirectory.

## Commands

### `mx list`

List Mendix projects and Studio Pro versions.

Flags:

- `--project`, `-p`: search and show projects only
- `--version`, `-v`: search and show Studio Pro versions only

Behavior:

- By default, searches both projects and versions.
- You can limit the search with `--project` or `--version`.
- If no matches are found, prints `No matches found.`
- Output lists only base directory names (not full paths).

Examples:

```powershell
mx list
mx list widget
mx list 10.6 -v
mx list Customer -p
```

### `mx open`

Open a Mendix project or a Studio Pro version.

Flags:

- `--project`, `-p`: search only projects
- `--version`, `-v`: search only Studio Pro versions
- `--all`, `-a`: open up to 3 matches instead of listing them

Behavior (current implementation):

- Searches projects and versions unless `--project` or `--version` is used.
- If exactly one match is found, it opens that item.
- If multiple matches are found and `--all` is not used, it prints a list and exits.
- With `--all`, it opens up to 3 matches.
- If no matches are found, prints `No matches found.`

Note: The command help text says "No arguments opens the latest Studio Pro," but the current implementation does not select a latest version. With no arguments, it behaves like a broad search and usually lists multiple matches.

Examples:

```powershell
mx open MyApp
mx open 10.6
mx open MyApp --project
mx open 10 --version -a
```

### `mx path`

Print the directory for a single project or Studio Pro version. This is useful for `cd`.

Flags:

- `--project`, `-p`: search only projects
- `--version`, `-v`: search only Studio Pro versions

Behavior:

- Searches projects and versions unless `--project` or `--version` is used.
- Requires exactly one match.
- If zero matches: returns error `no matches found`.
- If multiple matches: prints a list and returns error `multiple matches`.
- Output is the project or version directory (not the `.mpr` file).

Examples:

```powershell
mx path MyApp --project
mx path 10.10 --version
mx path "next dev"
```

### `mx convert`

Convert one or multiple Mendix projects to a specific Studio Pro version.

Flags:

- `--project`, `-p` (required): project name fragment to search for
- `--version`, `-v` (required): version string to search for
- `--all`, `-a`: convert all matching projects (otherwise only the first match)

Behavior:

- Version search must resolve to exactly one match; otherwise a list is shown and no conversion runs.
- Project search can yield one or many results. Without `--all`, only the first match is converted.
- Conversion runs `mx.exe` from the version's `modeler` directory with `convert --in-place`.
- Exit codes from `mx.exe` are reported per project:
  - `0`: success
  - `1`: internal error
  - `2`: error with options
  - `3`: conversion failed
  - any other code: treated as unexpected error

Examples:

```powershell
mx convert -p OrderApp -v 10.10
mx convert -p CustomerPortal -v 10.6
mx convert -p App -v 10.10 -a
```

### `mx config`

Open the configuration file in your default editor.

Example:

```powershell
mx config
```

## Quick cd into projects

`mx path` prints a single directory path. You can wrap it in a PowerShell function so you can `cd` directly:

```powershell
function mxcd {
    param(
        [Parameter(ValueFromRemainingArguments = $true)]
        [string[]]$Args
    )

    $path = & mx path @Args
    if ($LASTEXITCODE -ne 0) {
        return
    }
    if ([string]::IsNullOrWhiteSpace($path)) {
        return
    }

    Set-Location $path
}
```

Examples:

```powershell
mxcd MyApp
mxcd 10.10 --version
mxcd "next dev" --project
```

## Troubleshooting / FAQ

- No matches found
  - Your `ProjectDirectory` or `VersionDirectory` may be empty or incorrect. Check the config with `mx config`.
  - Your search tokens might be too strict. All tokens must match after normalization.

- Multiple matches
  - Refine your search terms or add `--project` or `--version` to limit scope.
  - For `mx open`, use `--all` to open up to 3 matches.
  - For `mx convert`, use `--all` to convert all matching projects.

- Project is not detected
  - A project must contain a `.mpr` file at the root directory. Nested `.mpr` files are ignored.

- Version is not detected
  - A version directory must contain a `modeler` subdirectory.

- Convert fails with missing `mx.exe`
  - `mx convert` expects `mx.exe` at `<VersionDirectory>\<version>\modeler\mx.exe`.
  - Ensure the version directory points to an installed Studio Pro with `mx.exe`.

- `mx` fails to start after config creation on Linux
  - The default `VersionDirectory` is empty on Linux and fails validation on subsequent runs. Set a valid directory in the config file.

## Verification steps

- Build locally with `go build -o mx.exe .` and run `./mx --help` plus `./mx <cmd> --help` to confirm help text.
- Manually test `mx path` and the `mxcd` function with a known project.

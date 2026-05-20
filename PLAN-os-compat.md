# Plan: OS Compatibility

## Goal

Support Windows, macOS, Linux, and WSL as first-class platforms. WSL uses `wslview`
to open files and applications on the Windows side. Clear errors are surfaced when
dependencies like `wslview` are not installed.

---

## Current State

| Area | Windows | macOS | Linux | WSL |
|---|---|---|---|---|
| `utils.OpenFile` | `cmd /c start` | `open` | `xdg-open` | falls through to `xdg-open` (wrong) |
| `config.Open` | `cmd /c start` | `open` | `xdg-open` | falls through to `xdg-open` (wrong) |
| `config.create` | `%ProgramFiles%\Mendix` | **hard error** | empty | falls through to Linux |
| `config.persistPAT` | `setx` | shell export hint | shell export hint | falls through to Linux |
| `version.Open` | `studiopro.exe` | `studiopro.exe` (wrong) | `studiopro.exe` (wrong) | `studiopro.exe` (correct, runs via WSL) |

---

## Changes

### 1. Platform detection — `utils/platform.go` (new file)

Add two functions:

```go
// IsWSL reports whether the process is running inside Windows Subsystem for Linux.
// Detection: /proc/version contains "microsoft" (case-insensitive).
func IsWSL() bool

// Platform returns the effective platform string: "windows", "darwin", "wsl", or "linux".
func Platform() string
```

WSL reports `runtime.GOOS == "linux"`, so the check must happen before the generic Linux
branch everywhere.

---

### 2. `utils/file.go` — add WSL case

```go
func OpenFile(path string) error {
    switch utils.Platform() {
    case "windows":
        cmd = exec.Command("cmd", "/c", "start", "", path)
    case "darwin":
        cmd = exec.Command("open", path)
    case "wsl":
        if err := checkWslview(); err != nil {
            return err
        }
        cmd = exec.Command("wslview", path)
    default: // linux
        cmd = exec.Command("xdg-open", path)
    }
    ...
}

// checkWslview returns a clear, actionable error if wslview is not on PATH.
func checkWslview() error {
    if _, err := exec.LookPath("wslview"); err != nil {
        return fmt.Errorf("wslview is not installed — install wslu with: sudo apt install wslu")
    }
    return nil
}
```

---

### 3. `config/config.go` — `Open` deduplication

`config.Open` currently duplicates the same platform switch as `utils.OpenFile`. Replace
its body with a single call to `utils.OpenFile(configPath)`. This removes duplication and
gets WSL support for free.

---

### 4. `config/config.go` — `create()` defaults

| Platform | `versionDirectory` default |
|---|---|
| Windows | `%ProgramFiles%\Mendix` |
| WSL | `/mnt/c/Program Files/Mendix` (Windows Mendix installs are accessible via WSL mount) |
| macOS | `""` (user must set; remove the hard error — macOS Studio Pro exists) |
| Linux | `""` (user must set) |

Remove the `return Config{}, fmt.Errorf("macOS is not supported")` line. macOS users
should reach the interactive wizard and set their own path.

---

### 5. `config/config.go` — `persistPAT`

| Platform | Behavior |
|---|---|
| Windows | `setx MX_PAT <value>` (existing) |
| WSL | Print export hint for `~/.bashrc` / `~/.zshrc` (same as Linux; `setx` is not appropriate inside WSL) |
| macOS / Linux | Print export hint (existing behavior) |

No code change needed here beyond ensuring WSL resolves to the hint path rather than
accidentally hitting Windows. Since WSL `runtime.GOOS == "linux"`, the existing `default`
branch already handles this correctly — no change required.

---

### 6. `version/version.go` — `Open` platform guard

Currently hardcodes `studiopro.exe`. This is correct for Windows and WSL (both run the
Windows binary). On bare Linux or macOS, the binary name would be different (or Studio Pro
may not be installed at all in that form).

Add a platform check:

```go
func Open(versionPath string) error {
    modelerPath, err := FindModelerSubdir(versionPath)
    if err != nil {
        return err
    }

    var exe string
    switch utils.Platform() {
    case "windows", "wsl":
        exe = filepath.Join(modelerPath, "studiopro.exe")
    case "darwin":
        exe = filepath.Join(modelerPath, "studiopro") // placeholder; adjust if macOS uses a .app bundle
    default:
        return fmt.Errorf("launching Studio Pro is not supported on this platform (%s)", runtime.GOOS)
    }

    return exec.Command(exe, "--enable-extension-development").Start()
}
```

> **Open question:** What is the correct Studio Pro binary path on macOS? If it ships as a
> `.app` bundle, `open -a <path>` would be more appropriate than executing the binary
> directly. Confirm before implementing the macOS case.

---

## Implementation Order

1. `utils/platform.go` — platform detection (no dependencies, everything else needs it)
2. `utils/file.go` — WSL case + `checkWslview`
3. `config/config.go` — deduplicate `Open`, fix `create()` defaults
4. `version/version.go` — platform guard on `Open`
5. Update `SPEC.md` to reflect new platform support and `utils/platform.go`

---

## Configurable Paths

Default paths are hardcoded per platform but users may need to override them (e.g. Windows
drive not mounted at `/mnt/c`, custom Studio Pro install location).

All platform-specific default paths — `versionDirectory` and `projectDirectory` — are
already stored in the config file (`~/.mendix-pvm.json`) and editable via `mx config`.
No new config fields are needed; the existing fields serve as the override mechanism.

The only requirement is that `create()` sets sensible per-platform defaults on first run.
After that, users edit the config directly to change any path.

---

## Out of Scope

- macOS Studio Pro binary path (needs research/confirmation before implementation)
- Full Linux Studio Pro support (no native Linux build exists at time of writing)
- `convert/convert.go` — `mx.exe` is Windows-only; conversion on non-Windows platforms
  is out of scope for this change

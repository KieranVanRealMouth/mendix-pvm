# Mendix Project Manager

A CLI tool for managing Mendix versions and projects.

## Features

The CLI offers the following features:

- View installed Mendix Studio Pro versions
- View downloaded Mendix projects
- Open Mendix Studio Pro versions
- Open projects in Studio Pro
- Convert a project in place to a new version of studio pro

## Technical Implementation

The CLI tool is built using Go.

It uses bubble tea for the TUI.

The users Mendix Studio Pro versions are installed on the users system in a specific folder. The default location is `C:\Program Files\Mendix\`.

The users their Mendix projects can be stored somewhere on their system. The tool scans preconfigured folders on the users system for folders which have a `.mpr` file.

An example of the users configuration:

```json
{
    "version_directory": "C:\\Program Files\\Mendix\\",
    "project_directories": [
        "C:\\Users\\username\\Mendix\\", "C:\\Users\\username\\Documents\\"
    ]
}
```

### Versions

#### Finding Installed Versions

The project versions are found in the specified directory from the configuration. Within this directory, a directory can be found for each version. Named after the version number. The version number is built like:

- Major: `24`
- Minor: `5`
- Patch: `3`
- Build: `24567`

Within the directory of a version, for example `24.5.3.24567`.

Within the version's directory, 2 directories are found:

- `modeler`: Contains files related to Mendix Studio Pro. This must be considered.
- `runtime`: Contains the files needed to start the Mendix runtime for a project. This can be ignored.

Within the `modeler` directory, the following files can be found:

- `studiopro.exe`: The executable to open Mendix Studio Pro in the version.
- `mx.exe`: A command line tool provided by Mendix to convert a project in place.

#### View Versions

Project versions are displayed based on their folder

#### Open Version

To open a version, the tool simply opens the studiopro.exe file of the version.

### Projects

#### Finding Installed Projects

The user will configure in what directories the tool should look for projects.

To find a project in the specified directories, the program looks for directories within the configured directory. A directory is identified as a Mendix project when it has a `.mpr` file in the root of it's project. This is the file used as input by the `studiopro.exe` file.

The tool does not perform a recursive search deeper than what is needed. Meaning it looks in the specified directory for directories, and in those directories it should find a `.mpr` file in their root (not in further sub-directories).

#### View Project

When projects are viewed, the user is shown the name of the directory which has the `.mpr` file.

#### Open Project

To open a project, simply opening the `.mpr` file with the default program on the system would open it in it's correct version.

#### Convert Project In Place

To convert a project to a newer version, the provided `mx.exe` tool is used. The `mx.exe` tool provides the `convert` command. Which takes the flag of `--in-place` and the route to the `.mpr` file of the project to convert. It will convert then the project to the newer version.

### Search

The tool offers a user friendly way to search for versions and projects and select them.

The user can search by providing a search string. It may be seperated with spaces. For example `"approval feat user edit"` should open the project `Approval-Tool_feat_user-edit`.

When the search string is more generic, multiple results may be found. For example `"10"` may result in all versions 10, as well as a project with 10 in it's name. Both should be displayed. But clearly differentiated.

### Commands

Instead of providing a string like `mx list "approval main"` the must also be able to use `mx list approval main` for all commands.

#### List

```bash
mx list
```

Will list all versions and projects installed.

User can also filter:

```bash
mx list "approval 10"
```

The user can also only list versions

```bash
mx list versions

# As well as filter
mx list versions "10 24"
```

The same can be done for projects

```bash
mx list projects 

mx list projects "approval main"
```

#### Open

```bash
mx open
```

When the user provides no search string, he is prompted to select what he wants to open, from a list of versions and projects.

The user may provde a search string:

```bash
mx open "approval"
```

If there're multiple results, then the user can select from the list.

The user can select using arrow- and enter keys.

To improve user experience in edge cases, the open command also provides flags:

```bash
# Only consider versions
mx open -version "approval"
mx open -v "approval"

# Only consider projects
mx open --project "approval"
mx open -p "approval"
```

#### Convert

```bash
mx convert
```

Will convert a project to a version. If no version is provided, the tool will list the versions. Allowing the user to select a version using the arrow keys and the enter key. If no project is provided, the tool will list the projects and allow the user to select one using the arrow- and enter keys.

```bash
mx convert --version="10 24 9" --project="approval feat user edit"
```

Or using the short hands:

```bash
mx convert -v="10 24 9" --project="approval feat user edit"
```

#### 

#### Edit Config

```bash
mx config
```

Opens the configuration file. The file is found at `~/.mendix-project-manager.json`.

### TUI

The UI is triggered conditionally. If the user follows the happy path, he may not need the TUI to open, as is defined in previous paragraphs.

For components, the program uses bubble tea's bubbles. The following are used:

- Help: To show the support text which is also shown when the CLI commands are called from the terminal using `--help`.
- List: When the user has defined a search query that matches multiple search results.
- Table: When the user has not defined a search query. When both versions and projects are displayed, each have their own table.
- Spinner: When the program is loading.

Navigation is always done using the arrow up and down keys.

Selection is always done using the enter key.

# zen CLI Project Structure

## High-Level Structure

```sh
zen/
├── .github/                    # GitHub workflows and templates
├── .vscode/                    # VS Code workspace configuration
├── build/                      # Build artifacts and packaging
├── cmd/                        # Zen CLI main entry point
├── docs/                       # Zen Documentation
├── internal/                   # Private Go packages specific to Zen
├── pkg/                        # Public Go packages (APIs for plugins)
├── plugins/                    # Official plugins
├── scripts/                    # Build, development, and utility scripts
├── templates/                  # Built-in templates (migrated from existing)
├── test/                       # Integration and end-to-end tests
├── tools/                      # Development tools and generators
├── ...
├── go.mod
└── README.md
```

**Key Packages**

At a high level, these areas make up the `zen` project:
- [`cmd/zen/`](../cmd/zen/) - `main` packages for building binaries such as the `zen` executable
- [`pkg/`](../pkg) - most other packages, including the implementation of individual zen commands
- [`docs/`](../docs) - documentation for maintainers and contributors
- [`scripts/`](../scripts) - build and release scripts
- [`internal/`](../internal) - Go packages hizenly specific to our needs and thus internal
- [`go.mod`](../go.mod) - external Go dependencies for this project, automatically fetched by Go at build time

## Command-line help text

Running `zen help issue list` displays help text for a topic. In this case, the topic is a specific command, and help text for every command is embedded in that command's source code. The naming convention for zen commands is:

```sh
pkg/cmd/<command>/<subcommand>/<subcommand>.go
```

// TODO - Update this process

Following the above example, the main implementation for the `zen issue list` command, including its help text, is in [pkg/cmd/issue/list/list.go](../pkg/cmd/issue/list/list.go)

Other help topics not specific to any command, for example `zen help environment`, are found in [pkg/cmd/root/help_topic.go](../pkg/cmd/root/help_topic.go).

During our release process, these help topics are [automatically converted](../cmd/gen-docs/main.go) to manual pages and published under https://zen.<domain>.com/docs/.

---


// TODO - Update this section

## How zen CLI works

To illustrate how zen CLI works in its typical mode of operation, let's build the project, run a command, and talk through which code gets run in order.

1. `go run script/build.go` - Makes sure all external Go dependencies are fetched, then compiles the
   `cmd/zen/main.go` file into a `bin/zen` binary.
2. `bin/zen issue list --limit 5` - Runs the newly built `bin/zen` binary (note: on Windows you must use
   backslashes like `bin\zen`) and passes the following arguments to the process: `["issue", "list", "--limit", "5"]`.
3. `func main()` inside `cmd/zen/main.go` is the first Go function that runs. The arguments passed to the
   process are available throuzen `os.Args`.
4. The `main` package initializes the "root" command with `root.NewCmdRoot()` and dispatches execution to it
   with `rootCmd.ExecuteC()`.
5. The [root command](../pkg/cmd/root/root.go) represents the top-level `zen` command and knows how to
   dispatch execution to any other zen command nested under it.
6. Based on `["issue", "list"]` arguments, the execution reaches the `RunE` block of the `cobra.Command`
   within [pkg/cmd/issue/list/list.go](../pkg/cmd/issue/list/list.go).
7. The `--limit 5` flag originally passed as arguments be automatically parsed and its value stored as
   `opts.LimitResults`.
8. `func listRun()` is called, which is responsible for implementing the logic of the `zen issue list` command.
9. The command collects information from sources like the GitHub API then writes the final output to
   standard output and standard error [streams](../pkg/iostreams/iostreams.go) available at `opts.IO`.
10. The program execution is now back at `func main()` of `cmd/zen/main.go`. If there were any Go errors as a
    result of processing the command, the function will abort the process with a non-zero exit status.
    Otherwise, the process ends with status 0 indicating success.

---

// TODO - Update this section

## How to add a new command

1. First, check on our issue tracker to verify that our team had approved the plans for a new command.
2. Create a package for the new command, e.g. for a new command `zen boom` create the following directory
   structure: `pkg/cmd/boom/`
3. The new package should expose a method, e.g. `NewCmdBoom()`, that accepts a `*cmdutil.Factory` type and
   returns a `*cobra.Command`.
   * Any logic specific to this command should be kept within the command's package and not added to any
     "global" packages like `api` or `utils`.
4. Use the method from the previous step to generate the command and add it to the command tree, typically
   somewhere in the `NewCmdRoot()` method.


// TODO - Update this section

## How to write tests

This task might be tricky. Typically, zen commands do things like look up information from the git repository
in the current directory, query the GitHub API, scan the user's `~/.ssh/config` file, clone or fetch git
repositories, etc. Naturally, none of these things should ever happen for real when running tests, unless
you are sure that any filesystem operations are strictly scoped to a location made for and maintained by the
test itself. To avoid actually running things like making real API requests or shelling out to `git`
commands, we stub them. You should look at how that's done within some existing tests.

To make your code testable, write small, isolated pieces of functionality that are designed to be composed
together. Prefer table-driven tests for maintaining variations of different test inputs and expectations
when exercising a single piece of functionality.

# nutshell

**English is the source code. The binary is the product.**

Given the death of hand-written code, Nutshell repurposes good old English as a
syntax-free, open-source programming language.

That is the bet. If an AI is writing the implementation anyway, the human should
write what the program does. Behavior. Constraints. Examples. Failure cases.
The compiler can deal with the semicolons.

Write a `.nut` file. Pick your AI compiler. Get a native executable.

```text
main.nut → AI compiler → implementation + tests → native binary
```

The AI chooses the implementation language and toolchain. Go, Rust, C, or another
native target it can build on your machine. Nutshell runs the declared build and
tests, checks the executable, then delivers the binary and its implementation
source. The Nutshell driver itself is written in Go.

## A program, in a nutshell

Save this as `main.nut`:

```text
Create a command-line program called greet.

With no arguments, print "Hello, world!" followed by a newline.
With one argument, greet that name instead.
With more than one argument, print "Usage: greet [name]" to standard error
and exit with status 2. Print no greeting in that case.

Do not access the network or write files.
```

Compile it:

```sh
nutshell main.nut -c codex -s ./src -o greet
./greet Ada
# Hello, Ada!
```

Use `-c grok`, `-c claude`, or [configure another CLI](docs/cli.md#compilation-agents).
There is no prescribed model. Each tool uses its configured default.

The local installer also provides `ns`, a short name for the same executable:

```sh
ns main.nut -c codex -s ./src -o greet
```

For a Go-installed binary, see the [one-time alias setup](docs/releases.md#the-ns-command).

`-s ./src` is explicit: that directory supplies existing implementation context
and receives the updated code. Edit your `.nut` file, run the same command, and
the AI is instructed to adapt the existing implementation, including manual edits.

## Install

Nutshell is early software. This checkout is **0.8.0**; the latest tagged release
is **0.7.0**. Linux is the primary tested host.

Install with Go 1.26 or newer:

```sh
go install github.com/ryangerardwilson/nutshell@latest
```

Add `$(go env GOPATH)/bin` to `PATH` (or your `GOBIN` if set). To install directly
into `~/.local/bin`, prefix the command with `GOBIN="$HOME/.local/bin"`.
Use `@v0.7.0` to pin the latest tagged release; repeat `@latest` to upgrade.

For local development, run `./install.sh from "$PWD"` inside the checkout. This
builds your current source and installs into `~/.local/bin`, or the directory set
by `NUTSHELL_INSTALL_DIR`. It does not install AI providers or fetch skills.
See [installation and releases](docs/releases.md) for local updates and version policy.

To compile `.nut` programs, you also need:

- An installed, authenticated Codex, Grok, Claude, or configured AI CLI.
- **`rgw-ast` on `PATH`**, required for agent code inspection and edits.
- A native toolchain the agent can use, such as Go, Rust, or a C compiler.

**Current distribution gap:** `rgw-ast` is a separate prerequisite and is not
bundled here. This repository does not yet provide a public installation path for
it. You can build Nutshell and run its tests without it; actual AI compilation
requires an existing installation. Making that dependency accessible is necessary
for a complete public onboarding path.

Provider access may cost money. Nutshell does not include a model subscription.

## The premise

Programming languages made humans translate intent into instructions a machine
could execute. AI changes where that translation can happen.

Nutshell puts the translation in the compiler. The source is the thing you meant.
The implementation is one way to make it happen.

This is a claim about the interface we want to build, not a claim that AI has
solved software correctness. English can be ambiguous. Models can be wrong.
Generated tests can share the same blind spots as generated code. Nutshell checks
that an implementation builds and passes its tests; it cannot prove that it
understood you. Keep the source, inspect the implementation, test the behavior.

## The rules

- **Describe outcomes.** Names, inputs, outputs, invariants, examples, and errors
  belong in the program. Specify implementation details when they matter to you.
- **The compiler makes the calls.** It resolves ambiguity, records assumptions,
  and keeps moving. Compilation has no question-and-answer phase.
- **Bring your compiler.** Use your chosen AI CLI and its default model.
- **Keep the result.** You get a native binary and ordinary implementation source.
  The executable does not need Nutshell to run. Its own runtime dependencies still apply.
- **Compile visibly.** Short, changing activity summaries show what the agent
  reports it is doing. Build logs stay in the temporary workspace.

“Syntax-free” means no mandatory grammar for describing program behavior.
There is one small driver directive for composing files:

```text
Include "rules.nut".
```

Everything else is passed to the AI as written. Read the
[language guide](docs/language.md) for programs that leave less room for guessing.

## What happens on your machine

All development, building, and testing happen in a private `/tmp` directory.
After verification, Nutshell publishes the executable and the source selected by
`-s`. Failed builds preserve the previous outputs. Concurrent source edits stop
publication so a long compile does not overwrite newer work.

**AI compilers run unattended with full permissions by default.** `/tmp` is a
working directory, not a security sandbox. The AI and generated build commands
can act with your user's permissions. Compile trusted programs in an environment
whose access you are prepared to give the agent. See the
[execution model](docs/cli.md#compilation-agents) and [security policy](SECURITY.md).

Every AI receives a mandatory `rgw-ast` workflow for inspecting and changing code.
Nutshell checks that the tool exists; it does not audit or enforce every agent
action. No provider hooks or global configuration are installed.

## Read more

| Document | What it covers |
| --- | --- |
| [Language guide](docs/language.md) | Writing `.nut` programs, examples, includes, assumptions |
| [CLI reference](docs/cli.md) | Flags, compiler adapters, source reuse, progress, failures |
| [Architecture](docs/architecture.md) | The Go driver, compiler contract, verification, publication |
| [Examples](examples/) | Small programs you can read and compile |
| [Installation and releases](docs/releases.md) | One-command install, local updates, version policy and releases |
| [Contributing](CONTRIBUTING.md) | Local development, tests, OpenSpec, useful contributions |
| [Accepted specification](openspec/specs/nutshell/spec.md) | The current behavioral contract |

## Build with us

Nutshell is small enough to understand and early enough to change.
Help make English programs precise, compilation inspectable, and generated
executables worth trusting. Bring failing examples and concrete improvements.

[MIT licensed](LICENSE). The source, the compiler driver, and the conversation
about what this language should become are open.

# Writing Nutshell

Write the program you want. Skip the ceremony.

A `.nut` file is UTF-8 text. `main.nut` is the default entry point. There are no
required class declarations, types, braces, keywords, or indentation rules for
describing behavior. The AI compilation agent reads your text and implements it.

Plain English is the intended interface. Markdown headings and lists can help
organize a longer program; they do not introduce special language semantics.

## Make the observable behavior concrete

This is a complete program:

```text
Print "Hello, world!" followed by a newline, then exit successfully.
```

For a useful command, describe the inputs, outputs, and cases that matter:

```text
Create a command-line program that totals integers from standard input.

Read one signed decimal integer per nonblank line. Ignore blank lines.
Accept surrounding whitespace. Use signed 64-bit arithmetic.

On success, print the total followed by a newline and exit with status 0.
An empty input has a total of zero.

If a line is not a signed decimal integer or the running total overflows,
print an error naming that line number to standard error and exit with
status 2. Print no total. Do not silently clamp or wrap numbers.

Examples:
Input "2\n3\n" produces "5\n".
Input "-4\n\n9\n" produces "5\n".
Input "2\nhello\n" fails on line 2.

Do not access the network or write files.
```

The examples are source, too. The AI is asked to generate tests of observable
behavior. It may still misunderstand both the prose and examples; review the
generated tests for requirements you particularly care about.

## Say what matters; leave the rest to the compiler

Useful constraints include:

- Input formats, valid ranges, and expected error messages.
- Ordering, duplicate handling, time zones, rounding, and persistence rules.
- Runtime requirements, resource limits, and deployment targets.
- Whether the program may use the network or depend on external services.

You can require Go, a particular database client, or a specific algorithm. If you
do not, the AI chooses an implementation it can compile into a native executable
on the host. The driver does not select a language for you or provide a cross
compilation flag.

Avoid assuming that words such as “fast,” “secure,” or “production-ready” specify
testable behavior. Name the property you need: a maximum input size, an error
condition, a permission boundary, or a measurable target.

## Split larger programs into files

`main.nut`:

```text
Include "rules.nut".

Create a command-line program that applies the included pricing rules.
Accept the quantity as the only argument and print the total in cents.
Reject invalid quantities with exit status 2 and a message to standard error.
```

`rules.nut`:

```text
A quantity is an integer from 1 through 1000, inclusive.
Each item costs 250 cents.
For quantities of at least 10, discount the total by 10 percent.
Use integer arithmetic; these prices never require rounding.
```

`Include` in that example is wording for the AI, not a directive recognized by
the driver. You could instead write `from rules.nut import pricing`, or simply
“Use the pricing rules described in rules.nut.” The AI decides what references,
symbols, and their spelling mean. No import syntax is mandatory or privileged.

The driver selects the explicit entry (default `main.nut`) and discovers available
`.nut` files recursively beneath its directory. Their exact text and relative
paths are supplied together. Other files named `main.nut` can exist in nested
folders; they are available context, not additional selected entry points.

Keep intended `.nut` context in this directory tree. Discovery skips hidden
directories and `node_modules`, `vendor`, `target`, `dist`, and `build`. Directory
symlinks are not followed; `.nut` file symlinks are rejected. Source must be
nonempty UTF-8, at most 128 files and 1 MiB total. Files outside that boundary
are not automatically fetched because a sentence mentions them. The compiler
makes assumptions about missing information as it does for other ambiguity.

The temporary workspace records the complete input bundle, including available
files the AI may not use. The binary does not embed these inputs. Move unrelated
`.nut` files outside the source tree if they should not be supplied. Adding or
removing an available file changes the bundle even when `main.nut` stays the same.

## Evolve a program

```sh
nutshell main.nut -c codex -s ./implementation -o app
# Change main.nut, then run the same command again.
nutshell main.nut -c codex -s ./implementation -o app
```

Nutshell copies the selected implementation into `/tmp`. The agent is instructed
to inspect and adapt it, preserving useful tests, assets, and manual work.
Requested behavior in `.nut` takes precedence where it conflicts with the old
implementation. Files outside the selected source directory are not implicitly
included as implementation context.

For a focused edit against that implementation:

```sh
ns main.nut -c grok -s ./implementation -o app -f "Reject negative quantities"
```

The `.nut` instructions take precedence. The AI first reviews all supplied `.nut`
files; a conflicting request fails with file/line diagnostics. To change a required
behavior, update the `.nut` instructions first. Compatible fixes preserve unrelated
behavior, language and architecture, and still undergo build/test verification.
`.nut` inputs are never rewritten by fine tuning. Append `-l 5` to tell the AI that
you expect completion within five minutes; `--timeout` sets a separate hard cutoff.

The program file is the durable statement of intent. Record a behavior change
there even if you also fix the generated implementation by hand.

## Ambiguity is a compiler decision

The compiler does not interview you. Missing details, conflicting requirements,
and design choices become assumptions in its internal build manifest. If you
want a particular choice, write it into the program and compile again. Fine tuning
is the exception: it must reject conflicts with `.nut` instructions rather than
resolve them by assumptions.

This makes compilation unattended. It also means you should expect different
implementations across models and runs. Nutshell does not currently offer
reproducible AI generation, formal semantics, or a proof that English intent and
executable behavior match.

See [the compiler contract](architecture.md) for what the driver actually checks.

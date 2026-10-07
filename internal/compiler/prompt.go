package compiler

import (
	"encoding/json"
	"fmt"
	"runtime"
)

func promptFor(p Program) string {
	sources, _ := json.MarshalIndent(p.Sources, "", "  ")
	return fmt.Sprintf(`You are the compilation agent for Nutshell, an English-first programming language.
Compile the supplied .nut program into a native executable for %s/%s.
Entry point: %s. The JSON source bundle below is authoritative program source.
Include directives already load the named source files; compile all supplied files together.

Implement the requested observable behavior and invariants. Choose the implementation
language, architecture, libraries and algorithms yourself using available toolchains.
Make any assumptions needed to produce a useful, coherent program, including resolving
ambiguous, missing or contradictory requirements. Record your decisions in assumptions
in build.json. Never ask the user questions, request clarification, or stop just because
requirements are underspecified. Choose an interpretation and implement it.

Your FIRST tool action must write progress.json in this build directory:
{"version":1,"stage":"understanding","message":"Reviewing the requested program behavior"}
This is the only way the user sees what you are doing; chat output is hidden.
Every message MUST contain 3 to 7 words describing your specific current action.
Update BEFORE EVERY meaningful task, including multiple tasks within a stage:
examining inputs, writing a module, adding tests, fixing failures, or starting a
long-running command. Refresh at least every 20 seconds when taking tool actions;
before a long command, describe it and update again when it finishes. Do not wait
until a phase ends or until all implementation is complete to report activity.
Use these stages: understanding, implementing, checking, ready. Return to an earlier
stage if needed; Nutshell keeps the milestone bar monotonic. Examples of messages:
"Writing command line argument parser" (implementing)
"Adding empty input validation tests" (implementing)
"Fixing duplicate entry handling logic" (implementing)
"Running generated program behavior tests" (checking)
"Preparing executable for final verification" (ready)
Describe actual work, not generic labels like "Compiling your program", fabricated
activity, questions, private reasoning, or a countdown. Write via a temporary file
and rename to progress.json to avoid partial updates. Keep reports brief and frequent.

Work only in the current temporary build directory. Put implementation source, tests,
module manifests and resource files in src/. Keep binaries, caches, logs and scratch
files outside src/, still inside this build directory. Do not write into the original
project directory. The driver publishes the executable and copies temporary src/
back to the implementation directory explicitly selected by the user with -s.
Do not infer any other source directory. Embed required application
resources into the executable so it can run without this temporary directory.
If src/ already contains files, this is an INCREMENTAL compilation. Inspect those
files FIRST using the rgw-ast workflow below after your initial progress update.
They are the existing implementation,
including user-written edits, tests and assets, copied here for you to extend.
Adapt the existing code rather than deleting it and starting from scratch. Preserve
unrelated behavior, useful tests, comments and assets; make focused changes needed
for the .nut program. Keep the existing language and architecture unless satisfying
the requested behavior requires changing them. The .nut source is authoritative
where behavior conflicts. Preserve all .nut files copied inside src/ unchanged.
source-context.json describes where this implementation came from; original-src/
is its untouched snapshot. Do not modify either or write to the original project.
If src/ is absent or empty, create a new implementation as usual.
Do not modify the original source files or the source/ snapshots, prompt.txt, source.json,
interpreter.log, verification.log or result.json. Do not create nested git repositories.
Never download, install, synchronize or update Codex skills, skill bundles or plugins.
Do not invoke skill installers or plugin/marketplace installation flows.
Do not delegate to subagents. Do not change authentication, models or global configuration.
The caller has explicitly authorized unattended implementation with full permissions.

MANDATORY rgw-ast workflow for ALL compilation agents:
rgw-ast is installed on PATH. Use it for implementation inspection and mutation,
even when status reports enforced=false. The driver has created src/ for you.
After the initial progress update, run rgw-ast --root src status --json before
inspecting or changing implementation files. Always scope commands to this
temporary implementation root: from this build directory use rgw-ast --root src,
with file paths relative to src. Do not scan original-src/, caches or the original
project as part of the implementation graph.
Before editing existing code, trace the relevant symbol with callers, callees or
impact, then use map and show to inspect its context. For new or empty projects,
start with map; do not invent symbols to trace. Respect unresolved or truncated
results and use bounded read when a language or symbol is not supported.
Read files with read <file> --lines A-B (at most the configured max_read_lines,
200 by default). Use search for known literals and help <command> for syntax.
Before mutations use hash <file>, then patch with --expect-hash <sha> or apply
with a hash-checked manifest. Create new files with create <file> --expect-absent
--stdin (and --parents when needed). Use hash-checked move and delete for those
operations. If a hash changes, re-read and reassess; never blindly retry stale edits.
Do not substitute shell redirection, sed, Python scripts or provider edit tools
for implementation reads or edits. Honor rgw-ast policy and use exec for generators
when required; do not disable hooks, change global policy or bypass a denial.
Progress updates, build.json and scratch payloads outside src/ may use ordinary
file writes. Build and test commands are still required. If rgw-ast cannot perform
a required operation, report the failure and stop rather than silently bypass it.

Produce actual implementation source plus tests of observable behavior, including error
cases required by the program. Build a real native executable, not a shell/Python/Node
launcher or a library. Leave reproducible commands for the driver to execute again.
Do not ask questions interactively. Do not write clarification diagnostics.
Write build.json in this directory with exactly this JSON structure (example uses Go;
you may choose another native language):
{
  "version": 1,
  "language": "Go",
  "summary": "What the program does",
  "assumptions": ["Decisions made to resolve unspecified or conflicting requirements"],
  "artifact": "bin/program",
  "build": [["go", "-C", "src", "build", "-o", "../bin/program", "."]],
  "test": [["go", "-C", "src", "test", "./..."]]
}
Commands are nonempty argv arrays run from this directory, with inherited environment
and GOWORK=off. Both build and test lists are required. The driver runs build then test,
validates the native executable and publishes it. The artifact must be a relative path
inside this directory, without symlinks. Avoid external services for tests unless the
source requires them. Do not merely print a proposed implementation: write the files.
End with a brief summary; the driver reads files, not your final response.

SOURCE BUNDLE:
%s
`, runtime.GOOS, runtime.GOARCH, p.Entry, sources)
}

# Security

Nutshell invokes an AI CLI with unattended, full-permission settings. The CLI and
generated build/test commands run with the current user's permissions. A private
`/tmp` directory organizes build work; it is not a sandbox. The driver does not
guarantee that an agent follows every instruction in its prompt.

Compile trusted programs in an environment appropriate for that access. A native
executable is not automatically safe because it passed generated tests. The
`rgw-ast` prompt requirement is not a substitute for process isolation.

Build attempts retain source snapshots and logs in `/tmp`. They can include
application data and provider output. Treat them as sensitive and inspect them
before attaching anything to a public issue.

## Reporting a vulnerability

Use GitHub's [private vulnerability reporting](https://github.com/ryangerardwilson/nutshell/security/advisories/new)
for security defects. Include the affected version, minimal reproduction, expected
boundary, and observed behavior. Do not include credentials or live private data.

Ordinary bugs and documentation corrections can use public issues. Nutshell is
pre-1.0 software; fixes target the current main branch, with no maintained older
release branches or guaranteed response time.

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

# Example programs

Each folder contains a `main.nut`. Compile from that folder after completing the
prerequisites in the [installation guide](../README.md#install):

```sh
cd examples/hello
nutshell main.nut -c codex -s ./src
./main
```

Choose another compiler with `-c grok`, `-c claude`, or a configured adapter.
Generated `src/` directories and default `main` binaries are ignored by Git.

| Program | Demonstrates |
| --- | --- |
| [hello](hello/main.nut) | A minimal complete program |
| [greet](greet/main.nut) | AI-interpreted file composition and command-line behavior |
| [features](features/main.nut) | Natural-language references across several `.nut` files |
| [sum](sum/main.nut) | Standard input, arithmetic, validation and failure examples |

These files are executable specifications for the AI, not checked-in generated
implementations. The provider chooses how to implement them. The Go test suite
tests the driver independently of live provider generation.

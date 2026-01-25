# Rosetta Code Corpus

This directory collects small, single-file Able programs that solve
popular [Rosetta Code](https://rosettacode.org/) tasks.  Each example is
written against the Able v11 specification and contains its own
`main` function so it can be executed directly with whichever Able
runtime you have available (for example, the TypeScript interpreter via
`bun run v11/interpreters/ts/src/cli.ts examples/rosettacode/<file>.able` or
the Go interpreter’s CLI once built).

## Starter Set — “Simplest Five”

| File | Rosetta Code task | Notes |
| --- | --- | --- |
| `hello_world.able` | [Hello World](https://rosettacode.org/wiki/Hello_world/Text) | Minimal Able program showing package declaration plus `main`. |
| `fizzbuzz.able` | [FizzBuzz](https://rosettacode.org/wiki/FizzBuzz) | Uses `if/or` chains and Able’s `for`/range syntax to print the classic 1–100 fizz/buzz sequence. |
| `fibonacci.able` | [Fibonacci sequence](https://rosettacode.org/wiki/Fibonacci_sequence) | Builds an `Array i64` iteratively and demonstrates helper functions that return collections. |
| `factorial.able` | [Factorial](https://rosettacode.org/wiki/Factorial#Procedural) | Implements both iterative and recursive factorial helpers to highlight pattern matching. |
| `bottles_of_beer.able` | [99 Bottles of Beer](https://rosettacode.org/wiki/99_Bottles_of_Beer) | Formats verses with simple branching and string interpolation. |

## Additional Examples

| File | Rosetta Code task | Notes |
| --- | --- | --- |
| `sieve_of_eratosthenes.able` | [Sieve of Eratosthenes](https://rosettacode.org/wiki/Sieve_of_Eratosthenes) | Implements the boolean sieve with nested `while` loops and produces primes ≤ *N*. |

## Contributing

When adding new Rosetta Code entries:

1. Follow the coding style from the existing examples (concise packages,
   generous use of helper functions).
2. Include a short comment header if a task needs extra context or
   references the spec.
3. Update the appropriate table above with the new file name, Rosetta Code link, and
   a one-line description of what the example demonstrates.

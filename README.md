# Gogen

A tool to generate golang program code based a definition language that similar
to [gherkin](https://cucumber.io/docs/gherkin/).

Example program definition:

```
Program: fizzbuzz

  Fizz buzz is a group word game for children to teach them about division.
  Players take turns to count incrementally, replacing any number divisible by
  three with the word "fizz", and any number divisible by five with the word
  "buzz".

  See https://en.wikipedia.org/wiki/Fizz_buzz
```

That will generate golang source code for a program named `fizzbuzz` with
following behavirors:

```sh
./fizzbuzz -h  # or -help

Fizz buzz is a group word game for children to teach them about division.
Players take turns to count incrementally, replacing any number divisible by
three with the word "fizz", and any number divisible by five with the word
"buzz".

See https://en.wikipedia.org/wiki/Fizz_buzz
```


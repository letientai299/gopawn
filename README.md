# Gogen

> Note: Experiment project

A tool to generate golang program code based a definition language that similar
to [gherkin](https://cucumber.io/docs/gherkin/).

Example program definition (see `testdata/fizzbuzz.prog`):

```
Program: fizzbuzz

Fizz buzz is a group word game for children to teach them about division.
Players take turns to count incrementally, replacing any number divisible by
three with the word "fizz", and any number divisible by five with the word
"buzz".

See https://en.wikipedia.org/wiki/Fizz_buzz
```

That will generate golang source code for a program named `fizzbuzz.go` with
following behavirors:

```sh
./fizzbuzz -h  # or -help

Fizz buzz is a group word game for children to teach them about division.
Players take turns to count incrementally, replacing any number divisible by
three with the word "fizz", and any number divisible by five with the word
"buzz".

See https://en.wikipedia.org/wiki/Fizz_buzz

  -help
        show usage
```

## Build and try it

Clone this project some where, and execute following commands:

```sh
$ make all # generate code and build gopawn
$ cd testdata
$ ../bin/gopawn fizzbuzz.prog # generate fizzbuzz program
$ cat fizzbuzz.go # view generated source
```

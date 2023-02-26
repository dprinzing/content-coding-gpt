# content-coding-gpt

This project demonstrates the use the OpenAI API for content-coding in psychology research.

## Learning Resources

* [Go Programming Language](https://go.dev/) Home Page
* [Go Standard Library](https://pkg.go.dev/std) package documentation
* [Learning Go](https://learning.oreilly.com/library/view/learning-go/9781492077206/), by Jon Bodner (highly
  recommended for learning modern, idiomatic Go)
* [OpenAI API Reference](https://platform.openai.com/docs/api-reference/introduction)

## Developer Workstation Setup

* Install the [Go Language](https://golang.org/doc/install), and set up
  a [GOPATH environment variable](https://github.com/golang/go/wiki/SettingGOPATH).
* Install an IDE, such as [VSCode](https://code.visualstudio.com/) with
  the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.go).
* Install some Go tools used in the [Makefile](Makefile) in your GOPATH bin folder:

```bash
make tools
```

* Configure the following environment variables for your OpenAI account. For
  simplicity, you could put them in your `.bashrc`, `.zshrc`, or equivalent.
  Substitute the placeholders below with your own values, of course.

```bash
export OPENAI_API_KEY="sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
export OPENAI_ORG_ID="org-xxxxxxxxxxxxxxxxxxxxxxxx"
```

* Build the applications for running in your local development environment:

```bash
make build
```

## Run the Command-line Application

You can explore the available commands, subcommands, and optional flags using
the `-h` or `--help` flags. For example:

```bash
./gpt -h
./gpt model -h
./gpt model list -h
./gpt model read -h
```

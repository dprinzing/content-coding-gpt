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

* Configure the following environment variables for your OpenAI account:

```bash
export OPENAI_API_KEY="sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
export OPENAI_ORG_ID="org-xxxxxxxxxxxxxxxxxxxxxxxx"
```

* Build the applications for running in your local development environment:

```bash
make build
```

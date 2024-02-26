# [WIP] make-help

Generate help message for Makefiles from from comments.

## Usage

Comments are used to generate the help message, for example:

```
## A comment line that begins with ## is considered to be a documentation comment that should be used in the help message
## @category example
.PHONY: example
example:
    echo "hello world"

## This help massage
## @category [shared] help
help:
    ./_bin/make-help
```

## How it works

The tool first runs `make -qpRr` to print all targets and variables, these are used to populate lookup maps. The `MAKEFILE_LIST` variable in `make` always contains a list of all the processed Makefiles, including "includes". We read this variable and scan all files listed for comments in the correct style (starting with `##`). If the comment precedes a make target, we add that to the internal target list.

The variables we have obtained are then used to expand the targets and usage comments and the help message is printed.
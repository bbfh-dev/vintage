# 📦️ Mime

Minecraft data-driven vanilla data & resource pack development kit powered by pre-processors and generators with minimum boilerplate and setup.

> Mime uses <q>simple by default, powerful when needed</q> philosophy.

> [!CAUTION]
> If you are using Windows (god bless your soul), the behavior of [executable inline templates](#1422-executable-inline-template) is undefined/untested.
>
> I have no plans of ever installing Windows to test or debug.

## Table of Contents

<!-- vim-markdown-toc GFM -->

* [Roadmap before v1](#roadmap-before-v1)
* [0 Why it exists](#0-why-it-exists)
* [1 Features](#1-features)
    * [1.1 Relative resource paths](#11-relative-resource-paths)
    * [1.2 Nested functions](#12-nested-functions)
    * [1.3 Mcmeta generation](#13-mcmeta-generation)
        * [1.3.1 Name](#131-name)
        * [1.3.2 Minecraft](#132-minecraft)
        * [1.3.3 Version](#133-version)
    * [1.4 Templates](#14-templates)
        * [1.4.1 Substitutions](#141-substitutions)
        * [1.4.2 Inline templates](#142-inline-templates)
            * [1.4.2.1 Simple inline template](#1421-simple-inline-template)
            * [1.4.2.2 Executable inline template](#1422-executable-inline-template)
            * [1.4.2.3 Invoking inline templates](#1423-invoking-inline-templates)
        * [1.4.3 Generator templates](#143-generator-templates)
            * [1.4.3.1 Definitions](#1431-definitions)
            * [1.4.3.2 Iterators](#1432-iterators)
* [2 CLI](#2-cli)
    * [2.1 Installation](#21-installation)
        * [2.1.1 Pre-built binaries](#211-pre-built-binaries)
        * [2.1.2 Using Go CLI](#212-using-go-cli)
    * [2.2 Usage](#22-usage)
        * [2.2.1 Init](#221-init)
        * [2.2.2 Main](#222-main)
* [3 Developer notes](#3-developer-notes)
    * [3.1 Generate for multiple versions](#31-generate-for-multiple-versions)
    * [3.2 Code validation](#32-code-validation)

<!-- vim-markdown-toc -->

# Roadmap before v1

- [ ] Add support for overlays.
- [ ] Add support for nested inline templates.
- [ ] Allow for whitespace in nested code.
- [ ] Add more tests.
- [ ] (Idea) automatically generate overlays / different packs for different versions[⁽¹⁾](#31-generate-for-multiple-versions)

# 0 Why it exists

There exist alternative long-established development kits such as [Beet](https://github.com/mcbeet/beet). **So why does this project exist?**

1. Mime **doesn't force you** to use a specific scripting language (e.g., Python or JavaScript);
2. Mime **is lightweight**. A project is just `pack.mcmeta` metadata, no environment setups are required.
3. Mime **is simple**. Any generated files are defined separately as [templates](#14-templates), functions still use `mcfunction` with the addition of [sugar-code](#1-features).
4. Mime **is a statically-linked binary**. That means it's portable and performant.

# 1 Features

## 1.1 Relative resource paths

Any mention of `./` will be replaced with the path to the current file as a resource.
`../` can be used to reference the parent (only once).

<table>
<tr><td>📁 Input</td></tr>
<tr><td>

```mcfunction
# File: data/example/function/load.mcfunction                          
function ./_my_nested_function
    # this would create an infinite loop
    function ../_my_nested_function
```

</td></tr>
</table>

<table>
<tr><td>📦️ Output</td></tr>
<tr><td>

```mcfunction
# File: data/example/function/load.mcfunction                          
function example:load/_my_nested_function
```

</td></tr>
</table>

## 1.2 Nested functions

Create nested functions by adding indentation (must be a tab or 4 spaces) to code, subsequent to a function call.

<table>
<tr><td>📁 Input</td></tr>
<tr><td>

```mcfunction
# File: data/example/function/test.mcfunction                          
execute as @e run function ./_my_nested_function
    say 123
    kill @s
```

</td></tr>
</table>

<table>
<tr><td>📦️ Output (1/2)</td></tr>
<tr><td>

```mcfunction
# File: data/example/function/test.mcfunction                          
execute as @e run function example:test/_my_nested_function
```

</td></tr>
</table>

<table>
<tr><td>📦️ Output (2/2)</td></tr>
<tr><td>

```mcfunction
# File: data/example/function/test/_my_nested_function.mcfunction      
say 123
kill @s
```

</td></tr>
</table>

You are also technically able to define functions from the same file without calling them by using comments:

<table>
<tr><td>📁 Input</td></tr>
<tr><td>

```mcfunction
# function example:any/function/here1
    say 123
    kill @s

# function example:any/function/here2
    say 123
    kill @s
```

</td></tr>
</table>

## 1.3 Mcmeta generation

`pack.mcmeta` files for both data & resource packs are automatically generated with the appropriate format based on supported Minecraft versions.

<table>
<tr><td>📁 Input</td></tr>
<tr><td>

```json
{                                                 
    "pack": {
        "description": "This is my pack"
    },
    "meta": {
        "name": "example",
        "minecraft": {
            "min": "1.20",
            "max": "1.21.1"
        },
        "version": "0.1.0-alpha"
    }
}
```

</td></tr>
</table>

<table>
<tr><td>📦️ Output</td></tr>
<tr><td>

```json
{                                                 
    "pack": {
        "description": "This is my pack",
        "pack_format": 48,
        "min_format": [48, 0],
        "max_format": [94, 1]
    },
    "meta": {
        "name": "example",
        "minecraft": {
            "min": "1.20",
            "max": "1.21.1"
        },
        "version": "0.1.0-alpha"
    }
}
```

</td></tr>
</table>

### 1.3.1 Name

Field `meta.name` must be a `string`.
It is used to generate `.zip` files[⁽²⁾](#22-usage).

### 1.3.2 Minecraft

Field `meta.minecraft` must be one of:
- a `string` with the Minecraft version;
- an `object` containing `min` and `max` fields with the minimum and maximum Minecraft versions, respectively;

### 1.3.3 Version

The project's version. It is recommended to use [semantic versioning](https://semver.org/), it is not enforced.
It is used to generate `.zip` files[⁽²⁾](#22-usage).

## 1.4 Templates

Templates enable data-driven generation of data & resource pack files.

They are defined using a `templates/<template_name>/manifest.json` file at the root of the project.

See [Examples / 02_templates](./examples/02_templates) for examples.

### 1.4.1 Substitutions

Template file contents and file/directory names can be formatted with variables by using `%[<name>]`.

Depending on the context, `"%[<name>]"` inside of `.json` files can be substituted with any value, not just a string.

- Iterators can contain an index (defaults to 0) when necessary e.g. `%[<iterator>.1]` or `%[<iterator>.2]`.
- Variables can contain modifiers e.g. `[<name>.<modifier>]`:
    - `to_file_name`: limits the value to `[a-z+_]` charset (whitespace gets replaced with an `_`).
    - `to_lower_case`: converts the string to all lower case.
    - `to_upper_case`: converts the string to all upper case.
    - `length`: converts to the length of the string.

```jsonc
{
    "key": "%[value]"
}

// Could become:
{
    "key": "my_value"
}

// Or, depending on the input:
{
    "key": {
        "literally": "anything"
    }
}
```

> [!NOTE]
> An `%[id]` is always provided as the unique identifier of the `definitions/` file (with all iterators substituted).
>
> Usually, all files must contain this in their path to prevent clashing (when different definitions try to create the same file).

### 1.4.2 Inline templates

Allow you to add new syntax to `mcfunction` files.

```jsonc
// manifest.json
{
    "$schema": "https://bbfh.me/mime/manifest/schema.json",
    "type": "inline",
    // (Example) requires 3 positional arguments.
    "arguments": [
        "arg1",
        "arg2",
        "arg3"
    ],
    // (Example) don't define "arguments" or set it to "null"
    // if you want any number of input arguments
    // NOTE: all arguments are provided as a single string argument
    "arguments": null
}
```

The contents of the template can be one of the following:

#### 1.4.2.1 Simple inline template

Create a `body.mcfunction` file. Its contents will be inserted in place of the invocation.

Use `%[...]` to place any nested code.

Example:
```mcfunction
# for/body.mcfunction
# this is an example '#!/for @s my_objective ..5' for-loop implementation
scoreboard players set %[target] %[objective] 0
function ./_for_each_%[target.to_file_name]
	%[...]
	scoreboard players add %[target] %[objective] 1
	execute if score %[target] %[objective] matches %[range] run function ../_for_each_%[target.to_file_name]
```

#### 1.4.2.2 Executable inline template

Create a `call` file; it can optionally have any extension (e.g., `call.py` or `call.sh` is valid) as long as it is an executable.

Example:
```sh
#!/bin/sh

echo "$*"
cat # this will print all nested code right after
```

#### 1.4.2.3 Invoking inline templates

While inside of `mcfunction` files, you can use `#!/<template_name> <args...>` syntax. The inline template will be written in place of the line.

Any nested code will be fed to the template.

```mcfunction
#!/for @s my_objective ..5
    say 123
    say abc
```

### 1.4.3 Generator templates

As the name suggests, they generate files from a user-defined list of items.

```jsonc
// manifest.json
{
    "$schema": "https://bbfh.me/mime/manifest/schema.json",
    "type": "generator",
    // (Example) define 2 iterators
    "iterators": {
        "material": [
            [
                "acacia",
                "acacia_planks"
            ],
            [
                "oak",
                "oak_planks"
            ],
            [
                "spruce",
                "spruce_planks"
            ],
            [
                "stone",
                "stone_bricks"
            ]
        ],
        "color": [
            "red",
            "green",
            "blue"
        ]
    }
}
```

Create your regular minecraft files inside the `data/` & `assets/` directories, which will be merged with the data/resource pack. Note, that paths should use substitutions preferably with `%[id]` to avoid clashing.

#### 1.4.3.1 Definitions

Define `definitions/<name>.json` files, optionally using previously defined iterators. The contents off the files will be provided to every data/resource pack file for substitution.

```jsonc
// Example, you can put anything in here:
{
    "block": "%[material.1]",
    "recipe": [
        {
            "id": 1,
            "name": "some item here"
        }
    ]
}
```

#### 1.4.3.2 Iterators

Iterators allow you to generate definitions for every unique combination of values.

Example:
```bash
# File:
%[color]_%[material]_chair.json

# Would generate (assuming you are using the iterators from the example above):
red_acacia_chair.json
green_acacia_chair.json
blue_acacia_chair.json
red_oak_chair.json
green_oak_chair.json
# etc...
```

The definition file will be substituted with the current iterator value.

# 2 CLI

## 2.1 Installation

### 2.1.1 Pre-built binaries

1. Download the [latest release](https://github.com/bbfh-dev/mime/releases/latest) for your OS and architecture.
2. Put it into a directory listed in your `$PATH`;

### 2.1.2 Using Go CLI

1. Assuming you have [Go](https://go.dev/) installed;
2. Run `go install github.com/bbfh-dev/mime@latest`

## 2.2 Usage

### 2.2.1 Init

This utility is used to initialize a new Mime project or convert an existing data/resource pack into the appropriate format.

Under the hood, it simply writes fields to `pack.mcmeta`, as this is the only requirement for Mime.

```
Initialize a new Mime project

[?] Usage:
    init [options...] <work-dir?>

[#] Options:
    --help
        # Print this help message and exit
    --name, -n <string> (default: untitled)
        # Specify the project name that will be used for exporting
    --minecraft, -m <string> (default: 1.21.11)
        # Specify the target Minecraft version. Use '-' to indicate version ranges, e.g. '1.20-1.21'
    --version, -v <string> (default: 0.1.0-alpha)
        # Specify the project version using semantic versioning
    --description, -d <string>
        # Specify the project description
```

Example usage:
```bash
# Note that all of the flags are optional
$ mime init ./examples/01_basic --name=untitled --pack-version=1.0.0 --minecraft=1.21.11 --description "Hello World!"
```

### 2.2.2 Main

```
Minecraft data-driven vanilla data & resource pack development kit powered by pre-processors and generators

[?] Usage:
    mime [options...] <work-dir?>

[>] Commands:
    init
        # Initialize a new Mime project

[#] Options:
    --help
        # Print this help message and exit
    --version
        # Print program version and exit
    --output, -o <string> (default: ./build)
        # Output directory relative to the pack working dir
    --zip, -z
        # Export data & resource packs as .zip files
    --debug, -d
        # Print verbose debug information
    --force, -f
        # Force build even if the project was cached
```

Example usage:
```bash
$ mime -o /tmp/mime-build --zip --debug ./examples/02_templates
```

# 3 Developer notes

- [ ] Refactor template code, because it's scattered around internal, language and mime.
- [ ] Add support for `NO_COLOR` and `TERM=dumb` environment variables. Use `"golang.org/x/term" term.IsTerminal(int(os.Stdout.Fd()))`

## 3.1 Generate for multiple versions

The idea is to automatically create overlays when possible, separate packs when necessary from a single codebase.

Since minecraft packs keep changing in ways that are hard to keep up with, making this system completely autonomous is too much work.

For this reason, I propose a system that is similar to how Go handles code that should compile differently depending on the target.

Example:
```bash
data/example/loot_table/
    - dir[1.13] # directory for 1.13 only
    - file.json # default version
    - file[1.20].json # version for 1.20 only
    - file[1.19.4].json # version for 1.19.4 only
```

## 3.2 Code validation

It would be neat if code could be validated for all target Minecraft versions. However, this task is way too much work for not enough to be gained.

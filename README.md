# Mime

Minecraft data & resource pack processor designed to be a useful tool for vanilla development rather than a new scripting language or ecosystem.

## Table of Contents

<!-- vim-markdown-toc GFM -->

* [1 Features](#1-features)
    * [1.1 Nested functions](#11-nested-functions)
    * [1.2 Mcmeta generation](#12-mcmeta-generation)
    * [1.3 Add-ons](#13-add-ons)
        * [1.3.1 Defining an add-on](#131-defining-an-add-on)
        * [1.3.2 Syntax](#132-syntax)
        * [1.3.3 Definitions file](#133-definitions-file)

<!-- vim-markdown-toc -->

# 1 Features

## 1.1 Nested functions

Create nested functions by adding indentation (must be a tab or 4 spaces) to code subsequent to a function call.

Resource location can use `./` for relative paths.

```mcfunction
# example:test
execute as @e run function ./_my_nested_function
    say 123
    kill @s
```

Would create the following files:

```mcfunction
# example:test
execute as @e run function example:test/_my_nested_function
```

```mcfunction
# example:test/_my_nested_function
say 123
kill @s
```

## 1.2 Mcmeta generation

`pack.mcmeta` files for both data & resource packs are automatically generated with the appropriate format based on supproted Minecraft versions.

```jsonc
// Local pack.mcmeta
{
    "pack": {
        "description": "This is my pack"
    },
    "meta": {
        "name": "example",
        "minecraft": "1.20-1.21.11",
        "version": "0.1.0-alpha"
    }
}
```

Generates the following data pack meta

```jsonc
// build/data_pack/pack.mcmeta
{
    "pack": {
        "description": "This is my pack",
        "pack_format": 48,
        "min_format": [48, 0],
        "max_format": [94, 1]
    },
    "meta": {
        "name": "example",
        "minecraft": "1.21-1.21.11",
        "version": "0.1.0-alpha"
    }
}
```

## 1.3 Add-ons

Generate data & resource pack files based on templates.

### 1.3.1 Defining an add-on

All add-ons are defined inside of the `<project_root>/addons/` directory, where `<project_root>` is the directory with `assets/`, `data/`, `pack.mcmeta`, etc.

An add-on is a directory that has the following structure:
```bash
./addons/my_example_addon/
├── data/
│   └── # (Optional) Template files for the data pack
├── assets/
│   └── # (Optional) Template files for the resource pack
└── definitions.json # Main file
```

### 1.3.2 Syntax

Whether be in a file or directory name, a JSON key or value, or an mcfunction command you can use the following syntax:

`%[<JSON selector>]` to insert a value from the current definition item (e.g. `%[id]` would insert `%[bubble_bench]` in the first iterator of the example below).

JSON file keys also have a special syntax `%-><JSON key>` that **expands** the value merging it with any existing key of that name (e.g. `"%->Tags": "%[tags]"` would append all elements of `tags` into `Tags` from the example below).

### 1.3.3 Definitions file

The JSON schema can be found here: `./mime/schemas/definitions.json`.

Example:

```jsonc
{
    "$schema": "(...)/definitions.json",
    "iterators": {
        "material": [
            "oak",
            "spruce"
        ],
        "color": [
            "red",
            "white",
            "yellow"
        ]
    },
    "definitions": [
        {
            // All of these fields are made up for this specific project.
            // Put whatever you want here, just make sure that
            // all keys are defined on ALL of the objects inside .definitions
            "id": "bubble_bench",
            "category": "furniture",
            "base": "barrel[facing=up]",
            "sound": "industrial",
            "facing": "player",
            "recipe": [
                {
                    "group": "block",
                    "id": "crafting_table",
                    "count": 1
                },
                {
                    "group": "item",
                    "id": "glass_bottle",
                    "count": 1
                }
            ],
            "material": {
                "index": 0,
                "name": "default"
            },
            "material_index": 0,
            "tags": [
                "--bbln.uses.gui",
                "--bbln.uses.brightness_fix"
            ]
        },
        {
            // You can use iterators like this
            // an iterator is just a for-each loop
            // that exposes %[<iterator_name>] for the current `array[i]` value
            // and %[<iterator_name>:index] for the current `i` value
            "iterate": "material",
            "return": {
                "category": "furniture",
                "id": "%[material]_table",
                "base": "structure_void",
                "sound": "wooden",
                "facing": "player",
                "recipe": [
                    {
                        "group": "block",
                        "id": "%[material]_planks",
                        "count": 4
                    }
                ],
                "material": {
                    "index": "%[material:index]",
                    "name": "%[material]"
                },
                "tags": []
            }
        }
    ]
}
```

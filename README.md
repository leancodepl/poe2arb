<div align="center">

# poe2arb

[![Latest GitHub release][github-release-img]][github-release-link]
[![CI status(main branch)][ci-status-img]][ci-status-link]

![][screenshot-img]
</div>

`poe2arb` is a CLI tool that lets the POEditor work with Flutter's native
localization solution (`flutter gen-l10n`).

## Installation

You can download latest or historical binary straight from the [GitHub
releases][releases] artifacts or using Homebrew:

```
brew tap leancodepl/poe2arb
brew install poe2arb
```

## Usage

`poe2arb` operates on POEditor's _JSON_ (not _JSON key-value_) export file
format.

### Full POEditor integration

`poe2arb poe` command is your Swiss Army Knife enabling integrating POEditor
into your Flutter workspace in one command:

1. Fetches all project languages from API.
2. Downloads JSON exports for all languages from API.
3. Converts JSON exports to ARB format.
4. Saves converted ARB files to the output directory.

#### Options

| Flag               | Description                                                                                                       | Default                                                                                                             |
|--------------------|-------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------|
| `-p, --project-id` | **Required.** POEditor project ID. It is visible in the URL of the project on POEditor website.                   | Defaults to `poeditor-project-id` from `l10n.yaml` file. If it's empty, then defaults to `POEDITOR_PROJECT_ID` env. |
| `-t, --token`      | **Required.** POEditor API read-only access token. Available in [Account settings > API access][poeditor-tokens]. | Defaults to `POEDITOR_TOKEN` env.                                                                                   |
| `-o, --output-dir` | ARB files output directory.                                                                                       | Defaults to `arb-dir` from `l10n.yaml`. If it's empty, then defaults to `.`.                                        |
| `--langs`          | Exported languages override.                                                                                      | Defaults to empty, no override.                                                                                     |

### Conversion

`poe2arb convert` command only converts the POE export to ARB format. Refer to
[Supported features](#syntax--supported-features) section.

For conversion, you need to pass the translation file language in the
`--lang/-l` flag.

By default, a template ARB file is generated. So no empty message is skipped and attributes are generated. If you want to skip that, pass `--no-template` flag.

Currently, only an stdin/stdout is supported for the `poe2arb convert` command.

```
poe2arb convert io --lang en < Hello_World_English.json > lib/l10n/app_en.arb
```

## Syntax & supported features

Term name must be a valid Dart field name, additionaly, it must start with a
lowercase letter ([Flutter's constraint][term-name-constraint]).

### Placeholders

Placeholders can be as simple as a text between brackets, but they can also be
well-defined with a type and format, to make use of date and number formatting.

By default, simple `{placeholder}` will have an `Object` type and will be `toString()`ed.

Each unique placeholder must be defined only once. I.e. for one `{placeholder,String}` you may have many
`{placeholder}` (that use the same definition), but no other `{placeholder,String}` must be found in the term.

Placeholders with type `DateTime` must have a format specified. The valid values are the names of
[the `DateFormat` constructors][dateformat-constructors], e.g. `yMd`, `jms`, or `EEEEE`.

Placeholders with type `num`, `int`, or `double` **may have\*** a format specified. The valid values are the names
of [the `NumberFormat` constructors][numberformat-constructors], e.g. `decimalPattern`, or `percentPattern`.
In plurals, the `count` placeholder must be of `int` or `num` type. It can be left with no definition.

Number placeholders without a specified format will be simply `toString()`ed.

**Only template files can define placeholders with their type and format.** In non-template languages, placeholders' types and formats
are ignored and no logical errors are reported.

> \*If you're using Flutter 3.5 or older, you need to specify format for numeric placeholders.
> Otherwise `flutter gen-l10n` will fail. You can look at the legacy placeholder syntax diagrams
> [for placeholders here][flutter35-placeholders-diagram] and for [plural's `count` placeholders here][flutter35-count-placeholders-diagram].

#### Examples

Below are some examples of strings that make use of placeholders. Simple and well-defined.

```
Hello, {name}!
```

```
Hello, {name,String}!
```

```
You have {coins,int,decimalPattern} coins left in your {wallet,String} wallet.
```

```
last modified on {date,DateTime,yMMMEEEEd}
```

#### Placeholder syntax diagram

![][placeholder-diagram-img]

##### `count` placeholder syntax diagram

![][count-placeholder-diagram-img]


### Plurals

POEditor plurals are also supported. Simply mark the the term as plural and
give it _any_ name (it's never used, but required by POEditor to enable plurals
for the term).

In translations, a `{count}` placeholder can be used. You can use other placeholders too. Example:

```
one:    Andy has 1 kilogram of {fruit}.
other:  Andy has {count} kilograms of {fruit}.
```

You must provide at least `other` plural category for your translations, otherwise it won't be converted.

## Contributing

### Formatting

We use [gofumpt][gofumpt], which is a superset of [gofmt][gofmt].

To make `gopls` in VS Code use `gofumpt`, add this to your settings:

```json
"gopls": {
    "formatting.gofumpt": true
},
```

### Linting

We use [staticcheck][staticcheck] with all checks enabled.

To make VS Code use `staticcheck`, add this to your settings:

```json
"go.lintTool": "staticcheck",
"go.lintFlags": ["-checks=all"],
```

### Building

All you need is Go 1.20.

```
go build .
```

### Releasing

Create a _lightweight_ git tag and push it. GitHub Actions with a GoReleaser
workflow will take care of the rest.

```
git tag v0.1.1
git push origin v0.1.1
```

[github-release-link]: https://github.com/leancodepl/poe2arb/releases
[github-release-img]: https://img.shields.io/github/v/release/leancodepl/poe2arb?label=version&sort=semver
[ci-status-link]: https://github.com/leancodepl/poe2arb/actions/workflows/go-test.yml
[ci-status-img]: https://img.shields.io/github/actions/workflow/status/leancodepl/poe2arb/go-test.yml?branch=main
[screenshot-img]: art/terminal-screenshot.png
[releases]: https://github.com/leancodepl/poe2arb/releases
[poeditor-tokens]: https://poeditor.com/account/api
[term-name-constraint]: https://github.com/flutter/flutter/blob/ce318b7b539e228b806f81b3fa7b33793c2a2685/packages/flutter_tools/lib/src/localizations/gen_l10n.dart#L868-L886
[dateformat-constructors]: https://pub.dev/documentation/intl/latest/intl/DateFormat-class.html#constructors
[numberformat-constructors]: https://pub.dev/documentation/intl/latest/intl/NumberFormat-class.html#constructors
[flutter35-placeholders-diagram]: https://github.com/leancodepl/poe2arb/blob/24be17d6721698526c879b3fada87183b359e8e8/art/placeholder-syntax.svg
[flutter35-count-placeholders-diagram]: https://github.com/leancodepl/poe2arb/blob/24be17d6721698526c879b3fada87183b359e8e8/art/count-placeholder-syntax.svg
[placeholder-diagram-img]: art/placeholder-syntax.svg
[count-placeholder-diagram-img]: art/count-placeholder-syntax.svg
[gofumpt]: https://github.com/mvdan/gofumpt
[gofmt]: https://pkg.go.dev/cmd/gofmt
[staticcheck]: https://staticcheck.io

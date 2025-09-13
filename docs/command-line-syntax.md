# Our command line syntax

## Literal text

Use plain text for parts of the command that cannot be changed.

_example:_
`zen help`
The argument help is required in this command.

## Placeholder values

Use angled brackets to represent a value the user must replace. No other expressions can be contained within the angled brackets.

_example:_
`zen task view <task-number>`
Replace `<task-number>` with an task number.

## Optional arguments

Place optional arguments in square brackets. Mutually exclusive arguments can be included inside square brackets if they are separated with vertical bars.

_example:_
`zen task checkout [--web]`
The argument `--web` is optional.

`zen task view [<number> | <url>]`
The `<number>` and `<url>` arguments are optional.

## Required mutually exclusive arguments

Place required mutually exclusive arguments inside braces, separate arguments with vertical bars.

_example:_
`zen task {view | create}`

## Repeatable arguments

Ellipsis represent arguments that can appear multiple times.

_example:_
`zen task close <task-number>...`

## Variable naming

For multi-word variables use dash-case (all lower case with words separated by dashes)

_example:_
`zen task checkout <task-number>`

## Additional examples

_optional argument with placeholder:_
`command sub-command [<arg>]`

_required argument with mutually exclusive options:_
`command sub-command {<path> | <string> | literal}`

_optional argument with mutually exclusive options:_
`command sub-command [<path> | <string>]`

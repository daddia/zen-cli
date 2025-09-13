# Foundations

Design concepts and constraints that can help create a better Terminal like experience for zen.

## Language

Language is the most important tool at our disposal for creating a clear, understandable product. Having clear language helps us create memorable commands that are clear in what they will do.

We generally follow this structure:

// TODO - Table to be updated

| **zen**  | **`<command>`** | **`<subcommand>`** | **[value]** | **[flags]** | **[value]** |
| -------- | --------------- | -------------------| ----------- | --------- | ------- |
| zen      | task            | view               | ABC-234     | --web     | -       |
| zen      | task            | create             | ABC-789     | --title   | “Title” |
| zen      | code            | generate           | ABC-234     | --title   | “Title” |
| zen      | code            | review             | 234         | --approve | -       |
| zen      | test            | status             | -           | -         | -       |
| zen      | task            | list               | -       | --state   | closed  |

**Command:** The object you want to interact with

**Subcommand:** The action you want to take on that object. Most `zen` commands contain a command and subcommand. These may take arguments, such as issue/PR numbers, URLs, file names, OWNER/REPO, etc.

**Flag:** A way to modify the command, also may be called “options”. You can use multiple flags. Flags can take values, but don’t always. Flags always have a long version with two dashes `(--state)` but often also have a shortcut with one dash and one letter `(-s)`. It’s possible to chain shorthand flags: `-sfv` is the same as `-s -f -v`

**Values:** Are passed to the commands or flags

- The most common command values are:
  - Issue or PR number
  - The “owner/repo” pair
  - URLs
  - Branch names
  - File names
- The possible flag values depend on the flag:
  - `--state` takes `{closed | open | merged}`
  - `--clone` is a boolean flag
  - `--title` takes a string
  - `--limit` takes an integer

_Tip: To get a better sense of what feels right, try writing out the commands in the CLI a few different ways._

<table>
  <tr>
    <td>
      Do: Use a flag for modifiers of actions.
      <img alt="`zen task review --approve` command" src="images/Language-06.png" />
    </td>
    <td>
      Don't: Avoid making modifiers their own commands.
      <img alt="`zen task approve` command" src="images/Language-03.png" />
    </td>
  </tr>
</table>

**When designing your command’s language system:**

- Use [Zen language](/getting-started/principles#make-it-feel-zen)
- Use unambiguous language that can’t be confused for something else
- Use shorter phrases if possible and appropriate

<table>
  <tr>
    <td>
      Do: Use language that can't be misconstrued.
      <img alt="`zen task create` command" src="images/Language-05.png" />
    </td>
    <td>
      Don't: Avoid language that can be interpreted in multiple ways ("open in browser"  or "open a pull request" here).
      <img alt="`zen task open` command" src="images/Language-02.png" />
    </td>
  </tr>
</table>

<table>
  <tr>
    <td>
      Do: Use understood shorthands to save characters to type.
      <img alt="`zen repo view` command" src="images/Language-04.png" />
    </td>
    <td>
      Don't: Avoid long words in commands if there's a reasonable alternative.
      <img alt="`zen repository view` command" src="images/Language-01.png" />
    </td>
  </tr>
</table>

## Typography

Everything in a command line interface is text, so type hierarchy is important. All type is the same size and font, but you can still create type hierarchy using font weight and space.

![An example of normal weight, and bold weight. Italics is striked through since it's not used.](images/Typography.png)

- People customize their fonts, but you can assume it will be a monospace
- Monospace fonts inherently create visual order
- Fonts may have variable unicode support

### Accessibility

If you want to ensure that a screen reader will read a pause, you can use a:
- period (`.`)
- comma (`,`)
- colon (`:`)

## Spacing

You can use the following to create hierarchy and visual rhythm:

- Line breaks
- Tables
- Indentation

Do: Use space to create more legible output.

<img alt="`zen task status` command indenting content under sections" src="images/Spacing-zen-pr-status.png" />

Don't: Not using space makes output difficult to parse.

<img alt="`zen task status` command where content is not indented, making it harder to read" src="images/Spacing-zen-pr-status-compressed.png" />

## Color

Terminals reliably recognize the 8 basic ANSI colors. There are also bright versions of each of these colors that you can use, but less reliably.

<img alt="A table describing the usage of the 8 basic colors." src="images/Colors.png" />

### Things to note
- Background color is available but we haven’t taken advantage of it yet.
- Some terminals do not reliably support 256-color escape sequences.
- Users can customize how their terminal displays the 8 basic colors, but that’s opt-in (for example, the user knows they’re making their greens not green).
- Only use color to [enhance meaning](https://primer.style/design/accessibility/guidelines#use-of-color), not to communicate meaning.

## Iconography

Since graphical image support in terminal emulators is unreliable, we rely on Unicode for iconography. When applying iconography consider:

- People use different fonts that will have varying Unicode support
- Only use iconography to [enhance meaning](https://primer.style/design/global/accessibility#visual-accessibility), not to communicate meaning

_Note: In Windows, Powershell’s default font (Lucida Console) has poor Unicode support. Microsoft suggests changing it for more Unicode support._

**Symbols currently used:**

```sh
✓ 	Success
- 	Neutral
✗ 	Failure
+ 	Changes requested
! 	Alert
```

<table>
  <tr>
    <td>
      Do: Use checks for success messages.
      <img alt="✓ Checks passing" src="images/Iconography-1.png" />
    </td>
    <td>
      Don't: Don't use checks for failure messages.
      <img alt="✓ Checks failing" src="images/Iconography-2.png" />
    </td>
  </tr>
</table>

<table>
  <tr>
    <td>
      Do: Use checks for success of closing or deleting.
      <img alt="✓ Issue closed" src="images/Iconography-3.png" />
    </td>
    <td>
      Do: Don't use alerts when closing or deleting.
      <img alt="! Issue closed" src="images/Iconography-4.png" />
    </td>
  </tr>
</table>

## Scriptability

Make choices that ensure that creating automations or scripts with Zen commands is obvious and frictionless. Practically, this means:

- Create flags for anything interactive
- Ensure flags have clear language and defaults
- Consider what should be different for terminal vs machine output

### In terminal

![An example of zen task list](images/Scriptability-zen-pr-list.png)

### Through pipe

![An example of zen task list piped through the cat command](images/Scriptability-zen-pr-list-machine.png)

### Differences to note in machine output

- No color or styling
- State is explicitly written, not implied from color
- Tabs between columns instead of table layout, since `cut` uses tabs as a delimiter
- No truncation
- Exact date format
- No header

## Customizability

Be aware that people exist in different environments and may customize their setups. Customizations include:

- **Shell:** shell prompt, shell aliases, PATH and other environment variables, tab-completion behavior
- **Terminal:** font, color scheme, and keyboard shortcuts
- **Operating system**: language input options, accessibility settings

The CLI tool itself is also customizable. These are all tools at your disposal when designing new commands.

- Aliasing: [`zen alias set`](https://zen.daddia.com/docs/gh_alias_set)
- Preferences: [`zen config set`](https://zen.daddia.com/docs/gh_config_set)
- Environment variables: `NO_COLOR`, `EDITOR`, etc

// Package main implements the documentation generator for Zen CLI
// This tool generates Markdown, Man page, and ReStructuredText documentation
// from the Cobra command definitions, ensuring docs stay in sync with code.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/daddia/zen/pkg/cmd/root"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func main() {
	// Command-line flags
	out := flag.String("out", "./docs/zen", "output directory for generated documentation")
	format := flag.String("format", "markdown", "output format: markdown|man|rest")
	front := flag.Bool("frontmatter", false, "prepend YAML front matter to markdown files")
	timestamp := flag.Bool("timestamp", false, "include generation timestamp in files")
	flag.Parse()

	// Ensure output directory exists
	if err := os.MkdirAll(*out, 0755); err != nil {
		log.Fatalf("failed to create output directory: %v", err)
	}

	// Get the root command
	rootCmd, err := root.Root()
	if err != nil {
		log.Fatalf("failed to get root command: %v", err)
	}

	// Disable auto-generated tag for stable, reproducible files (unless timestamp requested)
	if !*timestamp {
		rootCmd.DisableAutoGenTag = true
	}

	// Enhanced command preparation for better LLM readability
	prepareCommands(rootCmd)

	// Generate documentation based on format
	switch *format {
	case "markdown":
		if err := generateMarkdown(rootCmd, *out, *front); err != nil {
			log.Fatalf("failed to generate markdown: %v", err)
		}
		// Generate index file for markdown format
		if err := generateIndex(rootCmd, *out, *front); err != nil {
			log.Fatalf("failed to generate index: %v", err)
		}
		log.Printf("✓ Generated Markdown documentation with index in %s", *out)

	case "man":
		if err := generateMan(rootCmd, *out); err != nil {
			log.Fatalf("failed to generate man pages: %v", err)
		}
		log.Printf("✓ Generated Man pages in %s", *out)

	case "rest":
		if err := generateReST(rootCmd, *out); err != nil {
			log.Fatalf("failed to generate ReStructuredText: %v", err)
		}
		log.Printf("✓ Generated ReStructuredText documentation in %s", *out)

	default:
		log.Fatalf("unknown format: %s (must be markdown, man, or rest)", *format)
	}
}

// prepareCommands enhances command definitions for better documentation
func prepareCommands(cmd *cobra.Command) {
	// Walk the command tree and enhance each command
	for _, c := range cmd.Commands() {
		// Ensure examples are present and well-formatted
		if c.Example == "" {
			c.Example = generateExampleForCommand(c)
		}

		// Add more context to Long descriptions if they're missing
		if c.Long == "" && c.Short != "" {
			c.Long = c.Short
		}

		// Recursively prepare subcommands
		prepareCommands(c)
	}
}

// generateExampleForCommand creates example usage for commands that lack them
func generateExampleForCommand(cmd *cobra.Command) string {
	if cmd.HasSubCommands() {
		// For parent commands, show subcommand usage
		var examples []string
		for _, sub := range cmd.Commands() {
			if !sub.Hidden {
				examples = append(examples, fmt.Sprintf("  zen %s %s", cmd.Name(), sub.Name()))
			}
		}
		if len(examples) > 0 {
			return strings.Join(examples[:min(3, len(examples))], "\n")
		}
	}

	// For leaf commands, generate basic example
	return fmt.Sprintf("  zen %s", cmd.CommandPath())
}

// generateMarkdown generates Markdown documentation
func generateMarkdown(rootCmd *cobra.Command, outDir string, withFrontMatter bool) error {
	if withFrontMatter {
		// Custom prepender for front matter
		prepender := func(filename string) string {
			base := filepath.Base(filename)
			name := strings.TrimSuffix(base, filepath.Ext(base))
			title := strings.ReplaceAll(name, "_", " ")
			slug := strings.ReplaceAll(strings.ToLower(name), "_", "-")

			// Generate rich front matter for static site generators
			return fmt.Sprintf(`---
title: %q
slug: "/cli/%s"
description: "CLI reference for %s"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: %s
keywords:
  - zen
  - cli
  - %s
---

`,
				title,
				slug,
				title,
				time.Now().UTC().Format("2006-01-02"),
				strings.ReplaceAll(strings.ToLower(title), " ", "-"),
			)
		}

		// Custom link handler for better cross-references
		linkHandler := func(name string) string {
			return strings.ToLower(strings.ReplaceAll(name, "_", "-")) + ".md"
		}

		return doc.GenMarkdownTreeCustom(rootCmd, outDir, prepender, linkHandler)
	}

	// Generate standard markdown without front matter
	return doc.GenMarkdownTree(rootCmd, outDir)
}

// generateMan generates Man page documentation
func generateMan(rootCmd *cobra.Command, outDir string) error {
	header := &doc.GenManHeader{
		Title:   strings.ToUpper(rootCmd.Name()),
		Section: "1",
		Date:    &time.Time{},
		Source:  "Zen CLI",
		Manual:  "Zen CLI Manual",
	}

	return doc.GenManTree(rootCmd, header, outDir)
}

// generateReST generates ReStructuredText documentation
func generateReST(rootCmd *cobra.Command, outDir string) error {
	return doc.GenReSTTree(rootCmd, outDir)
}

// generateIndex creates an index.md file listing all commands
func generateIndex(rootCmd *cobra.Command, outDir string, withFrontMatter bool) error {
	indexPath := filepath.Join(outDir, "index.md")
	file, err := os.Create(indexPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write front matter if requested
	if withFrontMatter {
		fmt.Fprintf(file, `---
title: "Zen CLI Reference"
slug: "/cli"
description: "Command-line reference documentation for Zen CLI"
section: "CLI Reference"
date: %s
keywords:
  - zen
  - cli
  - reference
  - documentation
---

`, time.Now().UTC().Format("2006-01-02"))
	}

	// Write header
	fmt.Fprintln(file, "# Zen CLI Reference")
	fmt.Fprintln(file)
	fmt.Fprintln(file, "Command-line reference documentation for Zen CLI.")
	fmt.Fprintln(file)

	// Write main command section
	fmt.Fprintln(file, "## Main Command")
	fmt.Fprintln(file)
	fmt.Fprintf(file, "### [zen](zen.md)\n")
	fmt.Fprintf(file, "%s\n", rootCmd.Short)
	fmt.Fprintln(file)

	// Organize commands by group
	coreCommands := []string{}
	futureCommands := []string{}

	for _, cmd := range rootCmd.Commands() {
		if cmd.Hidden {
			continue
		}

		// Skip completion command from index
		if cmd.Name() == "completion" {
			continue
		}

		// Categorize commands
		switch cmd.Name() {
		case "init", "config", "status", "version":
			coreCommands = append(coreCommands, cmd.Name())
		default:
			futureCommands = append(futureCommands, cmd.Name())
		}
	}

	// Write Core Commands section
	if len(coreCommands) > 0 {
		fmt.Fprintln(file, "## Core Commands")
		fmt.Fprintln(file)
		for _, cmdName := range coreCommands {
			cmd := findCommand(rootCmd, cmdName)
			if cmd != nil {
				writeCommandEntry(file, cmd, "")
			}
		}
	}

	// Write Future Commands section
	if len(futureCommands) > 0 {
		fmt.Fprintln(file, "## Future Commands")
		fmt.Fprintln(file)
		fmt.Fprintln(file, "_These commands are planned for future implementation:_")
		fmt.Fprintln(file)
		for _, cmdName := range futureCommands {
			cmd := findCommand(rootCmd, cmdName)
			if cmd != nil {
				writeCommandEntry(file, cmd, "")
			}
		}
	}

	// Write Shell Completion section
	if cmd := findCommand(rootCmd, "completion"); cmd != nil {
		fmt.Fprintln(file, "## Shell Completion")
		fmt.Fprintln(file)
		writeCommandEntry(file, cmd, "")
	}

	// Write footer with generation info
	fmt.Fprintln(file, "---")
	fmt.Fprintln(file)
	fmt.Fprintln(file, "_This documentation is automatically generated from the Zen command definitions._")
	fmt.Fprintf(file, "_Last updated: %s_\n", time.Now().UTC().Format("2006-01-02"))

	return nil
}

// writeCommandEntry writes a single command entry to the index
func writeCommandEntry(w io.Writer, cmd *cobra.Command, prefix string) {
	fileName := strings.ReplaceAll(cmd.CommandPath(), " ", "_") + ".md"
	fmt.Fprintf(w, "### [zen %s](%s)\n", cmd.Name(), fileName)
	fmt.Fprintf(w, "%s\n", cmd.Short)
	fmt.Fprintln(w)
}

// findCommand finds a command by name in the command tree
func findCommand(root *cobra.Command, name string) *cobra.Command {
	for _, cmd := range root.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

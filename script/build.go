// Build tasks for the Zen CLI project.
//
// Usage:  go run script/build.go [<tasks>...] [<env>...]
//
// Known tasks are:
//
//   bin/zen:
//     Builds the main executable.
//     Supported environment variables:
//     - ZEN_VERSION: determined from git by default
//     - ZEN_BUILD_TAGS: additional build tags
//     - SOURCE_DATE_EPOCH: enables reproducible builds
//     - GO_LDFLAGS: additional linker flags
//
//   build-all:
//     Builds binaries for all supported platforms.
//
//   test:
//     Runs all tests (unit + integration + e2e).
//
//   test-unit:
//     Runs unit tests with coverage.
//
//   test-integration:
//     Runs integration tests.
//
//   test-e2e:
//     Runs end-to-end tests.
//
//   lint:
//     Runs linting checks.
//
//   security:
//     Runs security analysis.
//
//   deps:
//     Downloads and tidies dependencies.
//
//   clean:
//     Deletes all built files.
//
//   dev-setup:
//     Sets up development environment.
//

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var tasks map[string]func(string) error

func init() {
	tasks = map[string]func(string) error{
		"bin/zen": func(exe string) error {
			info, err := os.Stat(exe)
			if err == nil && !sourceFilesLaterThan(info.ModTime()) {
				fmt.Printf("%s: `%s` is up to date.\n", self, exe)
				return nil
			}

			ldflags := os.Getenv("GO_LDFLAGS")
			ldflags = fmt.Sprintf("-X github.com/daddia/zen/pkg/cmd/factory.version=%s %s", version(), ldflags)
			ldflags = fmt.Sprintf("-X github.com/daddia/zen/pkg/cmd/factory.commit=%s %s", commit(), ldflags)
			ldflags = fmt.Sprintf("-X github.com/daddia/zen/pkg/cmd/factory.buildTime=%s %s", buildTime(), ldflags)

			buildTags := os.Getenv("ZEN_BUILD_TAGS")

			args := []string{"go", "build", "-trimpath"}
			if buildTags != "" {
				args = append(args, "-tags", buildTags)
			}
			args = append(args, "-ldflags", ldflags, "-o", exe, "./cmd/zen")

			return run(args...)
		},
		"build-all": func(_ string) error {
			platforms := []string{
				"linux/amd64", "linux/arm64",
				"darwin/amd64", "darwin/arm64",
				"windows/amd64",
			}

			if err := os.MkdirAll("bin", 0755); err != nil {
				return err
			}

			fmt.Println("Building zen binaries for all platforms...")
			for _, platform := range platforms {
				parts := strings.Split(platform, "/")
				goos, goarch := parts[0], parts[1]

				output := fmt.Sprintf("bin/zen-%s-%s", goos, goarch)
				if goos == "windows" {
					output += ".exe"
				}

				fmt.Printf("  Building for %s/%s...\n", goos, goarch)

				ldflags := os.Getenv("GO_LDFLAGS")
				ldflags = fmt.Sprintf("-X github.com/daddia/zen/pkg/cmd/factory.version=%s %s", version(), ldflags)
				ldflags = fmt.Sprintf("-X github.com/daddia/zen/pkg/cmd/factory.commit=%s %s", commit(), ldflags)
				ldflags = fmt.Sprintf("-X github.com/daddia/zen/pkg/cmd/factory.buildTime=%s %s", buildTime(), ldflags)

				buildTags := os.Getenv("ZEN_BUILD_TAGS")

				args := []string{"go", "build", "-trimpath"}
				if buildTags != "" {
					args = append(args, "-tags", buildTags)
				}
				args = append(args, "-ldflags", ldflags, "-o", output, "./cmd/zen")

				cmd := exec.Command(args[0], args[1:]...)
				cmd.Env = append(os.Environ(),
					"GOOS="+goos,
					"GOARCH="+goarch,
				)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				if err := cmd.Run(); err != nil {
					fmt.Printf("  ✗ Failed to build for %s/%s\n", goos, goarch)
					return err
				}
				fmt.Printf("  ✓ Built: %s\n", output)
			}
			fmt.Println("✓ All binaries built successfully")
			return nil
		},
		"test": func(_ string) error {
			fmt.Println("Running all tests...")
			if err := tasks["test-unit"](""); err != nil {
				return err
			}
			if err := tasks["test-integration"](""); err != nil {
				return err
			}
			if err := tasks["test-e2e"](""); err != nil {
				return err
			}
			fmt.Println("✓ All tests completed")
			return nil
		},
		"test-unit": func(_ string) error {
			fmt.Println("Running unit tests...")
			if err := os.MkdirAll("coverage", 0755); err != nil {
				return err
			}
			args := []string{"go", "test", "-v", "-race",
				"-coverprofile=coverage/coverage.out", "-covermode=atomic",
				"-timeout=30s", "./internal/...", "./pkg/..."}
			if err := run(args...); err != nil {
				return err
			}
			fmt.Println("✓ Unit tests completed")
			return nil
		},
		"test-integration": func(_ string) error {
			fmt.Println("Running integration tests...")
			if err := os.MkdirAll("coverage", 0755); err != nil {
				return err
			}
			args := []string{"go", "test", "-v", "-tags=integration", "-timeout=60s",
				"-coverprofile=coverage/integration-coverage.out", "-covermode=atomic",
				"./test/integration/..."}
			return run(args...)
		},
		"test-e2e": func(_ string) error {
			fmt.Println("Running end-to-end tests...")
			// Build first
			if err := tasks["bin/zen"]("bin/zen" + exeSuffix()); err != nil {
				return err
			}
			args := []string{"go", "test", "-v", "-tags=e2e", "-timeout=120s", "./test/e2e/..."}
			if err := run(args...); err != nil {
				return err
			}
			fmt.Println("✓ End-to-end tests completed")
			return nil
		},
		"lint": func(_ string) error {
			fmt.Println("Running linter...")
			if _, err := exec.LookPath("golangci-lint"); err == nil {
				return run("golangci-lint", "run", "--timeout=5m")
			}

			fmt.Println("! golangci-lint not installed, running basic checks...")
			if err := run("gofmt", "-d", "-s", "."); err != nil {
				return err
			}
			if err := run("go", "vet", "./..."); err != nil {
				return err
			}
			fmt.Println("✓ Basic checks completed")
			return nil
		},
		"security": func(_ string) error {
			fmt.Println("Running security analysis...")
			if _, err := exec.LookPath("gosec"); err == nil {
				if err := run("gosec", "-quiet", "./..."); err != nil {
					return err
				}
				fmt.Println("✓ Security analysis completed")
			} else {
				fmt.Println("! gosec not installed, skipping security analysis")
				fmt.Println("  Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest")
			}
			return nil
		},
		"deps": func(_ string) error {
			fmt.Println("Downloading dependencies...")
			if err := run("go", "mod", "download"); err != nil {
				return err
			}
			if err := run("go", "mod", "tidy"); err != nil {
				return err
			}
			fmt.Println("✓ Dependencies updated")
			return nil
		},
		"clean": func(_ string) error {
			fmt.Println("Cleaning build artifacts...")
			if err := run("go", "clean"); err != nil {
				return err
			}
			dirs := []string{"bin", "coverage", "dist"}
			for _, dir := range dirs {
				if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
					return err
				}
			}
			fmt.Println("✓ Clean completed")
			return nil
		},
		"dev-setup": func(_ string) error {
			fmt.Println("Setting up development environment...")

			tools := map[string]string{
				"golangci-lint": "github.com/golangci/golangci-lint/cmd/golangci-lint@latest",
				"gosec":         "github.com/securecodewarrior/gosec/v2/cmd/gosec@latest",
			}

			for tool, pkg := range tools {
				if _, err := exec.LookPath(tool); err != nil {
					fmt.Printf("Installing %s...\n", tool)
					if err := run("go", "install", pkg); err != nil {
						return fmt.Errorf("failed to install %s: %w", tool, err)
					}
				}
			}

			fmt.Println("✓ Development environment setup completed")
			return nil
		},
	}
}

var self string

func main() {
	args := os.Args[:1]
	for _, arg := range os.Args[1:] {
		if idx := strings.IndexRune(arg, '='); idx >= 0 {
			os.Setenv(arg[:idx], arg[idx+1:])
		} else {
			args = append(args, arg)
		}
	}

	if len(args) < 2 {
		if isWindowsTarget() {
			args = append(args, filepath.Join("bin", "zen.exe"))
		} else {
			args = append(args, "bin/zen")
		}
	}

	self = filepath.Base(args[0])
	if self == "build" {
		self = "build.go"
	}

	for _, task := range args[1:] {
		t := tasks[normalizeTask(task)]
		if t == nil {
			fmt.Fprintf(os.Stderr, "Don't know how to build task `%s`.\n", task)
			fmt.Fprintln(os.Stderr, "\nAvailable tasks:")
			for taskName := range tasks {
				fmt.Fprintf(os.Stderr, "  %s\n", taskName)
			}
			os.Exit(1)
		}

		err := t(task)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintf(os.Stderr, "%s: building task `%s` failed.\n", self, task)
			os.Exit(1)
		}
	}
}

func isWindowsTarget() bool {
	if os.Getenv("GOOS") == "windows" {
		return true
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return false
}

func exeSuffix() string {
	if isWindowsTarget() {
		return ".exe"
	}
	return ""
}

func version() string {
	if versionEnv := os.Getenv("ZEN_VERSION"); versionEnv != "" {
		return versionEnv
	}
	if desc, err := cmdOutput("git", "describe", "--tags", "--always", "--dirty"); err == nil {
		return desc
	}
	return "dev"
}

func commit() string {
	if rev, err := cmdOutput("git", "rev-parse", "--short", "HEAD"); err == nil {
		return rev
	}
	return "unknown"
}

func buildTime() string {
	t := time.Now()
	if sourceDate := os.Getenv("SOURCE_DATE_EPOCH"); sourceDate != "" {
		if sec, err := strconv.ParseInt(sourceDate, 10, 64); err == nil {
			t = time.Unix(sec, 0)
		}
	}
	return t.UTC().Format("2006-01-02T15:04:05Z")
}

func sourceFilesLaterThan(t time.Time) bool {
	foundLater := false
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Ignore access denied errors on Windows
			if path != "." && isAccessDenied(err) {
				fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
				return nil
			}
			return err
		}
		if foundLater {
			return filepath.SkipDir
		}
		if len(path) > 1 && (path[0] == '.' || path[0] == '_') {
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}
		if info.IsDir() {
			if name := filepath.Base(path); name == "vendor" || name == "node_modules" || name == "bin" || name == "coverage" {
				return filepath.SkipDir
			}
			return nil
		}
		if path == "go.mod" || path == "go.sum" || (strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go")) {
			if info.ModTime().After(t) {
				foundLater = true
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}
	return foundLater
}

func isAccessDenied(err error) bool {
	var pe *os.PathError
	return errors.As(err, &pe) && strings.Contains(pe.Err.Error(), "Access is denied")
}

func announce(args ...string) {
	fmt.Println(shellInspect(args))
}

func run(args ...string) error {
	exe, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}
	announce(args...)
	cmd := exec.Command(exe, args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func cmdOutput(args ...string) (string, error) {
	exe, err := exec.LookPath(args[0])
	if err != nil {
		return "", err
	}
	cmd := exec.Command(exe, args[1:]...)
	cmd.Stderr = io.Discard
	out, err := cmd.Output()
	return strings.TrimSuffix(string(out), "\n"), err
}

func shellInspect(args []string) string {
	fmtArgs := make([]string, len(args))
	for i, arg := range args {
		if strings.ContainsAny(arg, " \t'\"") {
			fmtArgs[i] = fmt.Sprintf("%q", arg)
		} else {
			fmtArgs[i] = arg
		}
	}
	return strings.Join(fmtArgs, " ")
}

func normalizeTask(t string) string {
	return filepath.ToSlash(strings.TrimSuffix(t, ".exe"))
}

#!/usr/bin/env pwsh
# Build script for Zen CLI (PowerShell)
# Usage: .\script\build.ps1 [tasks...] [env_vars...]

param(
    [Parameter(ValueFromRemainingArguments=$true)]
    [string[]]$Arguments
)

# Pass all arguments to the Go build script
& go run script\build.go @Arguments

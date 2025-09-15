# Install zen from source

1. Verify that you have Go 1.25+ installed

   ```sh
   $ go version
   ```

   If `go` is not installed, follow instructions on [the Go website](https://golang.org/doc/install).

2. Clone this repository

   ```sh
   $ git clone https://github.com/daddia/zen.git zen
   $ cd zen
   ```

3. Build and install

   **Unix-like systems**

   ```sh
   # installs to '/usr/local' by default; sudo may be required, or sudo -E for configured go environments
   $ make install

   # or, install to a different location
   $ make install prefix=/path/to/gh
   ```

   **Windows**

   ```pwsh
   # PowerShell - build the `bin\zen.exe` binary
   > .\script\build.ps1
   
   # or use the batch file in Command Prompt
   > script\build.bat
   
   # or run directly with Go
   > go run script\build.go
   ```

   There is no install step available on Windows.

4. Run `zen version` to check if it worked.

   **Windows**

   Run `bin\zen version` to check if it worked.

## Cross-compiling binaries for different platforms

You can use any platform with Go installed to build a binary that is intended for another platform or CPU architecture. This is achieved by setting environment variables such as GOOS and GOARCH.

For example, to compile the `zen` binary for the 32-bit Raspberry Pi OS:

```sh
# on a Unix-like system:
$ GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 make clean bin/zen
```

```pwsh
# on Windows, pass environment variables as arguments to the build script:
> .\script\build.ps1 clean bin\zen GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0
# or
> go run script\build.go clean bin\zen GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0
```

Run `zen tool dist list` to list all supported values of GOOS/GOARCH.

Tip: to reduce the size of the resulting binary, you can use `GO_LDFLAGS="-s -w"`. This omits symbol tables used for debugging. See the list of [supported linker flags](https://golang.org/cmd/link/).

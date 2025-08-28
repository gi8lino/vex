# vex

[![Go Report Card](https://goreportcard.com/badge/github.com/gi8lino/vex?style=flat-square)](https://goreportcard.com/report/github.com/gi8lino/vex)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/gi8lino/vex)
[![Release](https://img.shields.io/github/release/gi8lino/vex.svg?style=flat-square)](https://github.com/gi8lino/vex/releases/latest)
[![GitHub tag](https://img.shields.io/github/tag/gi8lino/vex.svg?style=flat-square)](https://github.com/gi8lino/vex/releases/latest)
![Tests](https://github.com/gi8lino/vex/actions/workflows/tests.yml/badge.svg)
[![Build](https://github.com/gi8lino/vex/actions/workflows/release.yml/badge.svg)](https://github.com/gi8lino/vex/actions/workflows/release.yml)
[![license](https://img.shields.io/github/license/gi8lino/vex.svg?style=flat-square)](LICENSE)

---

`vex` is a fast, single-binary drop-in replacement for GNU `envsubst`.
It expands environment variables in text streams with full POSIX and extended operator support.
Unlike `envsubst`, `vex` uses a tokenizer + finite state machine written in Go, with **no regex dependencies**.
It runs blazingly fast, works with arbitrarily large files, and is safe for init containers where shells are unavailable.

## Features

- **POSIX-compatible expansions**:
  `$VAR`, `${VAR}`, `${VAR:-default}`, `${VAR:=assign}`, `${VAR:+alt}`, `${VAR:?error}`
- **Operator extensions**:

  - **Case transforms**: `${VAR^}`, `${VAR^^}`, `${VAR,}`, `${VAR,,}`
  - **Length**: `${#VAR}`
  - **Substring**: `${VAR:offset[:len]}`
  - **Trimming**: `${VAR#prefix}`, `${VAR##prefix}`, `${VAR%suffix}`, `${VAR%%suffix}`
  - **Replace**: `${VAR/pat/repl}`, `${VAR//pat/repl}`
  - **Quoting**: `${VAR@Q}` (shell), `${VAR@J}` (JSON), `${VAR@Y}` (YAML)

- **Colorized output** (`--colored`) with semantic colors:

  - Green → substituted value
  - Yellow → default/assignment value
  - Orange → empty variable
  - Magenta → unset variable
  - Red → engine/internal error
  - Purple → user error message
  - Gray → filtered variable

- **Strict modes**: exit on unset/empty values
- **Safe in-place mode**: temp write + atomic rename with backup support
- **Configurable allow-lists**: restrict by name, prefix, or suffix
- **Portable**: one static Go binary, no shell, no external deps

## Installation

### Go

```sh
go install github.com/gi8lino/vex@latest
```

### Binaries

Prebuilt binaries for Linux and macOS (amd64, arm64) are available on the [Releases page](https://github.com/gi8lino/vex/releases).

### Docker

```sh
docker run --rm -i ghcr.io/gi8lino/vex:latest < input.txt
```

## Usage

### Basic

```sh
# substitute stdin to stdout
echo 'Hello $USER' | vex
```

### Files

```sh
# concatenate files to stdout
vex file1.txt file2.txt

# edit files in place
vex -i config.yaml

# with backup
vex -i --backup=.bak config.yaml
```

### Flags

| Flag                   | Short | Description                                                     |
| :--------------------- | :---- | :-------------------------------------------------------------- |
| `--in-place`           | `-i`  | Edit files in place                                             |
| `--backup EXT`         | `-b`  | Create a backup file before replacing                           |
| `--colored`            | `-c`  | Colorize output (stdout + diagnostics)                          |
| `--strict`             | `-x`  | Equivalent to `--error-unset --error-empty`                     |
| `--error-unset`        | `-u`  | Error if a variable is unset                                    |
| `--error-empty`        | `-e`  | Error if a variable expands to empty                            |
| `--keep-unset`         | `-U`  | Keep `${VAR}` literal if unset                                  |
| `--keep-empty`         | `-E`  | Keep `${VAR}` literal if empty                                  |
| `--keep-vars`          | `-K`  | Keep all `${VAR}` literals (implies both)                       |
| `--no-ops`             |       | Treat operator forms as literal text (envsubst-compatible mode) |
| `--literal-dollar`     | `-l`  | Disable `\$` escaping (treat as backslash + dollar)             |
| `--prefix P`           | `-p`  | Only expand variables starting with `P`                         |
| `--suffix S`           | `-s`  | Only expand variables ending with `S`                           |
| `--variable V`         | `-v`  | Only expand variables named `V`                                 |
| `--extra-vars PATH...` | `-e`  | Read extra variables from file (use `-` for stdin)              |

## Operators with Examples

```sh
USER=alice HOME=/home/alice

# Defaulting
vex <<< 'Hello ${NAME:-world}'
# → Hello world

# Assign if unset
vex <<< '${NAME:=bob}'
# → bob (and sets NAME=bob in env)

# Alternate if set
vex <<< '${USER:+present}'
# → present

# Error if unset
vex <<< '${MISSING:?must be set}'
# → expansion error, exit code 2

# Case transform
vex <<< 'User: ${USER^}'
# → User: Alice

# Uppercase all
vex <<< '${USER^^}'
# → ALICE

# Lowercase all
vex <<< '${USER,,}'
# → alice

# Length
vex <<< '${#HOME}'
# → 11

# Substring
vex <<< '${HOME:6:5}'
# → alice

# Prefix trim
vex <<< '${HOME#/home/}'
# → alice

# Replace
vex <<< '${HOME/home/ROOT}'
# → /ROOT/alice

# Replace all
vex <<< '${HOME//a/A}'
# → /home/ Alice  (with A instead of a)

# Quoting
vex <<< '${USER@J}'
# → "alice"
```

## Providing Custom Variables (`--extra-vars`)

By default, `vex` expands variables from the current process environment (`os.Environ`).
With the `--extra-vars` flag you can override this by loading additional variables from a file.
Variables provided via `--extra-vars` **always override** values from the system environment.

**Example:**

```sh
# vars.env contains:
#   FOO=bar
#   BAZ=qux

vex config.txt --extra-vars .env
```

## Benchmarks

`vex` is optimized for speed with a streaming tokenizer and finite-state machine.
Throughput depends on how you use it:

- **CLI one-shot calls** (common in scripts, CI, init-containers): \~**50-75 MB/s** on large files, including process startup + pipes.
- **Library / long-running process** (no spawn overhead): \~**95-100 MB/s** steady-state parsing throughput.

Run the suite yourself:

```sh
make bench-all
```

_Environment:_ macOS (Darwin amd64), Intel i5-8257U, Go 1.25, `CGO_ENABLED=0`.
_Files are pre-created; reads are typically page-cache hits (so results mainly reflect parsing and pipe costs rather than disk I/O)._

```text
# CLI onse-shot calls
BenchmarkVexCLI_OneSmallFile-8     283     7.37ms/op     8.89 MB/s
BenchmarkVexCLI_ManySmallFiles-8    10   204.20ms/op    64.21 MB/s
BenchmarkVexCLI_OneBigFile-8        25   100.03ms/op    52.41 MB/s
BenchmarkVexCLI_ManyBigFiles-8       3   691.10ms/op    75.86 MB/s

# Libary / long-running process
BenchmarkVexRun_OneBigFile-8        44    53.88ms/op    97.10 MB/s
BenchmarkVexRun_ManyBigFiles-8       4   531.80ms/op    98.59 MB/s
```

_On Linux you can pin the CPU governor for more stable results:_

```sh
make perf-on   # performance governor
make bench-all
make perf-off  # powersave
```

## License

Apache 2.0 -- see [LICENSE](LICENSE)

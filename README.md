# PTO Calculator

Terminal-based PTO planner built with Go and Bubble Tea.

It lets you:

- set up your PTO profile
- calculate PTO on a future date
- add, list, and remove trips
- add, list, and remove holidays
- validate future trips against your projected PTO balance

## Requirements

- Go `1.26` or newer

## Quick Start

From the project root:

```bash
make run
```

That starts the TUI.

If you prefer plain Go commands:

```bash
go run .
```

## Common Commands

Run the TUI:

```bash
make run
```

Build the binary:

```bash
make build
```

Run formatting:

```bash
make fmt
```

Run tests:

```bash
make test
```

Clean the built binary:

```bash
make clean
```

## Data Files

The app reads and writes YAML files in the project root:

- `config.yml`
- `trips.yml`
- `holidays.yml`

On first run, the TUI will prompt for your config if `config.yml` is missing or empty.

## Running the App

The current entrypoint is:

```bash
go run . --type tui
```

The default is already `tui`, so `go run .` is enough.

## Notes

- PTO accrual is projected over time using your saved config.
- Trips are validated against overlapping dates and future balance impact.
- Holidays and off Fridays are excluded from trip PTO usage.

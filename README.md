# HEART Lattigo Hardware Bootstrapping Model

This repository contains a Go hardware golden model for CKKS bootstrapping on
top of [Lattigo v6](https://github.com/tuneinsight/lattigo). The current focus
is a fully hardware-modeled slim CKKS bootstrapping path with functional-unit
usage accounting.

## Repository Layout

```text
.
|-- *.go                         Go source and tests for the hardware model
|-- configs/
|   `-- bootstrap_config.example.json
|-- docs/
|   |-- BOOTSTRAPPING_DETAILS.md
|   `-- slim_bootstrapping.pdf
|-- reports/
|   `-- bootstrap_usage_report.csv
|-- scripts/
|   `-- publish_github.ps1
|-- .github/workflows/
|   `-- go.yml
|-- go.mod
`-- go.sum
```

## Main Components

| File | Purpose |
| --- | --- |
| `bootstrapping_hw.go` | Hardware-modeled CKKS bootstrapping flow |
| `hardware_evaluator.go` | CKKS evaluator wrapper routed through hardware units |
| `functional_unit.go` | NTT, base conversion, automorphism, EWU, and scaling units |
| `functional_unit_usage.go` | Functional-unit usage counters |
| `basic_operation.go` | Local ciphertext operation helpers |
| `bootstrap_config.go` | Bootstrapping parameter configuration |
| `bootstrap_usage_report.go` | Usage report generator |
| `docs/BOOTSTRAPPING_DETAILS.md` | Step-by-step bootstrapping method notes |

## Requirements

- Go 1.25 or newer
- Lattigo v6.2.0, resolved through `go.mod`

## Run Tests

```powershell
go test ./...
```

The full bootstrapping correctness test can take several minutes:

```powershell
go test -run TestSlimBootstrapperHW -v
```

## Generate Functional-Unit Usage

Use the default example parameter set:

```powershell
go run . -mode=bootstrap-usage -config .\configs\bootstrap_config.example.json -out .\reports\bootstrap_usage_report.csv -format csv -by-step=true
```

The command prints the per-stage usage and writes the CSV report.

## Change Bootstrapping Parameters

Edit:

```text
configs/bootstrap_config.example.json
```

The config exposes the ring size, Q/P prime layout, DFT levels, slot count,
EvalMod degree, double-angle rounds, and related bootstrapping settings.

## Current Method

The active code path is CKKS bootstrapping:

```text
SlotsToCoeffs -> ScaleDown -> ModUp -> CoeffsToSlots -> EvalMod -> Recombine
```

See `docs/BOOTSTRAPPING_DETAILS.md` for the detailed operation-level
description. The included slim bootstrapping paper is useful background, but it
describes FV/BGV slim mode. This code implements CKKS EvalMod-based
bootstrapping.

## GitHub Target

Suggested repository name:

```text
heart-lattigo-hw-bootstrap
```

Suggested remote URL:

```text
https://github.com/kenny0915/heart-lattigo-hw-bootstrap.git
```

After installing Git and GitHub CLI, authenticate once:

```powershell
gh auth login
```

Then publish this directory:

```powershell
.\scripts\publish_github.ps1 -Owner kenny0915 -RepoName heart-lattigo-hw-bootstrap -Visibility public
```

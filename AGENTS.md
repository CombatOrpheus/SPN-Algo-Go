# Agent Instructions

This document provides instructions for AI agents working on this repository.

## Project Overview

This is a Go project that generates benchmark datasets for Stochastic Petri Nets (SPNs). The main application is located in `cmd/spn-benchmark-ds`, and the core logic is in `internal/pkg`.

## Development Environment

The project uses Go modules for dependency management. To install the dependencies, run the following command:

```
go mod tidy
```

## Running Tests

To run the tests, use the following command:

```
go test ./...
```

## Code Style

The code should be formatted using `go fmt`. You can format the code by running the following command:

```
go fmt ./...
```

All public functions, methods, and structs should have comprehensive GoDoc docstrings.

## Data Structures

The `PetriNet` and `ReachabilityGraph` data structures use flattened slices for their matrices to improve performance. This convention should be maintained in any future modifications. The `At` and `Set` methods should be used to access and modify the matrix elements.

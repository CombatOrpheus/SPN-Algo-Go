# SPN Benchmark Dataset Generator

This project is a command-line tool for generating benchmark datasets for Stochastic Petri Nets (SPNs). It is a Go implementation of the Python project [SPN-Benchmark-DS](https://github.com/mr-shimo/SPN-Benchmark-DS).

## Architecture

The project is divided into two main parts: the `cmd` directory, which contains the main application, and the `internal` directory, which contains the core logic of the application.

The `internal` directory is further divided into the following packages:

*   `analysis`: Contains the logic for analyzing SPNs.
*   `augmentation`: Contains the logic for augmenting SPNs.
*   `generation`: Contains the logic for generating SPNs.
*   `petrinet`: Contains the data structures for representing SPNs.
*   `report`: Contains the logic for generating reports.
*   `spn`: Contains the protobuf definitions for SPNs.

## Setup

To build the project, you will need to have Go installed. You can then build the project by running the following command:

```
go build ./...
```

## Usage

To run the project, you will need to provide a configuration file. A sample configuration file is provided in `config.yaml`. You can run the project by running the following command:

```
go run cmd/spn-benchmark-ds/main.go cmd/spn-benchmark-ds/config.go --config config.yaml
```

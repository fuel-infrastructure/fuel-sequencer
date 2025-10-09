# Blob Profiler Comparison Tool

This tool creates a comparison scatter plot between two blob profiler systems by analyzing their performance data.

## Overview

The comparison tool extracts duration data from profiler reports and creates a scatter plot showing:
- **X-axis**: Blob size configurations (1x100 KiB, 2x300 KiB, 1x1 MiB, 2x3 MiB, 1x5 MiB)
- **Y-axis**: Total duration in minutes (0-4 minutes range)
- **Data points**: 10 points total (5 runs × 2 systems)

## Usage

```bash
cd scripts/blob_profiling/cmd/comparison
go run main.go -decoupled <decoupled_system_path> -fullpost <fullpost_system_path>
```

### Example

```bash
go run main.go \
  -decoupled ../../benchnet_eu_decoupled_output \
  -fullpost ../../benchnet_eu_fullpost_output
```

## Output

The tool generates:
- **PNG Image**: `comparison/images/system_comparison.png` - High-resolution scatter plot
- **CSV Data**: `comparison/data/decoupled_data.csv` and `comparison/data/fullpost_data.csv` - Raw data files
- **Gnuplot Script**: `comparison/data/comparison.gp` - Temporary script (cleaned up after execution)

## Data Structure

The tool expects profiler report JSON files in the following structure:
```
system_output/
├── blob_size_config/
│   └── timestamp/
│       └── profiler_report.json
└── ...
```

## Prerequisites

- **DuckDB**: For data processing (brew install duckdb / apt install duckdb)
- **Gnuplot**: For graph generation (brew install gnuplot / apt install gnuplot)

## Generated Plot Features

- **Title**: "Original Complete Blobs Posted vs Ongoing Decoupled Blobs (Metadata Only)"
- **Legend**: Blue circles (MsgPostBlobMetadata) vs Orange squares (MsgPostBlob)
- **Grid**: Enabled for better readability
- **Y-axis Range**: 0-4 minutes (focused on the relevant performance range)
- **High Resolution**: 1200×800 PNG output

## Data Interpretation

The scatter plot shows:
- **MsgPostBlobMetadata** (Blue circles): Decoupled blob metadata posting system
- **MsgPostBlob** (Orange squares): Full blob posting system

Each point represents one run's total duration for a specific blob size configuration. The comparison helps identify performance differences between the two systems across different blob sizes.

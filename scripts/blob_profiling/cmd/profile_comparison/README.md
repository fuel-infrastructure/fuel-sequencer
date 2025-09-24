# Blob Profiler Comparison Tool

This tool helps compare profiling results across different versions of the fuel-sequencer and blob-storage components by analyzing profiler reports and generating comparison tables.

## Features

- **Grouped Analysis**: Groups runs by composite key of fuel-sequencer and blob-storage commit hashes, profile type, and blob size
- **Version Tracking**: Tracks both fuel-sequencer and blob-storage versions from go.mod
- **Multiple Output Formats**: Table and CSV output formats
- **Comprehensive Metrics**: Calculates max throughput, min throughput, average throughput, variance, success rate, and duration across multiple runs
- **Performance Sorting**: Results sorted by max throughput in descending order (highest performance first)
- **Profile Type Comparison**: Compares different profile types (linear, burst, etc.) side by side
- **Blob Size Analysis**: Shows performance metrics for different blob sizes
- **Individual Run Analysis**: Option to view each individual profiling run separately

## Usage

### Basic Usage

```bash
# Generate a comparison table (grouped by commit, profile type, and blob size)
go run cmd/profile_comparison/main.go output/

# Output in CSV format (grouped by commit, profile type, and blob size)
go run cmd/profile_comparison/main.go output/ --csv

# Show individual runs instead of grouping
go run cmd/profile_comparison/main.go output/ --individual

# Show individual runs in CSV format
go run cmd/profile_comparison/main.go output/ --individual --csv

# Show help
go run cmd/profile_comparison/main.go --help
```

### Output Modes

#### Grouped Mode (Default)
Groups runs by composite key of fuel-sequencer and blob-storage commit hashes, profile type, and blob size. This is useful for comparing performance across different versions and configurations. Multiple runs with the same combination are aggregated together, showing min, max, average, and variance metrics.

#### Individual Mode (`--individual`)
Shows each individual profiling run as a separate row. This is useful for analyzing different profile configurations, blob sizes, and detailed performance metrics across all runs. Results are sorted by max throughput in descending order.

### Output Formats

#### Table Format (Default)
```
SEQUENCER COMMIT  SEQ SHORT  BLOB STORAGE COMMIT  BLOB SHORT  PROFILE TYPE  BLOB SIZE          MAX THROUGHPUT (MiB/s)  MIN THROUGHPUT (MiB/s)  AVG THROUGHPUT (MiB/s)  VARIANCE  SUCCESS RATE (%)  DURATION  RUNS
---------------   ---------  -------------------  ----------  -----------   ---------          -------------------     -------------------     -------------------     --------  ---------------   --------  ----
fed320a6          fed320a6   unknown                          linear        1 MiB blobs        3.50                    3.50                    3.50                    0.00      100.0             7.3m      1
c92e98ef          c92e98ef   unknown                          linear        1 MiB blobs        3.29                    2.63                    2.91                    0.08      99.8              7.4m      3
c92e98ef          c92e98ef   unknown                          linear        10 MiB blobs       2.89                    1.37                    2.13                    0.58      93.5              6.4m      2
```

#### Individual Mode Table Format
```
SEQUENCER COMMIT  SEQ SHORT  BLOB STORAGE COMMIT  BLOB SHORT  PROFILE TYPE  BLOB SIZE          MAX THROUGHPUT (MiB/s)  MIN THROUGHPUT (MiB/s)  AVG THROUGHPUT (MiB/s)  VARIANCE  SUCCESS RATE (%)  DURATION  TIMESTAMP
---------------   ---------  -------------------  ----------  -----------   ---------          -------------------     -------------------     -------------------     --------  ---------------   --------  ---------
fed320a6          fed320a6   unknown                          linear        1 MiB blobs        3.50                    3.50                    3.50                    0.00      100.0             7.3m      2025-09-24 12:02:48
c92e98ef          c92e98ef   unknown                          linear        1 MiB blobs        3.29                    3.29                    3.29                    0.00      100.0             6.9m      2025-09-19 10:40:16
c92e98ef          c92e98ef   unknown                          linear        Distributed blobs  2.08                    2.08                    2.08                    0.00      98.3              3.8m      2025-09-19 10:34:33
```

#### CSV Format
```csv
SEQUENCER_COMMIT,SEQ_SHORT,BLOB_STORAGE_COMMIT,BLOB_SHORT,PROFILE_TYPE,BLOB_SIZE,MAX_THROUGHPUT_MIB_PER_SEC,MIN_THROUGHPUT_MIB_PER_SEC,AVG_THROUGHPUT_MIB_PER_SEC,VARIANCE,SUCCESS_RATE_PERCENT,DURATION,RUNS
fed320a6,fed320a6,unknown,,linear,1 MiB blobs,3.50,3.50,3.50,0.00,100.0,7.3m,1
c92e98ef,c92e98ef,unknown,,linear,1 MiB blobs,3.29,2.63,2.91,0.08,99.8,7.4m,3
c92e98ef,c92e98ef,unknown,,linear,10 MiB blobs,2.89,1.37,2.13,0.58,93.5,6.4m,2
```

## Metrics Explained

- **SEQUENCER COMMIT**: Short commit hash of the fuel-sequencer version
- **BLOB STORAGE COMMIT**: Short commit hash of the blob-storage version (from go.mod)
- **PROFILE TYPE**: Type of profiling test (linear, burst, etc.)
- **BLOB SIZE**: Size of blobs used in the test
- **MAX THROUGHPUT**: Highest throughput achieved across all runs for this commit/profile combination
- **MIN THROUGHPUT**: Lowest throughput achieved across all runs for this commit/profile combination
- **AVG THROUGHPUT**: Average throughput across all runs for this commit/profile combination
- **VARIANCE**: Throughput variance across all runs for this commit/profile combination
- **SUCCESS RATE**: Percentage of successful blob submissions
- **DURATION**: Average test duration
- **RUNS**: Number of runs aggregated for this row

## Version Tracking

### Blob-Storage Version Tracking

The profiler now tracks blob-storage versions by parsing the go.mod file. This provides:

1. **Automatic Version Detection**: Extracts blob-storage version from go.mod dependencies
2. **Commit Hash Extraction**: Parses commit hash from version strings like `v0.0.0-20250919082051-2fc207358bc3`
3. **Composite Grouping**: Groups runs by both fuel-sequencer and blob-storage commit hashes
4. **Backward Compatibility**: Handles existing reports that don't have blob-storage version information

### How It Works

The profiler reads the go.mod file and looks for the blob-storage dependency:
```go
github.com/fuel-infrastructure/blob-storage v0.0.0-20250919082051-2fc207358bc3
```

It then extracts:
- **Version**: `v0.0.0-20250919082051-2fc207358bc3`
- **Commit Hash**: `2fc207358bc3`
- **Short Hash**: `2fc20735`

## File Structure

The tool expects profiler reports to be located in the following structure:

```
output/
├── profile_name_1/
│   ├── 2025-09-16_T_19_03_24/
│   │   ├── profiler_report.json
│   │   ├── throughput.parquet
│   │   └── blobs.parquet
│   └── 2025-09-19_T_10_40_16/
│       ├── profiler_report.json
│       └── ...
└── profile_name_2/
    └── ...
```

## Integration with CI/CD

This tool can be integrated into CI/CD pipelines to automatically generate performance comparisons:

```bash
# In your CI pipeline
go run cmd/profile_comparison/main.go output/ --csv > performance_comparison.csv
go run cmd/profile_comparison/main.go output/ --individual --csv > individual_runs.csv
```

The CSV output can be easily imported into spreadsheet applications for further analysis, or processed by other tools for historical tracking and performance monitoring.

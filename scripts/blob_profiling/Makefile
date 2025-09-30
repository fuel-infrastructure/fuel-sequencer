# Blob Profiling Makefile

# Configuration
PROFILER_DIR := cmd/blob_profiler
DATA_DIR := cmd/blob_profiler
OUTPUT_DIR := output
IMAGES_DIR := output/images
DATA_OUTPUT_DIR := output/data

# Default target
.PHONY: all
all: build-profiler

# Build the profiler (now includes visualization functionality)
.PHONY: build-profiler
build-profiler:
	@echo "Building blob profiler..."
	@go build -o $(PROFILER_DIR)/blob_profiler ./$(PROFILER_DIR)
	@echo "✓ Blob profiler built successfully"

# Generate all graphs (standalone)
.PHONY: generate-graphs
generate-graphs: build-profiler
	@echo "Generating blob profiling graphs..."
	@mkdir -p $(IMAGES_DIR) $(DATA_OUTPUT_DIR)
	@cd $(DATA_DIR) && ../blob_profiler/blob_profiler -graphs -parquet .
	@echo "✓ All graphs generated in $(IMAGES_DIR)/"

# Generate graphs with cleanup (standalone)
.PHONY: generate-clean
generate-clean: build-profiler
	@echo "Generating blob profiling graphs with cleanup..."
	@mkdir -p $(IMAGES_DIR) $(DATA_OUTPUT_DIR)
	@cd $(DATA_DIR) && ../blob_profiler/blob_profiler -graphs -cleanup -parquet .
	@echo "✓ All graphs generated in $(IMAGES_DIR)/ (data files cleaned up)"

# Clean up generated files
.PHONY: clean
clean:
	@echo "Cleaning up generated files..."
	@rm -rf $(OUTPUT_DIR)
	@rm -f $(PROFILER_DIR)/blob_profiler
	@echo "✓ Cleanup completed"

# Check dependencies
.PHONY: check-deps
check-deps:
	@echo "Checking dependencies..."
	@command -v duckdb >/dev/null 2>&1 || { echo "❌ DuckDB not found. Please install DuckDB."; exit 1; }
	@command -v gnuplot >/dev/null 2>&1 || { echo "❌ gnuplot not found. Please install gnuplot."; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo "❌ Go not found. Please install Go."; exit 1; }
	@echo "✓ All dependencies found"

# Check if parquet files exist
.PHONY: check-data
check-data:
	@echo "Checking for parquet data files..."
	@test -f $(DATA_DIR)/linear_5_kib_per_s_blobs.parquet || { echo "❌ Blobs parquet file not found: $(DATA_DIR)/linear_5_kib_per_s_blobs.parquet"; exit 1; }
	@test -f $(DATA_DIR)/linear_5_kib_per_s_throughput.parquet || { echo "❌ Throughput parquet file not found: $(DATA_DIR)/linear_5_kib_per_s_throughput.parquet"; exit 1; }
	@echo "✓ Parquet data files found"

# Install dependencies (macOS)
.PHONY: install-deps-macos
install-deps-macos:
	@echo "Installing dependencies on macOS..."
	@brew install duckdb gnuplot
	@echo "✓ Dependencies installed"

# Install dependencies (Ubuntu/Debian)
.PHONY: install-deps-ubuntu
install-deps-ubuntu:
	@echo "Installing dependencies on Ubuntu/Debian..."
	@sudo apt-get update
	@sudo apt-get install -y duckdb gnuplot
	@echo "✓ Dependencies installed"

# Compare profiling results
.PHONY: compare
compare:
	@echo "Generating profiling comparison..."
	@go run cmd/profile_comparison/main.go $(OUTPUT_DIR)
	@echo "✓ Comparison completed"

# Compare profiling results (CSV format)
.PHONY: compare-csv
# @echo "Generating profiling comparison (CSV)..."
# @echo "✓ CSV comparison completed"
compare-csv:
	@go run cmd/profile_comparison/main.go $(OUTPUT_DIR) --csv

# Show individual profiling runs
.PHONY: compare-individual
compare-individual:
	@echo "Generating individual profiling runs..."
	@go run cmd/profile_comparison/main.go $(OUTPUT_DIR) --individual
	@echo "✓ Individual runs completed"

# Show individual profiling runs (CSV format)
.PHONY: compare-individual-csv
compare-individual-csv:
	@echo "Generating individual profiling runs (CSV)..."
	@go run cmd/profile_comparison/main.go $(OUTPUT_DIR) --individual --csv
	@echo "✓ Individual CSV runs completed"

# Help
.PHONY: help
help:
	@echo "Blob Profiling Visualization Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  all              - Build profiler (default)"
	@echo "  build-profiler   - Build the Go profiler binary (includes visualization)"
	@echo "  generate-graphs  - Generate all performance graphs (standalone)"
	@echo "  generate-clean  - Generate graphs and cleanup data files (standalone)"
	@echo "  compare          - Generate profiling comparison table"
	@echo "  compare-csv      - Generate profiling comparison in CSV format"
	@echo "  compare-individual - Show individual profiling runs"
	@echo "  compare-individual-csv - Show individual runs in CSV format"
	@echo "  clean            - Remove generated files"
	@echo "  check-deps       - Check if required dependencies are installed"
	@echo "  check-data       - Check if parquet data files exist"
	@echo "  install-deps-macos   - Install dependencies on macOS (requires Homebrew)"
	@echo "  install-deps-ubuntu  - Install dependencies on Ubuntu/Debian"
	@echo "  help             - Show this help message"
	@echo ""
	@echo "Usage:"
	@echo "  make                    # Build profiler"
	@echo "  make generate-graphs    # Generate graphs from existing parquet data"
	@echo "  make generate-clean     # Generate graphs and cleanup data files"
	@echo "  make compare            # Generate profiling comparison table"
	@echo "  make compare-csv         # Generate profiling comparison in CSV format"
	@echo "  make compare-individual  # Show individual profiling runs"
	@echo "  make compare-individual-csv # Show individual runs in CSV format"
	@echo "  make clean              # Clean up generated files"
	@echo "  make check-deps         # Check dependencies"
	@echo "  make check-data         # Check data files"
	@echo ""
	@echo "Generated plots:"
	@echo "  - throughput_analysis.png    - Expected vs Actual Throughput"
	@echo "  - blob_size_distribution.png - Blob Size Distribution"
	@echo "  - blob_timeline.png          - Blob Event Timeline"
	@echo "  - store_to_blobpool.png      - Time from Store to Blobpool"
	@echo "  - store_to_finalized.png     - Time from Store to Finalized"
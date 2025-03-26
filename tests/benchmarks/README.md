# Performance Benchmarks

This directory contains benchmark tests for the Export Trakt for Letterboxd application. These benchmarks measure the performance of key operations and help identify potential bottlenecks.

## Running Benchmarks

### Prerequisites

To run the benchmarks with real API data, you need:

1. Trakt.tv API credentials (Client ID and Client Secret)
2. Authentication token (will be created automatically if you've previously authenticated)

### Setting Environment Variables

Set your Trakt.tv API credentials as environment variables:

```bash
export TRAKT_CLIENT_ID="your_client_id"
export TRAKT_CLIENT_SECRET="your_client_secret"
```

### Running All Benchmarks

```bash
go test -bench=. ./tests/benchmarks -v
```

### Running Specific Benchmarks

```bash
# Run only API-related benchmarks
go test -bench=API ./tests/benchmarks -v

# Run only CSV generation benchmarks
go test -bench=CSV ./tests/benchmarks -v
```

### Memory Profiling

To include memory allocation statistics:

```bash
go test -bench=. -benchmem ./tests/benchmarks -v
```

### CPU Profiling

To generate a CPU profile:

```bash
go test -bench=. -cpuprofile=cpu.prof ./tests/benchmarks -v
```

Then analyze it with:

```bash
go tool pprof cpu.prof
```

### Memory Heap Profiling

To generate a memory heap profile:

```bash
go test -bench=. -memprofile=mem.prof ./tests/benchmarks -v
```

Then analyze it with:

```bash
go tool pprof mem.prof
```

## Interpreting Results

### Benchmark Output Format

The benchmark output follows this format:

```
BenchmarkName-NumCPUs    iterations    time/op    bytes/op    allocs/op
```

Where:

- **BenchmarkName**: The name of the benchmark function
- **NumCPUs**: The number of CPUs used
- **iterations**: How many times the test ran
- **time/op**: Average time per operation
- **bytes/op**: Average memory allocated per operation (with `-benchmem`)
- **allocs/op**: Average number of allocations per operation (with `-benchmem`)

### Example Output

```
BenchmarkGetWatchedMovies-8           5         240155432 ns/op        5438934 B/op      65843 allocs/op
BenchmarkGetRatedMovies-8            10         120483921 ns/op        2564841 B/op      31245 allocs/op
BenchmarkCSVGeneration-8            500           2385622 ns/op         425632 B/op       3214 allocs/op
```

### What to Look For

1. **High time/op**: Operations taking a long time may need optimization
2. **High bytes/op**: Excessive memory usage may indicate inefficient data structures
3. **High allocs/op**: Many allocations can impact garbage collector performance

## Benchmarked Operations

| Benchmark                    | Description                                | What It Measures                        |
| ---------------------------- | ------------------------------------------ | --------------------------------------- |
| BenchmarkGetWatchedMovies    | Retrieves watched movies from Trakt.tv API | API client performance, network latency |
| BenchmarkGetRatedMovies      | Retrieves rated movies from Trakt.tv API   | API client performance, network latency |
| BenchmarkGetWatchlistMovies  | Retrieves watchlist from Trakt.tv API      | API client performance, network latency |
| BenchmarkExportWatchedMovies | Full export process for watched movies     | End-to-end export performance           |
| BenchmarkCSVGeneration       | Generation of CSV files from movie data    | CSV formatting and file I/O performance |
| BenchmarkMovieDataProcessing | Processing of movie data before export     | Data transformation performance         |

## Comparing Benchmark Results

When making code changes, compare benchmark results to ensure performance doesn't degrade:

```bash
# Save current benchmark results
go test -bench=. ./tests/benchmarks -v > benchmark_before.txt

# Make code changes...

# Run benchmarks after changes
go test -bench=. ./tests/benchmarks -v > benchmark_after.txt

# Compare results (requires benchstat tool)
# Install with: go install golang.org/x/perf/cmd/benchstat@latest
benchstat benchmark_before.txt benchmark_after.txt
```

## Notes on API Benchmarks

- API benchmarks depend on network conditions and Trakt.tv API responsiveness
- Results may vary between runs due to external factors
- Consider running offline benchmarks for consistent results during development
- Be mindful of API rate limits when running benchmarks frequently

## 2026-04-06 - Optimized State Serialization in Reachability Graph

**Learning:** `fmt.Sprintf("%v", slice)` is a significant performance bottleneck when called millions of times in hot paths, such as converting state arrays (markings) into map keys for visited state checks. It relies heavily on reflection and dynamic type-checking, adding unnecessary overhead (benchmarked at ~1389 ns/op).

**Action:** Replace `fmt.Sprintf` with a manual `strings.Builder` and `strconv.Itoa` implementation for hot path string generation. This reduces reflection overhead and allocations, improving execution time to ~150 ns/op (a ~9x speedup) while maintaining human-readable output (e.g., `[1 2 3]`). Avoid micro-optimizations that completely sacrifice readability (like converting raw integer slice memory to binary strings) unless absolutely necessary, as it breaks debugging tools and semantic meaning.

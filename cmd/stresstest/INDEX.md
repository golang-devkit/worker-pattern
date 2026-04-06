# Stress Test Tool - Documentation Index

## 📍 Start Here

**New to this tool?** Start with [QUICKREF.md](QUICKREF.md) for a one-page overview.

**Want detailed examples?** See the session guide: `~/.copilot/session-state/.../USAGE_EXAMPLES.md`

**Need complete reference?** Read [README.md](README.md) for comprehensive documentation.

---

## �� Documentation Files

### [QUICKREF.md](QUICKREF.md) - Start Here! (125 lines)

**Perfect for:** Quick lookup, getting started, common commands

- Build & run in 30 seconds
- Command cheat sheet
- Output interpretation
- 5 example scenarios
- Troubleshooting quick tips

### [README.md](README.md) - Complete Reference (295 lines)

**Perfect for:** Understanding all features, building, detailed usage

- Feature overview
- Building instructions
- All command-line flags explained
- Usage examples with expected outputs
- Latency percentile explanation
- Test scenarios for 1-4+ workers
- Comparing deployments
- Performance notes & analysis
- Troubleshooting guide
- Future enhancements

### [TESTING.md](TESTING.md) - Test Scenarios (353 lines)

**Perfect for:** Specific test scenarios, interpreting results

- Baseline performance test
- Consistency verification
- Request batching
- Timeout simulation
- Performance comparison
- Advanced ramp-up tests
- Real results from deployment
- CI/CD integration
- Continuous monitoring setup

### [main.go](main.go) - Tool Implementation (241 lines)

**Perfect for:** Understanding the code

- Main testing engine
- Concurrent request handling
- Statistics collection
- Latency percentile calculation
- Output formatting

---

## 🎯 Quick Navigation by Task

### "I want to run a quick test"

→ [QUICKREF.md](QUICKREF.md) → Run `make stress`

### "I need baseline metrics"

→ [TESTING.md](TESTING.md) → Scenario 1: Single Request Baseline

### "I want to test consistency"

→ [TESTING.md](TESTING.md) → Scenario 2: Sequential Requests

### "I need to stress test"

→ [TESTING.md](TESTING.md) → Scenario 3: Request Batching

### "I want all the options"

→ [README.md](README.md) → Command-Line Flags section

### "I need troubleshooting help"

→ [QUICKREF.md](QUICKREF.md) → Troubleshooting section
→ [README.md](README.md) → Troubleshooting & Deployment section

### "I want to compare deployments"

→ [TESTING.md](TESTING.md) → Scenario 5: Performance Comparison
→ [README.md](README.md) → Comparing Deployments section

### "I want to understand metrics"

→ [README.md](README.md) → When to Use section
→ [TESTING.md](TESTING.md) → Understanding the Results section

### "I want to integrate into CI/CD"

→ [TESTING.md](TESTING.md) → Integration with CI/CD section

### "I want to monitor continuously"

→ [TESTING.md](TESTING.md) → Real-World Scenario: Continuous Monitoring

---

## 🔗 External Session Guides

Located in: `~/.copilot/session-state/27fe00a1-8a58-48ba-9bc1-ae76f221353c/`

### STRESSTEST_SUMMARY.md

Quick overview of:

- What was created
- Build & run instructions
- Key features
- Example usage
- Understanding deployment

### USAGE_EXAMPLES.md

6+ detailed usage examples:

1. Single Request Baseline
2. Consistency Test (5 sequential)
3. Batch Load Test (20 requests)
4. Timeout Testing
5. Different Endpoint
6. Ramp-up Test (future)

Plus:

- Comparing deployments
- Real-world monitoring scenario
- Common workflows
- Interpretation guides
- Next steps

---

## 🛠️ Build & Run

### Build the tool

```bash
cd /Users/admin/go/src/lab_WorkerPattern
make stresstest
```

### Run default test

```bash
make stress
```

### Run manual test

```bash
./bin/stresstest [FLAGS]
```

See [QUICKREF.md](QUICKREF.md) for all flags.

---

## 📊 Example Commands

```bash
# Baseline
./bin/stresstest -requests=1

# Consistency
./bin/stresstest -requests=5 -timeout=120s

# Batch load
./bin/stresstest -requests=50 -timeout=120s

# With ramp-up
./bin/stresstest -concurrency=4 -requests=100 -rampup=4s

# Short timeout (test error handling)
./bin/stresstest -requests=3 -timeout=30s

# Different endpoint
./bin/stresstest -url http://27.71.229.15:3000/
```

---

## 📖 Reading Order (Recommended)

For **first-time users**:

1. This file (INDEX.md) - You are here ✓
2. [QUICKREF.md](QUICKREF.md) - Overview & quick start
3. [TESTING.md](TESTING.md) → Scenario 1 - Run first test
4. [README.md](README.md) - Deep dive when needed

For **experienced users**:

- [QUICKREF.md](QUICKREF.md) - Cheat sheet
- [README.md](README.md) - Reference

For **troubleshooting**:

- [QUICKREF.md](QUICKREF.md) → Troubleshooting
- [README.md](README.md) → Troubleshooting & Deployment

For **integration**:

- [TESTING.md](TESTING.md) → CI/CD Integration

---

## 💾 File Summary

| File | Lines | Purpose |
| --- | --- | --- |
| INDEX.md | This | Navigation |
| QUICKREF.md | 125 | Quick reference |
| README.md | 295 | Complete docs |
| TESTING.md | 353 | Test scenarios |
| main.go | 241 | Implementation |
| **Total** | **1,014** | Complete stress testing toolkit |

---

## ✨ What You Can Do

✅ Test performance with configurable concurrency  
✅ Measure latency percentiles (P50, P95, P99)  
✅ Calculate throughput (requests/second)  
✅ Track success rates and errors  
✅ Compare deployments  
✅ Test timeout handling  
✅ Ramp-up load gradually  
✅ Monitor trends over time  
✅ Validate scaling behavior  
✅ Integrate into CI/CD pipelines  

---

## 🚀 Getting Started (5 minutes)

1. **Read**: [QUICKREF.md](QUICKREF.md) (2 min)
2. **Build**: `make stresstest` (1 min)
3. **Run**: `make stress` (2 min)
4. **Interpret**: Check results using [QUICKREF.md](QUICKREF.md)

---

## 📞 Questions?

- **"How do I...?"** → Check the Quick Navigation section above
- **"What does this metric mean?"** → [README.md](README.md) → Key Metrics section
- **"What went wrong?"** → [QUICKREF.md](QUICKREF.md) → Troubleshooting
- **"Can I...?"** → [README.md](README.md) → Future Enhancements

---

**Last Updated:** 2026-04-06  
**Status:** ✅ Ready to Use  
**Tested Against:** <http://27.71.229.15:3000> (1 worker, systemd socket activation)

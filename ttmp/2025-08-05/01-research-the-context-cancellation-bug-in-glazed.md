# Research Brief: Context Cancellation Bug in Glazed Framework

**Date:** August 5, 2025  
**Issue:** Signal handling interference causing database connection hangs  
**Status:** Under investigation  

## Executive Summary

During implementation of signal handling for the mento-service worker command, we discovered that adding `signal.NotifyContext()` to the glazed framework's GlazeCommand path causes previously working applications (like sqleton) to hang during database connections, even when the signal-enhanced context is not used.

## Problem Statement

**Symptom:** Applications using GlazeCommand (like sqleton) that previously responded to Ctrl-C now hang indefinitely during PostgreSQL database connections via lib/pq driver.

**Root Cause:** Creating `signal.NotifyContext()` interferes with existing signal handling, causing lib/pq's TCP connection code to become unresponsive to signals.

## Technical Investigation

### Bisection Results

We systematically tested different versions to isolate the exact cause:

| Version | Changes | Sqleton Behavior | Mento-service Behavior |
|---------|---------|------------------|------------------------|
| **A** | Only debug logging, no context wrapping | ✅ Works (Ctrl-C cancels) | ❌ Hangs (no signal handling) |
| **B** | Add `context.WithCancel()` wrapping | ✅ Works (Ctrl-C cancels) | ❌ Hangs (no signal handling) |
| **C** | Add `signal.NotifyContext()` but use `cmd.Context()` | ❌ **Hangs** | ❌ Hangs |
| **D** | Full signal handling (use NotifyContext result) | ❌ Hangs | ✅ Works (Ctrl-C cancels) |

**Key Finding:** Version C proves that **just creating `signal.NotifyContext()` breaks existing signal handling**, even when the result isn't used.

### Signal Handler Interference

The issue appears to be signal handler interference:

1. **Sqleton/Cobra already has signal handling** for graceful shutdown
2. **Creating a second `NotifyContext` for SIGINT/SIGTERM** changes signal delivery behavior
3. **lib/pq's TCP connection code is sensitive** to signal handling state changes
4. **Result:** Previously working signal cancellation stops working

### Network Stack Analysis

The hang occurs deep in the Go network stack during TCP connection establishment:

```
goroutine 1 [IO wait]:
internal/poll.runtime_pollWait(0x7ec88abb85d8, 0x77)
internal/poll.(*pollDesc).waitWrite(...)
net.(*netFD).connect(0xc0001ec300, ...)
net.(*Dialer).DialContext(0xc00006f100, ...)
github.com/lib/pq.defaultDialer.DialContext(...)
```

This suggests the issue is at the **syscall/netpoll level**, where signal handling state affects network I/O behavior.

## Context Propagation Analysis

Our debugging confirmed that context cancellation **is working correctly**:

```
[GLAZED DEBUG] Context cancelled in Classic mode - signal received: context canceled
[CONFIG DEBUG] Context cancelled in Connect() - signal received: context canceled  
[CONFIG DEBUG] PingContext cancelled - reason: context canceled
```

**The issue is not context propagation** - contexts are being cancelled correctly. The problem is that lib/pq's underlying TCP connection code doesn't respect the cancellation when signal handlers have been modified.

## Known Issues with lib/pq

Research confirms lib/pq has documented signal handling issues:

1. **lib/pq is in maintenance mode** - no longer actively developed
2. **GitHub issue #620**: Context cancellation hangs with no network connectivity
3. **TCP connection phase doesn't respect context** when signal handling state changes
4. **pgx driver is recommended replacement** with better context support

## Research Questions for Further Investigation

### 1. Signal Handler Interaction Patterns
- **Question:** What is the exact mechanism by which `signal.NotifyContext()` interferes with existing signal handlers?
- **Research Areas:**
  - Go runtime signal delivery behavior with multiple handlers
  - Interaction between `os/signal.Notify()` and `signal.NotifyContext()`
  - Signal masking and delivery state changes
- **Test Approach:** Create minimal reproduction without database connections

### 2. Network I/O and Signal Handling
- **Question:** Why does signal handler state affect `net.(*netFD).connect()` behavior?
- **Research Areas:**
  - Go netpoll implementation and signal sensitivity
  - Interaction between `runtime_pollWait` and signal handling
  - Linux epoll behavior with signal state changes
- **Test Approach:** Test raw TCP connections with different signal handler configurations

### 3. lib/pq vs Signal State
- **Question:** What specific aspect of lib/pq makes it sensitive to signal handler changes?
- **Research Areas:**
  - lib/pq's signal handling implementation
  - Comparison with pgx driver behavior
  - PostgreSQL wire protocol and signal interaction
- **Test Approach:** Compare lib/pq vs pgx behavior under identical signal conditions

### 4. Context Cancellation Propagation
- **Question:** Are there edge cases where context cancellation doesn't propagate through `NotifyContext` chains?
- **Research Areas:**
  - Context value propagation through signal contexts
  - Deadline and timeout behavior with signal contexts
  - Memory leaks or goroutine leaks with signal contexts
- **Test Approach:** Stress test context cancellation patterns

## Recommended Research Methodology

### Phase 1: Isolated Signal Testing
Create test commands that isolate signal handling behavior:
1. **Basic signal test**: Test `NotifyContext` creation without any I/O
2. **Network I/O test**: Test raw TCP connections with different signal states
3. **Sleep cancellation test**: Test context cancellation with `time.Sleep`

### Phase 2: Driver Comparison
Compare behavior across database drivers:
1. **lib/pq vs pgx**: Same operations with both drivers
2. **Signal state variations**: Test with different signal handler configurations
3. **Context chain variations**: Test different context wrapping patterns

### Phase 3: Glazed Framework Analysis
Analyze glazed's interaction patterns:
1. **Command type differences**: Why GlazeCommand vs BareCommand behave differently
2. **Cobra integration**: How cobra's signal handling interacts with glazed
3. **Framework alternatives**: Test other CLI frameworks for comparison

## Immediate Workarounds

### 1. Selective Signal Handling (Implemented)
- **Approach:** Only add signal handling to command types that need it
- **Implementation:** Keep Classic mode (BareCommand) signal handling, remove GlazeCommand signal handling
- **Trade-off:** GlazeCommand applications rely on existing signal handling

### 2. Driver Replacement (Recommended)
- **Approach:** Replace lib/pq with pgx driver
- **Benefits:** Better context support, active development, signal handling robustness
- **Effort:** Moderate - requires DSN format changes and testing

### 3. Connection Timeout (Immediate)
- **Approach:** Add `connect_timeout=5` to all PostgreSQL DSNs
- **Benefits:** Prevents indefinite hangs
- **Limitation:** Doesn't fix the root cause, may mask other issues

## Success Criteria for Research

1. **Reproduce the issue** in isolation without database connections
2. **Identify the exact signal interference mechanism** causing the behavior change
3. **Determine if this is a Go runtime issue** or application-level problem
4. **Develop a robust solution** that works for all command types
5. **Create regression tests** to prevent future occurrences

## References

- Go source: `os/signal/signal_test.go` - Signal handling behavior documentation
- lib/pq issue #620: Context cancellation with network connectivity issues
- pgx documentation: Modern PostgreSQL driver with better context support
- Mento-service debugging logs: Detailed context cancellation flow analysis

---

**Next Steps:** Implement Phase 1 isolated testing to confirm signal interference hypothesis and develop minimal reproduction cases.

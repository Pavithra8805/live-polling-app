# Development Log

## Redis viewer presence and package compatibility

- **Failure:** The Redis viewer-count implementation initially had no expiry mechanism, so disconnected viewers could remain counted indefinitely after server crashes or network failures.
- **Cause:** A plain Redis set would only remove members on an explicit SSE disconnect, which is not guaranteed for abrupt network loss.
- **Fix:** Use a Redis sorted set with connection-expiry timestamps, remove stale members on join/heartbeat, refresh each connection for 60 seconds, and remove the member explicitly on context cancellation.

## Go dependency/toolchain compatibility

- **Failure:** The backend initially referenced Redis `v9.23.0` and `x/crypto v0.59.0` with a Go 1.25 module floor.
- **Cause:** Those dependencies require Go 1.26.
- **Fix:** Align `go.mod` and the Docker/toolchain requirement with Go 1.26 and record `go.sum`. Preserve the code and lock successful dependency resolution.

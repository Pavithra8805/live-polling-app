# Development Log

## Redis viewer presence and package compatibility

- **Failure:** The Redis viewer-count implementation initially had no expiry mechanism, so disconnected viewers could remain counted indefinitely after server crashes or network failures.
- **Cause:** A plain Redis set would only remove members on an explicit SSE disconnect, which is not guaranteed for abrupt network loss.
- **Fix:** Use a Redis sorted set with connection-expiry timestamps, remove stale members on join/heartbeat, refresh each connection for 60 seconds, and remove the member explicitly on context cancellation.

## Go dependency/toolchain compatibility

- **Failure:** The backend initially referenced Redis `v9.23.0` and `x/crypto v0.59.0` with a Go 1.25 module floor.
- **Cause:** Those dependencies require Go 1.26.
- **Fix:** Align `go.mod` and the Docker/toolchain requirement with Go 1.26 and record `go.sum`. Preserve the code and lock successful dependency resolution.

## Stale second viewer during live voting

- **Failure:** A browser vote updated the first tab, but a second tab remained at the earlier total until the next full-results refetch.
- **Cause:** SSE heartbeat events reused the `results` value captured when the stream opened instead of loading MongoDB/Redis-backed totals again. Pub/Sub vote events could also be partial count payloads.
- **Fix:** Publish a full `PollResults` object after MongoDB persistence succeeds, and make SSE heartbeats fetch fresh results plus the current Redis viewer count before emitting an update.

## Closed poll winner verification

- **Failure:** During local UI verification, the shared browser session was not signed in as the poll owner, so the Dashboard Close button could not be used directly.
- **Cause:** The browser session did not retain the owner's JWT after the earlier login flow.
- **Fix:** Verify the real close endpoint using a short-lived JWT signed with the local JWT secret for the poll owner ID. MongoDB and the API both confirmed `status: closed`; the final tie at 3 votes each is handled as a tied winner set rather than selecting only the first option.

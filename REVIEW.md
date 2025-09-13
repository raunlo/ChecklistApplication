ChecklistApplication — Ready for review
=====================================

Summary
-------
This branch implements moving realtime updates to Server-Sent Events (SSE) and removes the Firebase/FCM client integration from the frontend. Key parts:

- Backend: typed SSE event struct and an in-memory SSE broker that replays a small buffer and streams live events. Broker publishes are non-blocking; slow clients may drop messages.
- Frontend (studio): `use-sse` hook and integration into `use-checklist` for realtime updates. Reordering now sends integer order numbers and clamps to >= 1 to satisfy OpenAPI validation.
- Fixes: invalid React hook usage (static import), removed client-side id-based dedupe per request, ensured Go code formatting.

Files changed (high level)
--------------------------
- internal/sse/event.go           — typed Event struct
- internal/sse/server.go          — SSE broker and HTTP handler
- internal/core/service/*         — publishing SSE events after operations
- studio/src/hooks/use-sse.ts    — SSE client hook
- studio/src/hooks/use-checklist.ts — realtime integration and reorder logic
- studio/src/components/checklist-manager.tsx — fixed hook import

How to build and test locally
-----------------------------
1) Backend (repo root):

   - Compile: go build ./...
   - Vet: go vet ./...
   - Tests: go test ./...

   All three ran successfully in my run.

2) Frontend (studio):

   - Install: npm ci
   - Build (prod): npm run build

   A production build completed successfully in my run (Next.js build output shows prerendered pages).

Formatting & lint
-----------------
- `gofmt` was applied to Go source files.
- Next.js build skipped linting in this environment; if you want ESLint run locally use `npm run lint` inside `studio`.

Notes & outstanding decisions
-----------------------------
- Server-side event ids: currently the broker does not assign monotonic event IDs or set SSE `id:` lines. If you want reliable replay/dedupe via Last-Event-ID, we can add server-assigned ids and client handling. This is optional and will require coordinated backend+frontend changes.
- Order numbering: frontend now sends integer order numbers (neighbor +/-1 strategy) and clamps to >=1 to satisfy OpenAPI schema. If you want fractional midpoints instead, change backend schema and DB to accept decimals.
- Replay guarantees: broker keeps a small in-memory ring buffer and replays to newly connected clients; messages may be dropped for slow clients.

Review checklist (what I validated)
----------------------------------
- [x] Backend compiles (go build ./...)
- [x] go vet completed
- [x] Unit tests for packages with tests passed (`go test ./...`)
- [x] Go sources formatted with `gofmt`
- [x] Frontend production build completed (`npm run build` in `studio`)

Next steps (suggested)
----------------------
1. Decide whether to add server-assigned SSE ids + Last-Event-ID replay.
2. (Optional) Add a lightweight end-to-end smoke test that starts the server and verifies an SSE message is received by a test client.
3. Create a PR from branch `real-time-update` and assign reviewers.

Contact
-------
If anything needs adjusting before the PR (naming, tests, or more comments), tell me what to change and I'll update the branch.

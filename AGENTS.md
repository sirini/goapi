# GOAPI working agreement

## Product context

GOAPI is the Fiber v3 backend for the NUBO community and publishing platform. API behavior must remain suitable for multiple deployment types rather than one site-specific installation.

## Collaboration and workflow

- The product owner defines behavior and performs final QA; Codex critically reviews, implements, tests, commits, and pushes agreed changes.
- Do not merge stabilization or feature branches into `main` until product-owner QA is complete.
- Read `/Users/sirini/github/nubo.git/PROJECT_STATUS.md` at the beginning of work when it exists, and update it at meaningful milestones.
- Check frontend contract effects in `/Users/sirini/github/nubo.git` whenever request, response, cookie, authentication, upload, or download behavior changes.
- Use focused commits for coherent work units. After a successful commit, push the current branch without asking again.
- Run `gofmt` on changed Go files and run relevant tests plus `go test ./...` and `go vet ./...` when practical.
- Add regression tests for security and authentication fixes. Do not weaken authorization checks merely to preserve an invalid request pattern.
- Preserve unrelated user changes and avoid destructive Git operations.

## Current priority

Stabilize cross-board resource authorization, signup/password-reset verification, refresh-token behavior, download-token concurrency, legacy password migration, and blocked-user checks before feature development.

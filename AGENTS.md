# GOAPI working agreement

## Product context

GOAPI is the Fiber v3 backend for the NUBO community and publishing platform. API behavior must remain suitable for multiple deployment types rather than one site-specific installation.

## Collaboration and workflow

- The product owner defines behavior and performs final QA; Codex critically reviews, implements, tests, commits, and pushes agreed changes.
- Do not merge stabilization or feature branches into `main` until product-owner QA is complete.
- Read `docs/PROJECT_STATUS.md` in the sibling NUBO repository at the beginning of work when it exists, and update it at meaningful milestones.
- Use `docs/NUBO_COMMUNITY_OS_ROADMAP.md` in the sibling NUBO repository as adaptable long-term direction when product scope or priority is relevant; do not treat it as a mandatory implementation sequence.
- Check frontend contract effects in the sibling NUBO repository whenever request, response, cookie, authentication, upload, or download behavior changes.
- Use focused commits for coherent work units. After a successful commit, push the current branch without asking again.
- Run `gofmt` on changed Go files and run relevant tests plus `go test ./...` and `go vet ./...` when practical.
- Add regression tests for security and authentication fixes. Do not weaken authorization checks merely to preserve an invalid request pattern.
- Preserve unrelated user changes and avoid destructive Git operations.
- Build the x86-64 Linux binary bundled with NUBO only through `./scripts/build-ubuntu22.sh`; do not replace it with a binary compiled directly on the host OS.

## Current priority

Stabilize cross-board resource authorization, signup/password-reset verification, refresh-token behavior, download-token concurrency, legacy password migration, and blocked-user checks before feature development.

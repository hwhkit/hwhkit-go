## Summary

<1-3 bullets of what & why>

## Why

<problem this solves; link issue / discussion if any>

## Wire shape

<for API changes: a short Go snippet showing the new public surface>

## Test plan

- [ ] `go test ./...` green across all modules in `go.work`
- [ ] `go vet ./...` clean
- [ ] `golangci-lint run ./...` clean
- [ ] `go build ./...` succeeds on all `go.work` members
- [ ] godoc-style doc comments on new exported names
- [ ] CHANGELOG entry under `## [Unreleased]`

## Risk

- [ ] Additive only (no existing callers affected)
- [ ] Behaviour change (call it out; downstream impact?)
- [ ] Breaking change (pre-1.0 minor bump?)

## Out of scope

<things deliberately left for follow-ups, with issue links if open>

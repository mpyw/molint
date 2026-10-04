# Repository instructions

molint is a Go analyzer that enforces samber/mo. Absence is `mo.Option`, not a nil pointer or a trailing `bool`. An Option or a Result never holds or gives a nil it should not. Every rule, message, exemption and directive is specified in [design/rules.md](design/rules.md). The architecture and the rejected designs are in [design/implementation.md](design/implementation.md). They include why this project stopped proving pointers non-nil. Read the relevant section before changing a rule, a message, a directive, or how values are followed.

When a rule changes, change `design/rules.md` first, then its formal spec under `spec/` if it has one, then its fixtures under `testdata/src/`, then the code. Keep the README and `skills/molint-authoring/SKILL.md` in sync with the rules. The skill is installed by `molint skill install`, so a stale one misleads every agent that reads it.

Run `go test ./...` for Go changes and `mise x -- ./test_all.sh` before claiming the full repository gate passes. It runs golangci-lint, `declscope shrink`, declscope, the formal specs, and the coverage floor.

This repository runs declscope on itself with `qualify: ondemand` and `exported: true`. Read `.agents/skills/declscope-authoring/SKILL.md` before adding, naming, or moving a declaration.

# Repository instructions

nilproof is a Go analyzer that reports a returned pointer it cannot prove non-nil. It is strict on purpose: an unproven return is reported, not guessed at. The proof model, its rejected designs, and the test layout are recorded in [implementation notes](design/implementation.md). Read the relevant section before changing the proof, the rule on error results, diagnostics, or directives. The open design questions are tracked in [#1](https://github.com/mpyw/nilproof/issues/1).

Run `go test ./...` for Go changes and `mise x -- ./test_all.sh` before claiming the full repository gate passes. It runs golangci-lint, `declscope shrink`, declscope, and the coverage floor. Keep the README and analyzer behavior in sync.

This repository runs declscope on itself with `qualify: ondemand` and `exported: true`. Read `.agents/skills/declscope-authoring/SKILL.md` before adding, naming, or moving a declaration.

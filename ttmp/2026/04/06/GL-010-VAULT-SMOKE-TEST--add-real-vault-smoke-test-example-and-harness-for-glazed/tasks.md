# Tasks

## Phase 0: Ticket And Planning

- [x] Create the GL-010 ticket workspace for the real Vault smoke-test harness.
- [x] Add the implementation plan document.
- [x] Add the diary document.
- [x] Write the concrete implementation plan, task breakdown, and index summary for the ticket.
- [x] Commit the planning and documentation baseline.

## Phase 1: Example Harness

- [x] Add a new `cmd/examples/vault-smoke-test` example command that exercises real Vault-backed `TypeSecret` hydration.
- [x] Wire `vault-settings` and `command-settings` into the example command so it can prove bootstrap parsing and redacted parsed-field output.
- [x] Implement the main source chain with precedence `defaults -> config -> vault -> env -> args -> cobra`.
- [x] Add clear stdout output that makes resolved values and precedence behavior easy to inspect manually.
- [x] Add a `README.md` in the example directory explaining setup, expected precedence, and manual runs.
- [x] Run focused Go validation for the example package.
- [x] Commit the example harness code and docs.

## Phase 2: Real Smoke Script

- [ ] Add a shell smoke-test script in the example directory that starts `vault server -dev`, seeds secrets, runs the example, and asserts expected behavior.
- [ ] Use `tmux` for the long-running Vault dev server so the script matches repo process-handling conventions.
- [ ] Make the script verify that secret fields hydrate from Vault, non-secret fields do not, env overrides Vault, flags override env, and parsed-field output redacts secrets.
- [ ] Execute the smoke-test script successfully against the local Vault binary.
- [ ] Commit the smoke-test script and any follow-up fixes.

## Phase 3: Ticket Bookkeeping

- [ ] Update the diary with the implementation and validation steps.
- [ ] Relate the new example files to the diary.
- [ ] Update the ticket changelog for the example and smoke-test commits.
- [ ] Re-run `docmgr doctor --ticket GL-010-VAULT-SMOKE-TEST --stale-after 30`.
- [ ] Commit the final ticket bookkeeping updates.

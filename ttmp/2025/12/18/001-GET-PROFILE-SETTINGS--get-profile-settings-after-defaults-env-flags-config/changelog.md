# Changelog

## 2025-12-18

- Initial workspace created


## 2025-12-18

Created ticket + initial analysis and diary docs; related core Cobra/middleware/profile files for investigation.


## 2025-12-18

Wrote Glazed-focused analysis doc detailing Cobra flag registration/parsing and two bootstrap-based approaches to load profiles without circular dependency.


## 2025-12-18

Implemented Option A bootstrap parse + added smoke-test script. Validated precedence matrix (profiles < config < env < flags) and ensured unknown profile fails. Script: scripts/01-smoke-test-simple-inference-profiles.sh


## 2025-12-18

Ticket complete: implemented Option A bootstrap parse for profile-settings (defaults+config+env+flags), added smoke-test script with precedence matrix, and ensured unknown profile fails. Also bumped geppetto watermill to v1.5.1 to match AddConsumerHandler API used by CI typecheck.


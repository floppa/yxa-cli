# Changelog

## v1.2.0
- Added support for merging global and project `yxa.yml` configs.
- Project config now overrides global config for variables and commands with the same name.
- Introduced `YXA_CONFIG` environment variable for specifying config path.

## v1.1.1
- Added LICENSE file.

## v1.1.0
- Fixed and improved documentation.

## v1.0.5
- Incremental workflow improvements for CI/CD (fixing “needs” in workflows).

## v1.0.0
- Fixed documentation build workflow.

## Earlier versions
- Initial implementation of Yxa CLI core features.
- Load and execute commands from `yxa.yml`.
- Variable substitution, .env file, and environment variable support.
- Sequential and parallel command execution.
- Modular internal package structure: `cli`, `config`, `executor`, `variables`, `errors`.
- Initial test suite and documentation setup.
- Added binary for integration tests.
- Fixed interactive prompts for better CLI usability.
- Improved Hugo deployment process for documentation.
- Introduced reusable GitHub workflows.
- Added support for merging global and project `yxa.yml` configs.
- Project config now overrides global config for variables and commands with the same name.
- Added `YXA_CONFIG` environment variable for specifying config path.
- Improved test coverage to >88% across all packages.
- Fixed all lint issues (especially unchecked errors in tests).
- Updated Hugo documentation to reflect new config precedence and merging.
- Added new tests for config merging and resolution logic.
- No breaking changes; all previous workflows remain supported.

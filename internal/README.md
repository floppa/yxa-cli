# Internal Package Structure

This directory contains the internal implementation of the yxa-cli tool. These packages are not meant to be imported by external code.

## Package Structure

- `cli`: Command execution and handling logic
- `config`: Configuration loading and processing
- `errors`: Custom error types
- `executor`: Command execution implementation
- `variables`: Variable resolution and substitution

## Migration Plan

The current implementation uses compatibility layers to maintain backward compatibility with existing code. Once all code has been migrated to use the new internal packages, the compatibility layers can be removed.

Steps to complete migration:

1. Update imports in all files to use the internal packages directly
2. Remove the compatibility layers (files ending with `_compat.go`)
3. Update tests to use the new structure
4. Update the main.go file to use the new structure

This incremental approach ensures that the codebase remains functional during the migration process.

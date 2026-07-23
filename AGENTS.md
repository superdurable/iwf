# Agent instructions

## No backward compatibility (pre-launch)

This repository has **not officially launched**. Treat the surface area as unstable.

- Do **not** preserve backward compatibility when changing or planning APIs, interfaces, protos, SDKs, configs, persistence, or data formats.
- Prefer the cleanest design. **Breaking changes are expected and allowed.**
- Do **not** keep shim layers, dual-path logic, deprecated aliases, or migration adapters “just in case.”
- Do **not** add docs, changelogs, or comments that explain former behavior, migration from old shapes, or “previously X meant Y.”
- When replacing something, delete the old path and update all in-repo callers; do not leave dead compatibility code.

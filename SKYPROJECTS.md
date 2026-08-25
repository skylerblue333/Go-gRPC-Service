# SkyProjects — Wave 2 Slot #126 / Lane 12

**Status:** engineering beta / project-work domain core.

SkyProjects adds a bounded, concurrency-safe project registry to the existing Sky RPC Core repository. It deliberately owns project metadata, lifecycle, membership, and references to external task IDs while leaving task execution and task state to the separate SkyTasks product.

## Domain contract

A project contains:
- validated project and owner identifiers;
- bounded project name;
- lifecycle status: `planned`, `active`, `paused`, `completed`, or `archived`;
- immutable owner membership plus manager/member roles;
- optimistic version checks on mutations;
- deterministic sorted membership and task-reference snapshots;
- bounded references to external task IDs;
- `taskExecutionPerformed: false` on every snapshot.

Only owners can grant manager/member roles. Owners and managers can change lifecycle state or attach task references. Regular members cannot administer the project.

## SKYCOIN4444 integration

Recommended composition:

`SkyOrganizations / SkyTeams -> SkyProjects -> SkyTasks`

A SkyTasks service can use the project ID as a foreign reference and can expose task IDs back to SkyProjects. SkyProjects only stores those IDs as references; it does not execute jobs, schedule work, or claim that referenced tasks exist unless the caller verifies them through SkyTasks.

The package is intentionally added to the existing Go module so the repository's established `go test ./...`, race, vet, vulnerability, build, and container CI gates cover it automatically.

## Security and operational boundaries

The registry is process-local and in-memory. It does not provide durable project persistence, authentication, tenant isolation, distributed coordination, audit-log persistence, file storage, task execution, billing, or production deployment evidence. Callers must authenticate users and enforce tenant/project access at the API or RPC boundary before invoking this domain core.

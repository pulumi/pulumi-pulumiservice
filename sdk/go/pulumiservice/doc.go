// A native Pulumi package for creating and managing Pulumi Cloud constructs.
//
// Resources are organized into two surfaces under one package:
//
// - **v1 (package root)** — Mature, hand-maintained resources accessible directly off the package import (e.g. `pulumiservice.Stack`). **In maintenance mode**: bug fixes and security updates only; no new resources or features. Existing programs continue to work without any code changes.
// - **v2 (`pulumiservice.v2`)** — Actively developed, generated at runtime from the public Pulumi Cloud OpenAPI specification. **Recommended for new programs.** Coverage expands as new operations are mapped from the spec.
//
// Resources from both surfaces can be used in the same program. There is no forced migration: existing users stay on v1 indefinitely, or migrate individual resources to v2 by updating their IaC code (resource type + input shape) and adding Pulumi `aliases` on the new v2 declaration so state rebinds in place. v1 has known coverage gaps relative to the full Cloud API; v2 closes those gaps over time.
package pulumiservice

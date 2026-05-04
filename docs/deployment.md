# Deployment

<!-- TODO: Vollständige Deployment-Doku (Hosting, Ansible, Releases) wird in
     Session 11/12 ergänzt. -->

## Branch Protection

In den GitHub-Repo-Settings für `main` einstellen:

- **Require pull request before merging**
- **Require status checks to pass before merging**, mindestens:
  - `CI Backend / lint`, `CI Backend / test`, `CI Backend / build`
  - `CI Frontend / lint`, `CI Frontend / svelte-check`, `CI Frontend / test`, `CI Frontend / build`
  - `CI PyWorkers / lint`, `CI PyWorkers / typecheck`, `CI PyWorkers / test`
  - `CI Shared / openapi-lint`, `CI Shared / check-generated`, `CI Shared / commitlint`,
    `CI Shared / yaml-lint`, `CI Shared / markdown-links`
- **Require branches to be up to date before merging**
- **Require linear history** (empfohlen für saubere Historie)
- **Require signed commits** (optional, empfohlen)
- **Do not allow bypass for administrators** (für Disziplin)
- **Restrict force pushes**

Die Status-Check-Namen entstehen aus `<workflow-name> / <job-name>`. Erst
nach dem ersten erfolgreichen Lauf eines Workflows kann man die zugehörigen
Status Checks im Branch-Protection-UI auswählen.

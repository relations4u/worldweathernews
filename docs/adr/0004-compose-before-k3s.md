# 4. Docker Compose vor Kubernetes/K3s

Date: 2026-05-03
Status: Accepted

## Context

Self-Hosting auf eigenem Proxmox in der Forschungs-Phase, Maintainer-
Team von genau einer Person. Die Plattform läuft auf zwei VMs
(wwn-prod + wwn-mon), kein Lastenausgleich, kein Auto-Scaling-Bedarf
absehbar. Die Frage: Compose oder gleich K3s/Helm aufsetzen?

## Decision

Wir nutzen **Docker Compose** für lokales Development (Profile-basiert
mit/ohne Monitoring) und **Production** (separate `compose.prod.yml`
auf wwn-prod, eigener Caddy-Stack auf wwn-prod, monitoring-stack auf
wwn-mon — alle via Ansible bzw. shell-script gemanaged).

Kubernetes (K3s) und Helm bleiben **explizit** auf der Roadmap als
Wachstumspfad, kommen aber jetzt nicht.

## Consequences

**Positiv**:

- **Einfacher** mentaler Modell — eine compose.yml beschreibt den
  Stack, ein Volume ist ein Verzeichnis im Dateisystem, ein
  Restart ist ein Ein-Wort-Befehl.
- Niedrige Lernkurve für künftige Mitstreitende — Compose-Wissen ist
  weit verbreitet.
- Schneller Inner-Loop in Dev — `make dev` startet alles in unter 30
  Sekunden, kein Cluster-Boot, keine Helm-Templates.
- Keine zusätzlichen Operatoren oder Sidecars (cert-manager,
  ingress-controller, …) zu pflegen.
- Geringerer Resource-Overhead auf den 8-GB-VMs.
- Backups sind banal — `tar` über das Bind-Mount-Verzeichnis.

**Negativ**:

- **Kein deklaratives Self-Healing** über Pod-Restart hinaus. Wenn
  eine VM crashed, muss ein Mensch sie hochfahren.
- **Kein Multi-Host** ohne Swarm/K3s — derzeit irrelevant.
- **Rolling Updates** sind primitiv (`compose pull && up -d` taucht
  kurz). Für Forschungs-Phase OK, für echte Production später nicht.
- **Secrets-Handling** ist schlechter als K8s-Secrets — gemildert
  durch SOPS+age (siehe ADR-0005).
- **Lifecycle-Komplexität bei mehreren Stacks** — der Caddy-Stack
  läuft separat vom App-Stack, weil `network_mode: host` und eigener
  Update-Cycle. Das ist ein Special-Case, der in K8s mit Ingress-
  Controller eleganter wäre. Akzeptiert für jetzt.

## Migration auf K3s — wenn?

Sinnvoll, wenn mindestens zwei der folgenden Punkte zutreffen:

- ≥ 3 App-Hosts werden gebraucht (Lastenausgleich + Failover)
- Rolling Updates ohne Downtime werden Anforderung
- Mehrere Maintainer mit unterschiedlichen Service-Verantwortungen
- DDoS-Last erfordert horizontale Skalierung

Ansible-Playbooks und SOPS-Workflow sind so geschrieben, dass eine
Migration auf K3s ein neues Inventory + Helm-Charts wäre, ohne den
existierenden Code wegzuwerfen.

## Alternatives Considered

- **K3s direkt** — schlanke K8s-Distribution, einfacher als
  vollständiges K8s. Wäre tragbar, kostet aber Lernkurve und Operations-
  Aufwand, der in dieser Phase nicht zahlt.
- **Nomad** — leichter als K8s, aber kleineres Ökosystem, und HashiCorp-
  License-Drama (BSL) macht Self-Hosting unsicher.
- **Plain systemd-Units pro Service** — kein Container-Isolations-
  Layer, mehr OS-Pflege. Compose ist hier der bessere Trade-off.

# Pre-Session-Checkliste

Was du **vor Session 1** erledigt haben solltest. Plan: 1–2 Stunden, einmalig.

Die Reihenfolge ist sinnvoll — manches baut aufeinander auf. Was du schon hast,
hakst du ab und gehst zum nächsten Punkt.

## Reihenfolge auf einen Blick

Nicht alle Punkte sind streng sequenziell. Die kritische Abhängigkeitskette ist:

1. **Punkte 1–2** (System, Git, SSH, Signing) — alles auf deinem Rechner,
  ohne dass GitHub-Repos existieren müssen
2. **Punkt 3.1** (Org anlegen) — bevor du dem Repo einen Owner geben kannst
3. **Punkt 12** (Repo erzeugen und initialisieren) — der eigentliche Startschuss
4. **Erst danach** sinnvoll: Punkt 3.2 (PAT mit Repo-Auswahl), 3.3
  (Branch-Protection-Rules)

Punkte 4–11 sind parallel zu allem anderen erledigbar — sie betreffen
externe Dienste (DNS, Hosting, Editor, age, Notifications) und brauchen
weder Org noch Repo als Vorbedingung.

Punkt 13 (Claude Code) kommt sinnvollerweise **direkt vor Session 1**,
nachdem das Repo steht.

---

## 1. System-Voraussetzungen

### 1.1 Betriebssystem

- Linux (Ubuntu/Debian/Arch/Fedora) oder macOS
- Bei Windows: WSL2 mit Ubuntu — alle Befehle in der Anleitung gehen davon aus

**Warum:** Docker Compose, mise, Caddy, ansible — alles läuft am sauberesten
auf Unix-likes. Windows nativ erzeugt zu viel Reibung mit Line-Endings,
Path-Längen, Symlinks.

### 1.2 Mindest-Hardware

- 16 GB RAM (8 GB reicht zur Not, dann wird's eng wenn der Monitoring-Stack
mitläuft)
- 30 GB freier Plattenplatz (Container-Images, Postgres-Daten, Logs)
- Internet-Verbindung mit halbwegs Bandbreite (initiale Image-Pulls
sind ~5–8 GB)

### 1.3 Basis-Tools

```bash
# prüfen
git --version       # ≥ 2.40
docker --version    # ≥ 24
docker compose version  # ≥ 2.20 (Plugin, nicht docker-compose v1)
curl --version
ssh -V
```

- Alles vorhanden? → weiter
- Wenn nicht: nachinstallieren (Distro-Paketmanager)

**Falls Docker fehlt:** Auf Linux Docker Engine direkt vom offiziellen Repo
installieren, **nicht** Docker Desktop. Docker Desktop ist auf Linux unnötig
und macht mit File-Mounts mehr Probleme als es löst. Auf macOS: Docker Desktop
oder OrbStack (letzteres deutlich performanter für Compose-Workloads).

---

## 2. Git und Identität

### 2.1 Git-Konfiguration

```bash
git config --global user.name "Dein Name"
git config --global user.email "deine@email.tld"
git config --global init.defaultBranch main
git config --global pull.rebase true
git config --global core.autocrlf input  # auf Linux/macOS
```

- Name und Mail gesetzt
- Default-Branch ist `main`

### 2.2 SSH-Key für Git-Hosting

Brauchst du, um ohne Passwort zu pushen.

```bash
# prüfen, ob schon vorhanden
ls -la ~/.ssh/id_ed25519* 2>/dev/null

# neu erzeugen, falls nichts da
ssh-keygen -t ed25519 -C "deine@email.tld" -f ~/.ssh/id_ed25519
```

- SSH-Key existiert
- Public Key (`~/.ssh/id_ed25519.pub`) bei GitHub hinzugefügt
(Settings → SSH and GPG keys → New SSH key)
- Test: `ssh -T git@github.com` → "Hi !"-Meldung

### 2.3 Commit-Signing (GPG **oder** SSH)

Empfehlung: **SSH-Signing**. Einfacher als GPG, gleicher Key wie für `git push`,
keine zusätzlichen Tools nötig.

```bash
# Git auf SSH-Signing umstellen
git config --global gpg.format ssh
git config --global user.signingkey ~/.ssh/id_ed25519.pub
git config --global commit.gpgsign true
git config --global tag.gpgsign true

# allowed_signers-File für lokale Verifikation
mkdir -p ~/.config/git
echo "deine@email.tld $(cat ~/.ssh/id_ed25519.pub)" > ~/.config/git/allowed_signers
git config --global gpg.ssh.allowedSignersFile ~/.config/git/allowed_signers
```

- Signing aktiv
- Bei GitHub: Public Key zusätzlich als **Signing Key** hinterlegen
(Settings → SSH and GPG keys → New SSH key → Type: "Signing Key")
- Test: `git commit --allow-empty -m "test: verify signing"` und dann
`git log --show-signature -1` zeigt "Good signature"

**Warum überhaupt signieren:** Branch-Protection in Session 8 erfordert
"verified" Commits. Ohne Signing scheitert dein erster PR an dieser Regel.

### 2.4 GitHub CLI (`gh`)

```bash
# Installation: siehe https://cli.github.com
gh --version
gh auth login
# → GitHub.com → SSH → Key auswählen → Browser-Login
```

- `gh auth status` zeigt aktive Authentifizierung
- `gh repo list` liefert Ergebnisse

**Brauchst du für:** Repo-Erzeugung in Session 0, später für PR-Workflows
und Release-Trigger.

---

## 3. GitHub-Account und Organisation

### 3.1 Organisations-Entscheidung

- Entscheidung: persönlicher Account oder GitHub Organization?

**Empfehlung:** Organization, auch wenn du Solo bist.

- Trennt private von projekt-bezogenen Repos sauber
- Ermöglicht später Mitarbeit ohne Repo-Transfer (verliert Stars, Issues
bleiben aber erhalten — Transfer ist ärgerlich)
- ghcr.io-Container landen unter `ghcr.io/relations4u/wwn-backend` statt
`ghcr.io/<dein-name>/...` — fühlt sich professioneller an, falls die
Plattform mal öffentlich wird

```bash
# Organisation anlegen (im Browser):
# https://github.com/account/organizations/new → Free Plan
```

- Organisations-Name notiert: `__________________`

Diesen Namen brauchst du in `CLAUDE.md` als `relations4u`-Ersatz und in mehreren
Sessions (Modulpfade, Image-Namen).

### 3.2 GitHub Container Registry vorbereiten

ghcr.io braucht keine separate Aktivierung — beim ersten Push entsteht das
Package automatisch. Du musst hier vor Session 1 **nichts** anlegen.

**Was ggf. später nötig wird** (frühestens nach Session 9, also nach dem
ersten Release-Build):

- Default-Visibility neuer Packages auf "Private" stellen — geht erst,
wenn die Org existiert. Pfad: Org-Profil → **Settings** → in der linken
Sidebar **Packages** → "Default package visibility" auf Private.
- Personal Access Token (PAT) für lokale `docker pull` von ghcr.io —
brauchst du nur, wenn du Container-Images vom Laptop ziehen willst.
GitHub Actions selbst braucht keinen, das nutzt `GITHUB_TOKEN` automatisch.
Server in Session 11 bekommen einen Deploy-Token per SOPS-Secret.

**Wenn du den PAT später wirklich brauchst:**

Pfad: Profilbild oben rechts → **Settings** → in der linken Sidebar ganz
nach unten scrollen → **Developer settings** (eigene Sektion am Ende) →
**Personal access tokens** → **Fine-grained tokens**.

Direkt-URL: `https://github.com/settings/personal-access-tokens`

Token-Konfiguration:

- Name: `wwn-ghcr-local-pull`
- Resource owner: deine Organization
- Expiration: 90 Tage
- Repository access: nur `relations4u/worldweathernews`
- Account permissions → **Packages: Read-only**

Sichern und einloggen:

```bash
mkdir -p ~/.config/wwn && chmod 700 ~/.config/wwn
echo "ghp_..." > ~/.config/wwn/ghcr-token
chmod 600 ~/.config/wwn/ghcr-token

cat ~/.config/wwn/ghcr-token | \
  docker login ghcr.io -u <dein-github-username> --password-stdin
```

**Vor Session 1 hier nichts zu tun** — nur das Bewusstsein, dass es diesen
Punkt gibt, wenn er später relevant wird.

### 3.3 Branch-Protection — nur Konzept, **nicht** jetzt aktivieren

Wird nach Session 8 (CI-Setup) aktiviert. Wenn du Branch-Protection jetzt
schon einschaltest, blockierst du dir den ersten Commit-Push aus Session 1.

Notiere dir nur die Regeln, die du später setzen willst, dann hast du sie
nach Session 8 griffbereit:

- `main`-Branch:
  - Require a pull request before merging
  - Require approvals: 1 (bei Solo-Repo: Self-Approval erlauben oder
  auf 0 setzen — beides legitim)
  - Require status checks: `ci-backend`, `ci-frontend`, `ci-pyworkers`,
  `ci-shared`
  - Require signed commits
  - Require linear history (kein Merge-Commit-Wahn)
  - Block force-pushes
- Regeln gelesen, Aktivierung später
- Setting-Pfad merken: Repo → Settings → Branches → Add branch
protection rule (geht erst nach Repo-Erzeugung in Punkt 12)

---

## 4. Domain und DNS

### 4.1 Domain-Status

- `worldweathernews.com` ist registriert
- Du hast Zugang zum DNS-Provider (Registrar oder externer DNS-Anbieter
wie Cloudflare, Hetzner DNS, deSEC)

### 4.2 DNS-Strategie entscheiden

**Empfehlung: Cloudflare als DNS-Anbieter**, auch ohne Proxy. Schnelle
Propagation, kostenloses DNSSEC, später optional Proxy/CDN aktivierbar.

- DNS-Anbieter: `__________________`
- Subdomains geplant (Session 11 wird das aufgreifen):
  - `worldweathernews.com` → Frontend
  - `www.worldweathernews.com` → Frontend (Redirect auf apex)
  - `api.worldweathernews.com` → Backend
  - `staging.worldweathernews.com` → Staging-Frontend
  - `api.staging.worldweathernews.com` → Staging-Backend
- Optional jetzt schon: `status.worldweathernews.com` für Uptime-Kuma
später

**Noch nichts in DNS eintragen** — die IPs gibt es erst nach Terraform-Apply
in Session 11. Nur strukturell entscheiden.

### 4.3 Email-Adressen

Für Domain-Validierung (Let's Encrypt), Abuse-Kontakt, Maintainer-Mail:

- `admin@worldweathernews.com` oder `hostmaster@...` als Forward
eingerichtet (oder echtes Postfach beim Provider)
- Diese Adresse wird Caddy als ACME-Account-Mail verwenden

---

## 5. Hosting-Provider

Brauchst du erst für Session 11, aber Account-Erstellung hat manchmal
Verifikations-Wartezeiten — also jetzt parallel anstoßen.

### 5.1 Hosting-Entscheidung

**Empfehlung für dich** (Solo, DSGVO-Fokus, Self-Hosting-Gedanke,
Standort Brandenburg): **Hetzner Cloud**, Region Falkenstein oder Nürnberg.

Vorteile:

- DE-Standort, AVV problemlos
- Sehr günstig: CX22 (4 GB RAM, 2 vCPU) für ~4 €/Monat reicht für Staging
- API gut, Terraform-Provider stabil und gepflegt
- Deutscher Support

Alternativen, falls relevant:

- **Strato/IONOS:** Wenn schon Bestandskunde
- **Eigene Hardware** (Homeserver, Hetzner-Dedicated): bei Datenmengen
die Cloud-Preise sprengen
- **Scaleway/OVHcloud:** EU-basiert, ähnlich wie Hetzner
- Entscheidung: `__________________`

### 5.2 Account vorbereiten

- Account angelegt
- Zahlungsmethode hinterlegt (Hetzner: SEPA möglich)
- **Zwei-Faktor-Auth aktiviert** (Pflicht aus Sicherheitsgründen)
- API-Token erzeugt mit Read+Write-Scope, sicher abgelegt
(z. B. in Passwort-Manager, **nicht** in einer Textdatei im Klartext)

### 5.3 SSH-Key beim Provider hinterlegen

- Selbst-erzeugter SSH-Key (`~/.ssh/id_ed25519.pub`) im Provider-Webinterface
als "Default Key" hinterlegt — Terraform/Cloud-Init nutzt den, damit du
direkt auf neue VMs kommst

---

## 6. Secret-Management vorbereiten

### 6.1 age-Schlüsselpaar für SOPS

Brauchst du erst für Session 11, aber das Generieren ist trivial und du
hast den Key dann griffbereit.

```bash
# age installieren (Hetzner-Linux: apt install age, macOS: brew install age)
age-keygen -o ~/.config/sops/age/keys.txt
chmod 600 ~/.config/sops/age/keys.txt

# Public Key auslesen (das ist was später in .sops.yaml landet)
grep "public key:" ~/.config/sops/age/keys.txt
```

- age-Key erzeugt
- Private Key (`keys.txt`) **außerhalb** des Repos abgelegt
- Backup des Private Keys an sicherem Ort:
  - Verschlüsselter USB-Stick **oder**
  - Passwort-Manager (1Password, Bitwarden, KeePassXC)
  - **Nicht** in iCloud/Drive/Dropbox unverschlüsselt
- Public Key notiert: `age1...__________________`

**Warum so paranoid:** Wenn der Private Key verloren geht und du keinen
zweiten Recovery-Key hast, kannst du keine Secrets mehr entschlüsseln.
Die Plattform läuft weiter (Secrets liegen entschlüsselt auf den Servern),
aber Secret-Rotation und neue Deployments sind blockiert.

### 6.2 Recovery-Strategie

- **Mindestens zwei** age-Keys in `.sops.yaml` (in Session 11):
  - Dein Haupt-Key
  - Ein Recovery-Key, der offline und sicher gelagert ist
  (z. B. ausgedruckt im Tresor, oder zweiter Stick beim Vertrauen)
- Optional: ein dritter Key, falls jemand mitarbeitet später

---

## 7. Editor und Entwicklungsumgebung

### 7.1 Editor

Egal welcher, aber:

- EditorConfig-Plugin installiert (Session 1 erzeugt `.editorconfig`)
- Sprach-Server für Go, TypeScript/Svelte, Python aktiv

**Bei VS Code empfohlene Extensions:**

- `golang.go`
- `svelte.svelte-vscode`
- `bradlc.vscode-tailwindcss`
- `ms-python.python` + `charliermarsh.ruff`
- `editorconfig.editorconfig`
- `eamodio.gitlens`
- `ms-azuretools.vscode-docker`
- `redhat.vscode-yaml`
- `tamasfe.even-better-toml`

**Bei Neovim:** Mason oder Lazy als Plugin-Manager, LSP-Configs für die obigen
Sprachen. Wenn du das schon hast, lass es so wie es ist.

### 7.2 Terminal-Setup

- Shell mit anständigem Prompt (zsh + starship, oder fish, oder oh-my-bash)
- `bat`, `eza`/`exa`, `ripgrep`, `fzf`, `jq`, `yq`, `httpie` (oder curl)
installiert — keine Pflicht, aber spart Zeit

### 7.3 Browser

- Chrome/Firefox/Safari, alles fein
- Browser-DevTools können `app.localhost`-Domain auflösen
(siehe nächster Punkt)

### 7.4 `.localhost`-Resolution prüfen

Caddy in Session 3 routet `*.localhost`-Subdomains. Auf den meisten Systemen
funktioniert das automatisch (RFC 6761), aber:

```bash
# Test
ping -c 1 anything.localhost
# Sollte 127.0.0.1 auflösen
```

- Funktioniert? → fertig
- Funktioniert nicht (Linux mit altem nss)? → entweder
  - in `/etc/hosts` eintragen: `127.0.0.1 api.localhost app.localhost`
  - **oder** dnsmasq lokal aufsetzen mit `address=/.localhost/127.0.0.1`

---

## 8. Sprach-Toolchains (vorab nicht nötig — mise macht das)

mise wird in Session 1 alle Tools (Go, Node, Python, pnpm, uv, golangci-lint,
sqlc, goose, …) auf den richtigen Versionen installieren.

**Nicht** schon manuell `apt install golang` oder `brew install python`
ausführen — das produziert Versions-Konflikte mit dem späteren mise-Setup.

- mise selbst installiert: `curl https://mise.run | sh` und in der Shell
aktiviert (Eintrag in `~/.bashrc` / `~/.zshrc`)
- Test: `mise --version`

Ausnahme: Wenn du bereits ein etabliertes Setup mit pyenv/nvm/asdf hast
und das nicht über den Haufen werfen willst, kannst du auch das nutzen —
dann musst du in Session 1 die `.mise.toml`-Anweisung ignorieren und die
Tool-Versionen manuell sicherstellen. Empfehlung: trotzdem mise nehmen,
es koexistiert mit asdf und ist deutlich schneller.

---

## 9. Notification-Setup (optional aber empfohlen)

Damit du mitbekommst wenn CI failt, Deploys schiefgehen, Server-Health
schlecht wird.

### 9.1 Discord-Server **oder** Slack-Workspace

Für späteres Alerting (Grafana, Sentry, GitHub Actions). Pflicht ist nichts,
aber spart später Friktion.

**Empfehlung Discord** für Solo/Small-Scale:

- Kostenlos, Webhooks trivial einzurichten
- Mobile-App brauchbar
- Keine 90-Tage-History-Limits
- Discord-Server angelegt: z. B. `wwn-ops`
- Channels: `#deploys`, `#alerts`, `#security`, `#general`
- Webhook-URL pro Channel notiert (Server-Settings → Integrations → Webhooks)
- **Webhook-URLs als Secret behandeln** — nicht in Repos, nicht in
Screenshots

**Slack-Alternative:** kostenloser Plan reicht für Solo, aber 90-Tage-Message-
History ist schmerzhaft, wenn du Alerts retrospektiv anschauen willst.

### 9.2 Email für Status-Mails

- Eine Adresse, die du wirklich liest, für GitHub-Notifications,
Hetzner-Monitoring-Mails, Domain-Renewal-Reminder. Kein Catch-All
"geht in Spam"-Postfach.

---

## 10. Externe Konten für später

Brauchst du noch nicht in Session 1, aber jetzt-anlegen-und-verifizieren spart
Wartezeit später. Alles optional bis zur jeweiligen Session.

### 10.1 Sentry (Session 10 oder später)

- Account auf sentry.io (Free-Plan reicht für Anfang)
- **Alternative:** GlitchTip selbst hosten (in Session 10 ergänzbar);
dann hier nichts vorbereiten

### 10.2 Cloudflare (Session 11 / 12)

- Account angelegt (kostenloser Plan)
- Domain `worldweathernews.com` als Site hinzufügen, Nameserver am
Registrar umstellen — **erst tun**, wenn DNS-Records klar sind,
sonst ist die Site offline während der Propagation

### 10.3 Plausible Analytics (irgendwann später)

- Self-Hosting geplant: kein externer Account nötig
- Hosted: Account auf plausible.io (kostenpflichtig nach Trial)

### 10.4 Mailversand-Provider (Feature-Phase)

Brauchst du erst, wenn die Plattform User-Mails verschicken soll
(Registrierung, Passwort-Reset, Wetter-Alerts). Nicht jetzt entscheiden,
nur Optionen kennen:

- **Postmark, Mailgun, SES:** für transactional Mail
- **Brevo, Listmonk (self-hosted):** für Newsletter
- **MailHog:** lokal in Compose, nur Development

---

## 11. Maintainer-Hygiene

### 11.1 Passwort-Manager

- Passwort-Manager im Einsatz: 1Password, Bitwarden, KeePassXC, …
- Alle Provider-Logins drin (GitHub, Hetzner, DNS, Cloudflare, …)
- Recovery-Codes für 2FA gespeichert

### 11.2 Backup-Strategie schon mal gedacht

Pflicht erst nach Production-Deploy, aber:

- Plan im Kopf: Was wird gebackuppt (Postgres-Dumps, User-Uploads, Configs),
wohin (S3-kompatibel: Hetzner Storage Box, Backblaze B2, BorgBase),
wie oft, wie lange retainen
- Wird in Session 12 (Runbook) explizit gemacht — jetzt nur Bewusstsein

### 11.3 Zeitbudget realistisch einplanen

Sessions 1–12 haben grob 20–25 Stunden Maintainer-Reviewzeit (nicht Claude-
Wartezeit). Plus Pausen zwischen Sessions, plus deinen normalen Job, plus
Lernkurve bei Tools, die du noch nie genutzt hast.

- Nicht „in einer Woche mache ich die DevOps-Pipeline" planen.
Realistisch: 2–4 Wochen, mit Pausen für Sacken-Lassen
- Nach Phase A (Sessions 1–2) und Phase B (3–6) je 1–2 Tage Pause
einplanen — zum Reflektieren, was du eigentlich gebaut hast

---

## 12. Repository-Initialisierung (direkt vor Session 1)

Wenn die Punkte 1–11 abgehakt sind, ist das hier der letzte Schritt
**vor** dem ersten Claude-Code-Aufruf:

```bash
# 1. Verzeichnis anlegen
mkdir -p ~/projects/worldweathernews
cd ~/projects/worldweathernews

# 2. Git initialisieren
git init -b main

# 3. Remote anlegen via gh CLI
gh repo create relations4u/worldweathernews \
  --private \
  --source=. \
  --remote=origin \
  --description="Global weather and climate community platform"

# 4. CLAUDE.md aus dem Übergabe-Paket ins Repo kopieren
cp /pfad/zu/wwn-handover/CLAUDE.md ./CLAUDE.md

# 5. sessions/ ins Repo kopieren
cp -r /pfad/zu/wwn-handover/sessions ./

# 6. Pre-Session-Checkliste auch ins Repo (für Nachfolge-Maintainer)
cp /pfad/zu/wwn-handover/pre-session-checklist.md ./

# 7. relations4u in CLAUDE.md ersetzen — damit Claude Code in Session 1
#    direkt den richtigen Modulpfad nutzt
sed -i.bak "s|<org>|relations4u|g" CLAUDE.md && rm CLAUDE.md.bak
# Bei macOS: sed -i '' "s|<org>|<dein-org-name>|g" CLAUDE.md

# 8. Minimale .gitignore (siehe Strategie-Plan)
cat > .gitignore <<'EOF'
.env
.env.local
*.sops.decrypted
node_modules/
.svelte-kit/
build/
dist/
__pycache__/
*.pyc
.venv/
bin/
tmp/
.DS_Store
*.swp
.idea/
.vscode/
!.vscode/settings.json
EOF

# 9. Initial-Commit
git add .
git commit -m "chore: initial commit with handover documents"
git push -u origin main
```

- Repo auf GitHub angelegt und sichtbar
- CLAUDE.md, sessions/, pre-session-checklist.md, .gitignore committet
- `<org>` in CLAUDE.md ersetzt
- Erster Push erfolgreich
- Branch-Protection für `main` **noch nicht** aktivieren — sonst kommst
du bei Session 1 nicht mehr weiter ohne PR-Workflow. Aktivieren in
Session 8 nach CI-Setup.

---

## 13. Claude Code installieren und konfigurieren

```bash
# Installation
npm install -g @anthropic-ai/claude-code

# Erstes Login (Browser öffnet sich)
cd ~/projects/worldweathernews
claude
# → Mit Anthropic-Account einloggen
```

### 13.1 Empfohlene Konfiguration

In `~/.claude/settings.json` (wird beim ersten Start angelegt):

```json
{
  "autoAcceptEdits": false,
  "verboseToolUse": true
}
```

- `autoAcceptEdits: false` — kritisch, sonst editiert Claude Code Files
ohne dass du sie siehst
- Plan-Mode-Shortcut bekannt: **Shift+Tab** im Prompt-Eingabefeld
- Slash-Commands bekannt:
  - `/clear` — Context leeren (zwischen Sessions)
  - `/compact` — Context komprimieren wenn lang
  - `/cost` — Token-Kosten der Session anzeigen
  - `/exit` — Session beenden
  - `/help` — alle Commands

### 13.2 Erstes „Hello World" mit Claude Code

Vor Session 1, einfach um zu prüfen dass alles läuft:

```
> Lies CLAUDE.md und fasse in 3 Sätzen zusammen, was das Projekt ist
  und welche Tech-Stack-Entscheidungen getroffen wurden. Mache nichts
  außer dieser Zusammenfassung.
```

- Claude Code antwortet sinnvoll → alles bereit
- Claude Code antwortet komisch / liest CLAUDE.md nicht → prüfen
ob du im richtigen Verzeichnis bist (`pwd`)

---

## 14. Letzter Sanity-Check

Bevor du Session 1 startest, einmal durchgehen:

- Repo ist initialisiert und auf GitHub sichtbar
- CLAUDE.md im Repo-Root, mit deinem `<org>`-Namen
- sessions/ im Repo-Root mit step01.md … step12.md
- Git-Commits werden signiert (`git log --show-signature -1` zeigt Good Sig)
- mise ist installiert (`mise --version`)
- Docker läuft (`docker ps` ohne Fehler)
- Claude Code ist installiert und im Repo-Verzeichnis lauffähig
- `autoAcceptEdits` ist aus
- age-Public-Key liegt griffbereit für Session 11
- Hetzner-Account (oder andere Provider-Wahl) angelegt für Session 11
- Discord/Slack-Webhooks angelegt für spätere Alerts
- Du hast mindestens 2 Stunden zusammenhängende Zeit für Session 1
eingeplant (1–2h Aufgabe + Puffer)

---

## 15. Optional: Was du später bereuen wirst, wenn du es jetzt nicht machst

Punkte aus Erfahrung mit ähnlichen Projekten. Nicht Pflicht, aber gut:

### 15.1 Domain-Privacy aktivieren

- Beim Registrar: WHOIS-Privacy aktiv. Sonst steht deine Privatadresse
öffentlich im WHOIS und du bekommst Spam und Cold-Mails von SEO-Agenturen
bis ans Lebensende.

### 15.2 GitHub-Account härten

- 2FA auf GitHub aktiv (Hardware-Key bevorzugt: YubiKey, Solokey)
- „Vigilant mode" für Commit-Verification aktiv
- Personal Access Tokens haben Ablaufdatum (max. 90 Tage)

### 15.3 Lokales Repo-Verzeichnis hat Backup

- `~/projects/` ist im Backup-Cycle deines Systems
(Time Machine, BorgBackup, restic, Snapshots, …)
- Vor allem: `~/.config/sops/age/keys.txt` und `~/.ssh/id_ed25519` sind
gesichert (verschlüsselt) — wenn die weg sind, ist viel Arbeit weg

### 15.4 Read-Liste anlegen

Bei den ersten Sessions wirst du auf Tools stoßen, die du nicht im Detail
kennst (Caddy-Konfig-Sprache, Chi-Middleware-Konzepte, sqlc-Codegen-Modi,
Ansible-Inventory-Pattern). Lege jetzt schon ein Bookmark/Notion/Obsidian-
Sammelort an für „muss ich nochmal vertiefen". Spart Frustration in den
Sessions, wenn du Claude Code etwas erklären lassen willst, statt es zu
googlen.

---

## Du bist bereit, wenn …

- Du **Punkte 1–4 sowie 12–14 vollständig** abgehakt hast
- Punkte 5–11 zumindest vorgedacht und in groben Zügen angegangen sind

Punkte 5 (Hosting) und 6 (age) brauchst du erst spät — aber das frühzeitige
Anlegen der Accounts vermeidet Wartezeiten mitten im Flow.

Wenn du an einem Punkt stecken bleibst (z. B. SSH-Key-Setup auf einem
exotischen System): kurze Recherche, oder zurückkommen mit konkreter Frage.
Diese Checkliste ist nicht „mach alles ohne nachzudenken", sondern „kenne
diese Themen, bevor du Session 1 startest, und entscheide bewusst".

---

## Stichworte für eigenes Recherche-Briefing

Falls du in einzelnen Themen tiefer einsteigen willst:

- **SSH-Commit-Signing**: GitHub Docs „About commit signature verification"
- **age + SOPS**: getsops/sops Repo-README, FiloSottile/age Repo-README
- **Hetzner Cloud + Terraform**: registry.terraform.io/providers/hetznercloud/hcloud
- **mise**: mise.jdx.dev (besonders „Migrating from asdf")
- **Conventional Commits**: conventionalcommits.org/v1.0.0
- **MADR (ADR-Format)**: adr.github.io/madr
- **Caddy**: caddyserver.com/docs/getting-started

Keine Links als URL hier eingebaut, weil die brechen — Suchbegriffe sind
stabiler als URLs.

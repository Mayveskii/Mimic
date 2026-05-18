```yaml
repo: Mayveskii/netbootxyz
url: https://github.com/Mayveskii/netbootxyz
language: Ansible/Jinja2
status: partial
last_sync: "2025-05-17"

description: |
  Fork of netbootxyz/netboot.xyz (8K+ stars). PXE boot infrastructure using Ansible-driven
  template generation for 200+ OS images. iPXE menu system rendered from Jinja2 templates
  with multi-target build matrix, config-driven endpoint registry, and分层 OS classification
  (live, installer, utility, manual).

advantages:
  - id: nb_template_driven_generation
    what: Jinja2 templates generate iPXE menus for 200+ OS images from structured YAML defaults; single source of truth for menu entries
    evidence: "roles/netbootxyz/templates/menu/ — *.ipxe.j2 templates (disks, linux, utilities); roles/netbootxyz/defaults/main.yml — OS endpoint definitions"

  - id: nb_task_pipeline_ansible
    what: Ansible task pipeline: generate endpoints → generate menus → generate disks → download checksums; idempotent with retry
    evidence: "roles/netbootxyz/tasks/generate_endpoints.yml → generate_menus.yml → generate_disks.yml → main.yml orchestration"

  - id: nb_multi_target_build_matrix
    what: Build matrix across release channels (stable/LTS/rolling), architectures (x86/amd64/arm64), and firmware types (BIOS/UEFI) with conditional template rendering
    evidence: "roles/netbootxyz/defaults/main.yml — release_type per OS; templates/menu/disks.ipxe.j2 — BIOS vs UEFI conditional; allow_legacy boolean"

  - id: nb_config_driven_registry
    what: YAML-driven endpoint registry: each OS defines release, checksums, mirror URLs, kernel params; changes = edit YAML, not templates
    evidence: "roles/netbootxyz/defaults/main.yml — netbootxyz_endpoints dict with 200+ entries; checksums auto-fetched via URI module"

applications:
  - advantage_id: nb_template_driven_generation
    implemented_in: internal/tool/template.go
    mechanism: "YAML defaults → Go template engine → iPXE menu output; template = single source, data = config"
    invariant: "Every menu entry rendered from exactly one template + one config entry. No hardcoded URLs in templates."
    status: planned

  - advantage_id: nb_task_pipeline_ansible
    implemented_in: internal/orchestrator/pipeline.go
    mechanism: "Sequential task pipeline: generate_registry → generate_menus → generate_disks → download_checksums; each step idempotent"
    invariant: "Pipeline is idempotent — re-run produces same output. Failed step → retry, not restart."
    status: planned

  - advantage_id: nb_multi_target_build_matrix
    implemented_in: internal/orchestrator/matrix.go
    mechanism: "Build matrix: OS × arch × firmware → conditional template rendering; filter applicable targets before generation"
    invariant: "Every combination produces valid iPXE script or is explicitly excluded. No partial outputs."
    status: planned

  - advantage_id: nb_config_driven_registry
    implemented_in: data/registry/
    mechanism: "YAML endpoint registry → parse → validate required fields (release, checksum, mirror) → generate config structs"
    invariant: "No endpoint without release + checksum + mirror_url. Invalid entries rejected at parse time."
    status: planned

control:
  - advantage_id: nb_template_driven_generation
    verification: "Unit test: add new OS to YAML defaults → verify menu entry appears in generated iPXE output"
    update_trigger: "Re-analyze when netbootxyz releases new version"
    last_verified: never

  - advantage_id: nb_task_pipeline_ansible
    verification: "Integration test: run pipeline twice → verify idempotent output; fail step 2 → verify retry succeeds"
    update_trigger: "Re-analyze when netbootxyz releases new version"
    last_verified: never

  - advantage_id: nb_multi_target_build_matrix
    verification: "Unit test: x86_64 UEFI + arm64 BIOS combinations → verify correct conditional rendering per target"
    update_trigger: "Re-analyze when netbootxyz releases new version"
    last_verified: never

  - advantage_id: nb_config_driven_registry
    verification: "Unit test: missing mirror_url → verify rejected; valid entry → verify parsed into config struct"
    update_trigger: "Re-analyze when netbootxyz releases new version"
    last_verified: never
```

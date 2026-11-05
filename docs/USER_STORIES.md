# User Stories

## Identity and workspace
- US-001 P0: As a user, I want to register and log in so that my projects are protected.
- US-002 P0: As a user, I want to create a workspace so that I can group security projects.
- US-003 P0: As an owner, I want to invite members and assign roles so that access is controlled.
- US-004 P0: As a member, I must not access another workspace's resources.

## Project and asset
- US-005 P0: As an engineer, I want to create a project to organize an assessment.
- US-006 P0: As an engineer, I want to register domains, IPs, and URLs as assets.
- US-007 P0: As an engineer, I want to tag assets and record criticality/context.

## Authorization and scope
- US-008 P0: As an authorized tester, I want to record written authorization before scanning.
- US-009 P0: As an approver, I want to review scope and validity so that testing is limited to approved targets.
- US-010 P0: As a system, I must reject targets outside scope or with expired authorization.
- US-011 P0: As an operator, I want private/internal targets denied by default.

## Scan lifecycle
- US-012 P0: As an engineer, I want to select an approved scan profile.
- US-013 P0: As an engineer, I want scans to run asynchronously so that the UI remains responsive.
- US-014 P0: As an engineer, I want to view scan status and errors.
- US-015 P0: As an engineer, I want to cancel a scan safely.
- US-016 P0: As an operator, I want resource limits so scans cannot exhaust the host.

## Findings and remediation
- US-017 P0: As an engineer, I want normalized findings from different scanners.
- US-018 P0: As an engineer, I want duplicate observations grouped.
- US-019 P0: As a developer, I want clear evidence and remediation guidance.
- US-020 P0: As a reviewer, I want to mark findings open, false positive, accepted risk, or resolved.
- US-021 P1: As a lead, I want to assign owners and due dates.

## Reporting and operations
- US-022 P0: As a tester, I want an HTML/JSON report with scope and limitations.
- US-023 P0: As an auditor, I want to see who changed authorization, scans, and findings.
- US-024 P1: As a DevSecOps engineer, I want API tokens and webhooks.
- US-025 P1: As an operator, I want metrics and traces for workers and scans.
- US-026 P2: As a user, I want AI-assisted explanations after MVP.

# ProfitCtl AgentShift Sprint Backlog

Verified: 2026-10-07

Status: Confirmed planning synchronization. All 100 mapped records are in Backlog. No development work, release publication, dates, or new assignments were started by this sync.

Project: **ProfitCtl** (`dc3a0fc2-fd0e-4707-8a6d-3cdc33aa7dc7`). Acceptance source: [Versioned Sprint Plan](SPRINT_PLAN.md).

## What was synchronized

- 10 proposed release/sprint parent groups, 20 stories, and 70 tasks.
- 97 new records and 3 reused records. Every create/update was read back.
- 91 native links verified: 86 blocker links and 5 related references. Four additional blocker requirements use verified description notes; see the exception table below.
- The connector has no sprint/cycle or release fields. Parent issues group each proposed version; these are not scheduled native sprints.
- Existing PCTL-15, PCTL-26, and PCTL-23 retain their prior descriptions, acceptance requirements, metadata, and history. Their planning notes and parent grouping were updated.
- Existing active work remains in its prior state. No existing items were closed or reopened. All 32 records outside the reuse set retain their inspected title, state, priority, and parent.
- The connector returned no issue URLs. The verified PCTL identifiers and UUIDs below support exact lookup in AgentShift.

## Reused work

| Plan reference | AgentShift record | Existing outcome retained |
| --- | --- | --- |
| S02-A1 | PCTL-15 | Condere runtime ledger calibration prototype |
| S04-B | PCTL-26 | Condere service, scaling, and change-impact evidence |
| S09-B | PCTL-23 | Cost anomaly, budget, and policy outputs |

PCTL-19 supplies the existing ledger foundation; PCTL-29 supplies existing system economics. PCTL-21 remains the active variance/explanation prerequisite. PCTL-22, PCTL-24, and PCTL-27 remain separate skill, assessment, and demonstration work; related links preserve those boundaries.

## Dependency readback exceptions

For these four edges, the connector could not confirm a native blocker. The dependent record contains a read-back-verified “Required dependency — connector fallback” note. Check it before pulling work. The local sprint graph remains authoritative for these edges; absence from a native blocker list is not permission to start.

| Prerequisite | Dependent | Handling |
| --- | --- | --- |
| PCTL-23 | PCTL-124 | Required blocker recorded in the dependent description |
| PCTL-23 | PCTL-122 | Required blocker recorded in the dependent description |
| PCTL-21 | PCTL-23 | Required blocker recorded in the dependent description |
| PCTL-21 | PCTL-115 | Required blocker recorded in the dependent description |

Only unverified links created by this sync were removed. The pre-existing relationship that returned a conflict was preserved. No unrelated relationships were removed.

## Verified issue map

### S01 — v0.4.0 — PCTL-36

Acceptance: [S01 sprint plan](SPRINT_PLAN.md#s01). All records below are Backlog.

| Plan reference | AgentShift ID | Parent | Action | UUID |
| --- | --- | --- | --- | --- |
| S01 | PCTL-36 | — | created | `64595952-bd3c-443f-85af-77009b2ad5af` |
| S01-A | PCTL-37 | PCTL-36 | created | `ec15878c-207a-4b94-a0ed-501cf8a64dd3` |
| S01-A1 | PCTL-38 | PCTL-37 | created | `5398c085-fd67-4995-a42b-3ab4ee571299` |
| S01-A2 | PCTL-39 | PCTL-37 | created | `d0d49d77-ec8c-45a2-b34e-15660190961f` |
| S01-A3 | PCTL-40 | PCTL-37 | created | `11bb1d07-5577-4e4b-8281-72a4156e9256` |
| S01-B | PCTL-41 | PCTL-36 | created | `5facad81-8950-4d0f-99b8-69d586a5bebb` |
| S01-B1 | PCTL-42 | PCTL-41 | created | `33319c82-23ab-48a1-ab36-29efb12a35d3` |
| S01-B2 | PCTL-43 | PCTL-41 | created | `1e69af5d-3b48-436f-a926-029e761e53b4` |
| S01-B3 | PCTL-44 | PCTL-41 | created | `66dc719f-bf79-437c-ad19-64743ca7ad54` |
| S01-R | PCTL-45 | PCTL-36 | created | `39a7d0b4-b49a-4392-b406-270c662cfb7c` |

### S02 — v0.5.0 — PCTL-46

Acceptance: [S02 sprint plan](SPRINT_PLAN.md#s02). All records below are Backlog.

| Plan reference | AgentShift ID | Parent | Action | UUID |
| --- | --- | --- | --- | --- |
| S02 | PCTL-46 | — | created | `209a7a7b-25ba-4dcf-b828-b4982b2e5ffb` |
| S02-A | PCTL-47 | PCTL-46 | created | `74123b70-52da-4e96-86cd-8e7251c7a3f7` |
| S02-A1 | PCTL-15 | PCTL-47 | reused | `a404d83b-3047-44ff-932a-3667b30970e2` |
| S02-A2 | PCTL-48 | PCTL-47 | created | `d95b6a88-6e3d-4f3b-8d23-41be06954c2f` |
| S02-A3 | PCTL-49 | PCTL-47 | created | `7d07e941-22b8-4be3-814c-f99f1ef21b4c` |
| S02-B | PCTL-50 | PCTL-46 | created | `6b6f28b4-65e5-42f0-b668-c9b731c41d56` |
| S02-B1 | PCTL-51 | PCTL-50 | created | `2ca8ab02-cb02-4400-a915-e4a8ee455b61` |
| S02-B2 | PCTL-52 | PCTL-50 | created | `ada4e9d4-9ad1-4318-84fe-e1a1d2ed16f1` |
| S02-B3 | PCTL-53 | PCTL-50 | created | `e18e6f74-a4c5-4280-9961-16fb99b1dfbb` |
| S02-R | PCTL-54 | PCTL-46 | created | `f71fdcdf-3cec-4d93-be58-f325aaa328f8` |

### S03 — v0.6.0 — PCTL-55

Acceptance: [S03 sprint plan](SPRINT_PLAN.md#s03). All records below are Backlog.

| Plan reference | AgentShift ID | Parent | Action | UUID |
| --- | --- | --- | --- | --- |
| S03 | PCTL-55 | — | created | `ac10ada1-98a9-4ae8-8eb8-d275bf8ee702` |
| S03-A | PCTL-56 | PCTL-55 | created | `19aa0c08-9687-4ffa-bbe9-46a73f3f0c2b` |
| S03-A1 | PCTL-57 | PCTL-56 | created | `eb68eec2-a8a3-4133-9769-4a145e92588a` |
| S03-A2 | PCTL-58 | PCTL-56 | created | `eb452b72-17b0-4cdb-889e-7d5ca02c563a` |
| S03-A3 | PCTL-59 | PCTL-56 | created | `9b2b26db-925e-492c-a541-9f1d177be370` |
| S03-B | PCTL-60 | PCTL-55 | created | `cec738e2-1bab-447e-8319-ed41c13fea62` |
| S03-B1 | PCTL-61 | PCTL-60 | created | `2e746ad8-04b5-4ea7-8005-a3446fd76516` |
| S03-B2 | PCTL-62 | PCTL-60 | created | `0628223d-b17c-4b04-9aab-634e55d87e20` |
| S03-B3 | PCTL-63 | PCTL-60 | created | `3fec4244-79c9-4443-a2e5-30bea8bb6963` |
| S03-R | PCTL-64 | PCTL-55 | created | `22376303-14c4-4a4f-9c2d-b12eb32f9d1d` |

### S04 — v0.7.0 — PCTL-65

Acceptance: [S04 sprint plan](SPRINT_PLAN.md#s04). All records below are Backlog.

| Plan reference | AgentShift ID | Parent | Action | UUID |
| --- | --- | --- | --- | --- |
| S04 | PCTL-65 | — | created | `b54bdb5a-4330-4bb4-8b18-993823da9d16` |
| S04-A | PCTL-66 | PCTL-65 | created | `1987e2fb-6213-453d-a6db-d9d8f28a1d76` |
| S04-A1 | PCTL-67 | PCTL-66 | created | `0431b60b-9d94-41e3-99e0-ce221ef89e90` |
| S04-A2 | PCTL-68 | PCTL-66 | created | `edf80604-0fe4-451d-a1c1-38ce5dd791d5` |
| S04-A3 | PCTL-69 | PCTL-66 | created | `65ae3ea6-970c-4fb8-8afd-3652137add5e` |
| S04-B | PCTL-26 | PCTL-65 | reused | `e44f7f30-c536-4a7e-81df-9759d19f7ffe` |
| S04-B1 | PCTL-70 | PCTL-26 | created | `8741bfaf-288b-4a92-9f98-c71905f69186` |
| S04-B2 | PCTL-71 | PCTL-26 | created | `b3bb6da7-e0fd-46a8-8442-ee6e140ddd07` |
| S04-B3 | PCTL-72 | PCTL-26 | created | `57f541cc-903a-44ba-a9ac-18afa0784e2b` |
| S04-R | PCTL-73 | PCTL-65 | created | `10f1e0bd-25e0-4e02-8e45-f0390d69ef5a` |

### S05 — v0.8.0 — PCTL-74

Acceptance: [S05 sprint plan](SPRINT_PLAN.md#s05). All records below are Backlog.

| Plan reference | AgentShift ID | Parent | Action | UUID |
| --- | --- | --- | --- | --- |
| S05 | PCTL-74 | — | created | `fc4361cf-cdcb-4b7f-a163-147a04fa120e` |
| S05-A | PCTL-75 | PCTL-74 | created | `61cd4128-f833-4267-b6f7-f3f7c614a7cd` |
| S05-A1 | PCTL-76 | PCTL-75 | created | `f6077ecc-1666-4927-aa07-85773ed2e9a4` |
| S05-A2 | PCTL-77 | PCTL-75 | created | `83d8c3fd-9add-4c3f-858c-662d353672f8` |
| S05-A3 | PCTL-78 | PCTL-75 | created | `e0fdcf56-f686-49a0-b3c5-6f13077279ff` |
| S05-B | PCTL-79 | PCTL-74 | created | `dbfc3844-83d9-4434-b7d6-badc770b7764` |
| S05-B1 | PCTL-80 | PCTL-79 | created | `f23ae218-25b7-4a32-b139-2b05392fdaaa` |
| S05-B2 | PCTL-81 | PCTL-79 | created | `268aa043-a445-4aca-8f79-28d0a1d3c00f` |
| S05-B3 | PCTL-82 | PCTL-79 | created | `b042e170-486e-4fd3-927b-2a5b978afdd1` |
| S05-R | PCTL-83 | PCTL-74 | created | `4754cc86-cb84-44aa-9472-d3838e53b8e8` |

### S06 — v0.9.0 — PCTL-84

Acceptance: [S06 sprint plan](SPRINT_PLAN.md#s06). All records below are Backlog.

| Plan reference | AgentShift ID | Parent | Action | UUID |
| --- | --- | --- | --- | --- |
| S06 | PCTL-84 | — | created | `f73a7420-59bd-4272-a0f8-ba3e43c590ab` |
| S06-A | PCTL-85 | PCTL-84 | created | `d4e406f3-caa4-408a-a953-a496b9ad72a7` |
| S06-A1 | PCTL-86 | PCTL-85 | created | `68f288a1-d1be-45d7-a6e1-c5816551f94b` |
| S06-A2 | PCTL-87 | PCTL-85 | created | `9a285ec5-aab0-4c95-9c95-37bbc3486315` |
| S06-A3 | PCTL-88 | PCTL-85 | created | `78e14adf-1280-4c6b-bda6-679a290b96f1` |
| S06-B | PCTL-89 | PCTL-84 | created | `40a5fce2-91d1-40a9-8935-b8fa4eeabf1f` |
| S06-B1 | PCTL-90 | PCTL-89 | created | `2d4c1508-f496-4977-b0ca-38709667d95b` |
| S06-B2 | PCTL-91 | PCTL-89 | created | `4a89db4f-ef36-4cfe-ad47-859a8dbbdc8c` |
| S06-B3 | PCTL-92 | PCTL-89 | created | `d5104023-8e36-4b25-9c58-04eac9225f65` |
| S06-R | PCTL-93 | PCTL-84 | created | `867e261a-4e0d-4ad8-85b5-c4dd60022dbd` |

### S07 — v0.10.0 — PCTL-94

Acceptance: [S07 sprint plan](SPRINT_PLAN.md#s07). All records below are Backlog.

| Plan reference | AgentShift ID | Parent | Action | UUID |
| --- | --- | --- | --- | --- |
| S07 | PCTL-94 | — | created | `dc674ee4-daf5-403b-a3aa-49b59debe2b9` |
| S07-A | PCTL-95 | PCTL-94 | created | `82f311e6-6163-4578-bf66-5c6a1a6ce650` |
| S07-A1 | PCTL-96 | PCTL-95 | created | `79fb07ca-634f-438b-9985-fdee25b4feb3` |
| S07-A2 | PCTL-97 | PCTL-95 | created | `78ee1c7b-519f-4832-aad6-086197263184` |
| S07-A3 | PCTL-98 | PCTL-95 | created | `fb1e3d7b-2626-4821-8a12-6468121cb349` |
| S07-B | PCTL-99 | PCTL-94 | created | `e2a10737-a299-4ab5-ba32-044c4931e381` |
| S07-B1 | PCTL-100 | PCTL-99 | created | `47bb693c-2ec7-4160-b6d1-53de7e4bca77` |
| S07-B2 | PCTL-101 | PCTL-99 | created | `80b492eb-391a-4309-8091-582b04ce7bfa` |
| S07-B3 | PCTL-102 | PCTL-99 | created | `047a6cd7-6ef1-4d4c-b960-d393ece65fe1` |
| S07-R | PCTL-103 | PCTL-94 | created | `68caa8d8-35b8-418d-8411-7c9606b33c0a` |

### S08 — v0.11.0 — PCTL-104

Acceptance: [S08 sprint plan](SPRINT_PLAN.md#s08). All records below are Backlog.

| Plan reference | AgentShift ID | Parent | Action | UUID |
| --- | --- | --- | --- | --- |
| S08 | PCTL-104 | — | created | `f376da1e-094f-4da7-8f99-330ac3856028` |
| S08-A | PCTL-105 | PCTL-104 | created | `a19fb15e-320b-4aed-9e62-0caebbe5db87` |
| S08-A1 | PCTL-106 | PCTL-105 | created | `4294e20a-1853-4a43-a42d-c1d9dacfb08e` |
| S08-A2 | PCTL-107 | PCTL-105 | created | `c382c7bf-de98-4c94-b60c-c51a7981eefa` |
| S08-A3 | PCTL-108 | PCTL-105 | created | `d2e1cb6e-6025-450c-bb88-3b4764199d8c` |
| S08-B | PCTL-109 | PCTL-104 | created | `97440b1d-706a-450d-9495-f0b1d1b22411` |
| S08-B1 | PCTL-110 | PCTL-109 | created | `672b46f8-c2bd-43b6-8039-00d8a2f736d0` |
| S08-B2 | PCTL-111 | PCTL-109 | created | `78de0316-65e7-4a57-a5fc-7c84c11e8832` |
| S08-B3 | PCTL-112 | PCTL-109 | created | `bf72c8f4-a236-4013-9ff7-2fe717fbeaee` |
| S08-R | PCTL-113 | PCTL-104 | created | `b9f6b1e2-ff25-49af-ac61-13fd2ef52e4b` |

### S09 — v0.12.0 — PCTL-114

Acceptance: [S09 sprint plan](SPRINT_PLAN.md#s09). All records below are Backlog.

| Plan reference | AgentShift ID | Parent | Action | UUID |
| --- | --- | --- | --- | --- |
| S09 | PCTL-114 | — | created | `21f2f8a6-5e33-4a91-8cdb-170bb8f0cd3c` |
| S09-A | PCTL-115 | PCTL-114 | created | `7ec533f1-6c5e-489c-8c4a-74ad38b3f1be` |
| S09-A1 | PCTL-116 | PCTL-115 | created | `573e0ed8-b032-4bf1-9b20-f998f1eb6bb8` |
| S09-A2 | PCTL-117 | PCTL-115 | created | `8ae30ccf-4ced-4ce4-9357-0064c077db8f` |
| S09-A3 | PCTL-118 | PCTL-115 | created | `fe52d4ca-2116-4f63-a24e-e8daf07003d2` |
| S09-B | PCTL-23 | PCTL-114 | reused | `58576892-b963-4b08-9718-cb8ef4990a38` |
| S09-B1 | PCTL-119 | PCTL-23 | created | `68399141-0e3e-47ac-befc-ff396ee0081d` |
| S09-B2 | PCTL-120 | PCTL-23 | created | `48693d96-25ba-4f62-9dc3-8d9971b1dc8d` |
| S09-B3 | PCTL-121 | PCTL-23 | created | `1e358124-32fc-419b-8913-355b317fdb2b` |
| S09-R | PCTL-122 | PCTL-114 | created | `cbe6a00e-5c9e-43b5-927b-329dfd4c28a6` |

### S10 — v1.0.0 — PCTL-123

Acceptance: [S10 sprint plan](SPRINT_PLAN.md#s10). All records below are Backlog.

| Plan reference | AgentShift ID | Parent | Action | UUID |
| --- | --- | --- | --- | --- |
| S10 | PCTL-123 | — | created | `b907218f-1cf5-4100-ad71-6b40ddf28a8f` |
| S10-A | PCTL-124 | PCTL-123 | created | `5b93f907-378e-4e90-8677-9f80dcf96a69` |
| S10-A1 | PCTL-125 | PCTL-124 | created | `6a561568-b5f5-403e-9ea2-e382317d836d` |
| S10-A2 | PCTL-126 | PCTL-124 | created | `51006017-1748-4a14-94ca-cf1eb16b9cfb` |
| S10-A3 | PCTL-127 | PCTL-124 | created | `ad173a28-42d6-4a8d-8865-170eb3364969` |
| S10-B | PCTL-128 | PCTL-123 | created | `5a615279-6a4c-4d78-977d-00549f82fef3` |
| S10-B1 | PCTL-129 | PCTL-128 | created | `ef7af83f-45d0-443f-95a6-eac04efbbb98` |
| S10-B2 | PCTL-130 | PCTL-128 | created | `b492875f-5efb-4485-967a-7d7c280aeba6` |
| S10-B3 | PCTL-131 | PCTL-128 | created | `6160b742-4875-4e5a-89b7-5f6ef45c954f` |
| S10-R | PCTL-132 | PCTL-123 | created | `6f538b84-f9f9-4bf4-bc22-8c0387d676eb` |

## Verified native links

A blocker row means the source must finish before the target is pulled. Tasks also inherit parent story and sprint readiness. Related links supply context; they do not establish completion. The four description fallbacks above are additional required blockers.

| Source | Target | Native relation |
| --- | --- | --- |
| PCTL-67 | PCTL-68 | blocking |
| PCTL-62 | PCTL-63 | blocking |
| PCTL-73 | PCTL-74 | blocking |
| PCTL-66 | PCTL-73 | blocking |
| PCTL-66 | PCTL-26 | blocking |
| PCTL-56 | PCTL-64 | blocking |
| PCTL-56 | PCTL-60 | blocking |
| PCTL-43 | PCTL-44 | blocking |
| PCTL-64 | PCTL-65 | blocking |
| PCTL-81 | PCTL-82 | blocking |
| PCTL-51 | PCTL-52 | blocking |
| PCTL-90 | PCTL-91 | blocking |
| PCTL-61 | PCTL-62 | blocking |
| PCTL-66 | PCTL-29 | related |
| PCTL-42 | PCTL-43 | blocking |
| PCTL-45 | PCTL-46 | blocking |
| PCTL-89 | PCTL-95 | blocking |
| PCTL-89 | PCTL-93 | blocking |
| PCTL-106 | PCTL-107 | blocking |
| PCTL-83 | PCTL-84 | blocking |
| PCTL-100 | PCTL-101 | blocking |
| PCTL-120 | PCTL-121 | blocking |
| PCTL-91 | PCTL-92 | blocking |
| PCTL-126 | PCTL-127 | blocking |
| PCTL-38 | PCTL-39 | blocking |
| PCTL-116 | PCTL-117 | blocking |
| PCTL-128 | PCTL-132 | blocking |
| PCTL-124 | PCTL-128 | blocking |
| PCTL-124 | PCTL-132 | blocking |
| PCTL-41 | PCTL-45 | blocking |
| PCTL-41 | PCTL-47 | blocking |
| PCTL-75 | PCTL-83 | blocking |
| PCTL-75 | PCTL-79 | blocking |
| PCTL-110 | PCTL-111 | blocking |
| PCTL-119 | PCTL-120 | blocking |
| PCTL-103 | PCTL-104 | blocking |
| PCTL-86 | PCTL-87 | blocking |
| PCTL-125 | PCTL-126 | blocking |
| PCTL-50 | PCTL-56 | blocking |
| PCTL-50 | PCTL-54 | blocking |
| PCTL-66 | PCTL-24 | related |
| PCTL-47 | PCTL-50 | blocking |
| PCTL-47 | PCTL-54 | blocking |
| PCTL-111 | PCTL-112 | blocking |
| PCTL-97 | PCTL-98 | blocking |
| PCTL-96 | PCTL-97 | blocking |
| PCTL-115 | PCTL-23 | blocking |
| PCTL-115 | PCTL-122 | blocking |
| PCTL-101 | PCTL-102 | blocking |
| PCTL-95 | PCTL-103 | blocking |
| PCTL-95 | PCTL-99 | blocking |
| PCTL-77 | PCTL-78 | blocking |
| PCTL-72 | PCTL-22 | related |
| PCTL-93 | PCTL-94 | blocking |
| PCTL-70 | PCTL-71 | blocking |
| PCTL-117 | PCTL-118 | blocking |
| PCTL-109 | PCTL-115 | blocking |
| PCTL-109 | PCTL-113 | blocking |
| PCTL-87 | PCTL-88 | blocking |
| PCTL-41 | PCTL-27 | related |
| PCTL-105 | PCTL-109 | blocking |
| PCTL-105 | PCTL-113 | blocking |
| PCTL-15 | PCTL-48 | blocking |
| PCTL-52 | PCTL-53 | blocking |
| PCTL-71 | PCTL-72 | blocking |
| PCTL-130 | PCTL-131 | blocking |
| PCTL-113 | PCTL-114 | blocking |
| PCTL-107 | PCTL-108 | blocking |
| PCTL-122 | PCTL-123 | blocking |
| PCTL-60 | PCTL-66 | blocking |
| PCTL-60 | PCTL-64 | blocking |
| PCTL-39 | PCTL-40 | blocking |
| PCTL-47 | PCTL-19 | related |
| PCTL-85 | PCTL-89 | blocking |
| PCTL-85 | PCTL-93 | blocking |
| PCTL-48 | PCTL-49 | blocking |
| PCTL-79 | PCTL-83 | blocking |
| PCTL-79 | PCTL-85 | blocking |
| PCTL-99 | PCTL-103 | blocking |
| PCTL-99 | PCTL-105 | blocking |
| PCTL-26 | PCTL-73 | blocking |
| PCTL-26 | PCTL-75 | blocking |
| PCTL-58 | PCTL-59 | blocking |
| PCTL-57 | PCTL-58 | blocking |
| PCTL-37 | PCTL-45 | blocking |
| PCTL-37 | PCTL-41 | blocking |
| PCTL-68 | PCTL-69 | blocking |
| PCTL-129 | PCTL-130 | blocking |
| PCTL-80 | PCTL-81 | blocking |
| PCTL-76 | PCTL-77 | blocking |
| PCTL-54 | PCTL-55 | blocking |

## Next pull

Refine S01-A / PCTL-37 and its first task PCTL-38. Confirm ownership and active WIP before starting. The plan does not authorize execution, spending, publication, deployment, or merge.

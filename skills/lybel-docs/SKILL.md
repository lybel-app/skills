---
name: lybel-docs
description: |
  Navigation assistant for Lybel's Confluence knowledge base (space `lybel` at lybel.atlassian.net). Timeless skill: stores no specific data — at every session, first reads the Home page (pageId 164232) of Confluence, which is the source of truth for current taxonomy, aliases, status and page index. Provides only the structure, workflows and templates; real state lives in Confluence via MCP. Use when the user asks to search, create, list or update documentation, processes, partners, advisors, investors, accelerators, roadmap, strategy or any organizational artifact. Triggers (Portuguese, as the team speaks): "onde fica X", "me dá a página de Y", "cria página pra Z", "lista páginas de W", "qual o status de X", "tem doc sobre Y?", "adiciona isso na wiki", "procura no Confluence", "documenta esse processo", "atualiza a página de Q", "adiciona advisor/parceiro/investidor". Always replies in Brazilian Portuguese (pt-BR) with full URLs to lybel.atlassian.net.
allowed-tools: |
  mcp__atlassian__getConfluencePage
  mcp__atlassian__searchConfluenceUsingCql
  mcp__atlassian__getPagesInConfluenceSpace
  mcp__atlassian__getConfluencePageDescendants
  mcp__atlassian__createConfluencePage
  mcp__atlassian__updateConfluencePage
  mcp__atlassian__getConfluenceSpaces
  mcp__atlassian__search
  Bash(./bin/lybel-docs *)
  Read
  Write
---

# Lybel Docs — Confluence Knowledge Base Assistant

## Overview

Skill that connects Claude to Lybel's Confluence (`lybel.atlassian.net`, space key `lybel`) to search, create, list and update documentation in natural language — in Portuguese, without manually opening Confluence.

The skill is **deliberately timeless**: it stores no names, lists or state (advisors, investors, partners, accelerators, page IDs). All of that lives in Confluence and is read fresh in every session starting from the Home. The `reference/` files here are generic fallback, not the source of truth.

## Language rule

**This document is in English for Claude's performance (Claude is trained primarily on English data, and English instructions yield more robust reasoning). However, all user-facing output MUST be in Brazilian Portuguese (pt-BR).**

When you respond to the user:
- Use Brazilian Portuguese
- Match the user's tone (formal/informal as they write)
- Keep page titles, category names, and content IN PORTUGUESE (they exist in Portuguese in Confluence)
- Only technical terms or proper nouns stay in English

## Mandatory bootstrap

In **EVERY new session**, before answering anything about the KB:

1. **Read the Confluence Home:**
   ```
   mcp__atlassian__getConfluencePage(
     cloudId="ab1dada3-b25e-40ad-9dbc-682caeea8d00",
     pageId="164232",
     contentFormat="markdown"
   )
   ```

2. **Use the Home content as source of truth** for:
   - Current taxonomy (categories and sub-structures)
   - "Where do I put X?" decision map
   - Aliases (keywords → pages)
   - Page ID Index (if present)
   - Organization rules

3. **Fall back to the generic reference files** (`reference/taxonomy.md`, `reference/aliases.md`, etc.) **only if the Home is inaccessible**.

This skill is deliberately **timeless**. It stores no specific names (advisors, investors, retailers, accelerators) — everything comes fresh from Confluence each session.

## Reference files

- `reference/bootstrap.md` — principle + detailed bootstrap procedure
- `reference/taxonomy.md` — generic structure of the space (fallback)
- `reference/aliases.md` — generic alias patterns (fallback)
- `reference/templates.md` — formats by page type (partner sheet, meeting notes, ADR, etc.)
- `reference/workflows.md` — standard steps (search, create, update, status)

## Default workflows

### 1. Search — "onde fica X" / "tem doc sobre Y?"

1. Look up in the Page ID Index / aliases from the Home (read during bootstrap).
2. If mapped → `getConfluencePage` directly by `pageId`.
3. If not mapped → `searchConfluenceUsingCql` with `space = "lybel" AND (title ~ "<term>" OR text ~ "<term>")`.
4. Return up to 5 results: `- **Title** — summary (full URL)`.

### 2. Create — "cria página pra Z"

1. Use the Home's "Where do I put X?" map to discover the correct category/parent.
2. Choose the template in `reference/templates.md`.
3. **Confirm with the user** the final title, parent and template before creating.
4. Generate the content:
   - **Preferred:** `./bin/lybel-docs adf <template> <args>` (generates rich ADF with tables, expand, TOC, status macros).
   - **Fallback:** `contentFormat: "markdown"` in the MCP call.
5. `createConfluencePage` with the correct `parentId`. Return the final URL.

### 3. Update — "atualiza a página de X" / "adiciona seção Y"

**Never build ADF by hand, and never use `contentFormat: "markdown"` to update a page with macros** (TOC, Expand, panel). Markdown update flattens macros and silently destroys structure.

Use `./bin/lybel-docs edit` — it reads current ADF, applies a section-level operation, and returns new ADF. Macros outside the touched section are preserved verbatim.

**Workflow** (preferred path, binary available):

1. Fetch the current page in ADF:
   ```
   mcp__atlassian__getConfluencePage(cloudId=..., pageId=..., contentFormat="adf")
   ```
2. Extract the `body` JSON from the response and save to a temp file (e.g. `/tmp/lybel-edit/current.json`).
3. Write the new content as markdown fragment to another temp file (e.g. `/tmp/lybel-edit/fragment.md`). For section replacement, include the heading in the fragment.
4. Run one operation:
   ```
   ./bin/lybel-docs edit -i /tmp/lybel-edit/current.json \
     --append /tmp/lybel-edit/fragment.md > /tmp/lybel-edit/new.json

   ./bin/lybel-docs edit -i /tmp/lybel-edit/current.json \
     --replace-section "📇 Page ID Index" /tmp/lybel-edit/fragment.md > /tmp/lybel-edit/new.json

   ./bin/lybel-docs edit -i /tmp/lybel-edit/current.json \
     --insert-after "🔍 Research" /tmp/lybel-edit/fragment.md > /tmp/lybel-edit/new.json

   ./bin/lybel-docs edit -i /tmp/lybel-edit/current.json \
     --delete-section "TODO antigo" > /tmp/lybel-edit/new.json
   ```
5. Send the result back with `contentFormat: "adf"`:
   ```
   mcp__atlassian__updateConfluencePage(
     cloudId=..., pageId=..., title=<same>, contentFormat="adf",
     body=<contents of /tmp/lybel-edit/new.json>,
     versionMessage="<what changed>"
   )
   ```

**Section matching**: exact case-sensitive match on trimmed heading text. First match wins. Section = heading + all following top-level nodes up to (but not including) the next heading of equal-or-higher level.

**When the binary is NOT available**: fall back to `contentFormat: "markdown"` ONLY if the page has no macros to preserve. If it does (TOC, Expand, panel, status), warn the user that the update may destroy structure and ask them to install the CLI first.

### 4. List — "quais aceleradoras temos" / "lista parceiros"

1. Identify the category via the Home.
2. Use `getPagesInConfluenceSpace` or `getConfluencePageDescendants` on the category parent.
3. Return as bullets ordered by title or status.

### 5. Status — "qual o status de X"

See Workflow 5 in `reference/workflows.md` (labels + properties). Always cite the date of the last update.

### 6. Add relationship — "adiciona advisor/parceiro/investidor X"

1. Verify in the Home which department/category is correct (advisor ≠ investor ≠ commercial partner).
2. Confirm template (Advisor Sheet, Investor Sheet, Partner Sheet).
3. Create under the correct parent. Always confirm location before.

## Editorial patterns for Lybel pages

Every **decision**, **proposal**, **strategy**, or **spec** page must follow two conventions. Index pages, reference pages and glossaries can skip them.

### Pattern 1 — Page header

Start every page with a small header block (quote/callout) containing:

> **Contexto:** (one-line summary of what the page is about)
> **Criado em:** YYYY-MM-DD | **Atualizado em:** YYYY-MM-DD | **Criado por:** [nome]

When updating a page, bump the "Atualizado em" date. On first creation, author comes from the current user (confirm if unknown).

### Pattern 2 — Contexto → Problema → Solução structure

For decision/proposal/strategy pages, organize the body around three sections:

1. **Contexto** — where the page comes from, which project/moment it serves, what bigger scope it sits in
2. **Problema** — what is being solved, what hurts, what constraints exist
3. **Solução (possível)** — proposed approach, options, tradeoffs, concrete scope

See `reference/templates.md` → "Decision / Proposal / Strategy" template for the canonical form.

**Why this matters**: the knowledge base becomes predictable — any reader (human or AI) knows where to find the motivation, the pain, and the plan without re-reading everything. Reduces onboarding cost and review time.

**When creating a page**: apply these patterns by default; deviate only if the page type is clearly an index/reference/glossary.

**When updating an existing page** that violates the patterns: offer the user a refactor alongside the content change — don't silently rewrite structure.

## Tool preferences

- **MCP Atlassian** is priority for everything (search, read, create, update).
- **`./bin/lybel-docs adf`** for rich new pages (tables, expand, TOC, status macros) — only if the binary is installed at `./bin/`.
- **`./bin/lybel-docs edit`** for updating existing pages without destroying macros (append, insert, replace-section, delete-section). See Workflow 3.
- **Fallback:** `contentFormat: "markdown"` when the binary doesn't exist. Limitation: loses macros, so use only for plain-text pages.
- **CQL:** prefer `title ~` before `text ~`. **Always** filter by `space = "lybel"`.
- **Batch:** multiple reads in parallel within the same tool-call block.

## Report style

- Reply in **Brazilian Portuguese (pt-BR)**.
- Full URLs always: `https://lybel.atlassian.net/wiki/spaces/lybel/pages/<id>`.
- Concise — the team includes non-technical people.
- **Confirm exact title and location** (parent + category) before creating any page.
- Listings as bullets: `- **Title** — summary (URL)`.
- If the search is empty, suggest 2-3 variations before giving up.

## Language

**Always respond to the user in Brazilian Portuguese (pt-BR)**, regardless of this document being in English. The user (Lybel team, non-technical) expects Portuguese responses.

## Locked configuration

- **cloudId:** `ab1dada3-b25e-40ad-9dbc-682caeea8d00`
- **Space key:** `lybel`
- **Home page ID:** `164232`
- **Base URL:** `https://lybel.atlassian.net/wiki`

Don't ask the user — pass these values directly to the MCP tools.

# Bootstrap — How the Skill Orients Itself

## Language reminder

**This document is in English for Claude's performance. All user-facing output MUST be in Brazilian Portuguese (pt-BR).** When you reply to the user, always use pt-BR. Page titles, category names and content remain in Portuguese (they exist in Portuguese in Confluence).

## Principle

This skill **does not store Lybel-specific data**. It knows:

- Generic taxonomic structure (6 categories)
- Templates per content type
- Standard workflows
- Organization and tie-breaking rules

The **current data** (who the advisors are today, which accelerators are in progress, which investor is in conversation, pageIds of each categorical parent, target retailers in the active roadmap) lives **exclusively in Lybel's Confluence**. The skill always fetches current state before acting.

---

## Bootstrap procedure

In every new session involving the Lybel KB, Claude must:

1. **Read the Confluence Home:**
   ```
   mcp__atlassian__getConfluencePage(
     cloudId="ab1dada3-b25e-40ad-9dbc-682caeea8d00",
     pageId="164232",
     contentFormat="markdown"
   )
   ```

2. **Extract from the Home:**
   - **"Onde coloco X?" table** — current decision map for routing
   - **"Aliases" section** — keywords → pages (including current proper names: people, companies, vendors)
   - **"Page ID Index" section** (if present) — IDs of structural parents (categories, sub-categories, departments)
   - **Categories with current links** — current state of the 6 categories

3. **Use this data as source of truth** for the rest of the conversation.

---

## Fallback

If the Home is inaccessible (auth error, MCP offline, page deleted), use the static files:

- `taxonomy.md` — generic structure of the 6 categories
- `aliases.md` — generic keyword patterns
- `templates.md` — page formats per type
- `workflows.md` — steps per action

This always works, but **with less precision** (no current state: pageIds may be wrong, proper names unknown, statuses outdated). Warn the user when operating in degraded mode.

---

## Why this design

- **Timeless skill:** never goes stale. If an investor leaves the pipeline, an advisor changes area, a new retailer enters the roadmap — Confluence reflects, the skill needs no update.
- **Public-safe skill:** no specific names in the repo, can be open-sourced.
- **Fork-friendly:** any company can use it by swapping just the Home page (and the cloudId).
- **Single source of truth:** Confluence. No divergence between skill and real KB.

---

## When bootstrap is NOT needed

- Purely conceptual question ("quais são as 6 categorias?", "como funciona o template de advisor?") — can be answered directly from the static files.
- Conversation continuing a previous session that was already bootstrapped.

**For any action that creates, edits, moves or references a specific pageId → bootstrap is mandatory.**

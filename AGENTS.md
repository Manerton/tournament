# AGENTS.md

## 1. General Communication & Language Guidelines
- **Communication & Code:** All conversations, reasoning steps, tool usage, code comments, and documentation must be strictly in **English** (to optimize token usage and maintain clarity).
- **Frontend UI Text:** The end-user site interface, labels, tooltips, placeholders, and error messages shown to users must be **strictly in Russian**.
- **Diffs & Edits:** **Never rewrite entire files in chat.** Apply targeted edits directly via file editing tools. In chat responses, only show and discuss the modified snippets/diffs.

---

## 2. Environment & Package Management (Deno Only)
The frontend project (in `frontend/`) is powered by **Deno**.
- **PROHIBITED:** Never use `npm`, `npx`, `yarn`, or `pnpm` under any circumstances (even if a `package.json` or `package-lock.json` is present in the workspace).
- **Execution:** All frontend commands and scripts must be run exclusively via **Deno** (`deno install`, `deno add`, `deno task`, etc.).
- **Dependencies:** Managed via `frontend/package.json` and `frontend/deno.lock`.
  - Install existing dependencies:
    ```bash
    cd frontend && deno install
    ```
  - Add a runtime package (Deno automatically writes to `package.json`):
    ```bash
    deno install npm:<package>
    ```
  - Add a dev dependency:
    ```bash
    deno install -D npm:<package>
    ```
  - Remove a package:
    ```bash
    deno remove <package>
    # or: deno uninstall <package>
    ```

---

## 3. Security & Secrets
- **NEVER** read, inspect, create, or commit `.env*` files or private key files.
- **NEVER** expose secrets, private credentials, or API tokens inside client-facing variables (e.g., `VITE_*`).

---

## 4. Local Project Skills (`.agents/skills/`)
Local skills are located in `.agents/skills/`. The rules in this `AGENTS.md` hold **absolute priority** over any conflicting instructions inside those skills:

Available skills: `context-management`, `frontend-design`, `golang-patterns`, `incremental-implementation`.

### Mandatory Skill Activation
- When searching the project or planning multi-step tasks, always activate:
  - `context-management`
  - `incremental-implementation`

### Token-Saving Overrides (`context-management`)
- Always inspect signatures and patterns with `grep`/`glob` before opening files. Never read large files blindly.
- **Strict Blacklist (Never read completely or index):**
  - Generated Go files in `backend/internal/` (e.g., sqlc-generated code — inspect only the underlying `.sql` files and `sqlc.yaml` when necessary)
  - Lock files (`deno.lock`, `package-lock.json`, `go.sum`)
  - Binary files and static assets (`frontend/public/`, `frontend/node_modules/`)

### Implementation & Testing Overrides (`incremental-implementation` & `golang-patterns`)
- **No Mandatory Tests:** Ignore requirements to write unit tests or follow strict TDD (`go test`) unless explicitly requested by the user.
- **Verification via Compilation:** Instead of running tests after incremental steps, verify integrity by compiling the affected service:
  - Backend: `cd backend && go build ./...`
  - Frontend: `cd frontend && deno task build`

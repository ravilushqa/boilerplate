# Agent skills

Skills vendored into this repository so every agent and every developer works
from the same guidance. Claude Code loads them automatically; other agents
are pointed here by `AGENTS.md`.

| Skill | Source | License |
|---|---|---|
| `go/` | [spf13/go-skills](https://github.com/spf13/go-skills/tree/main/go), commit `c6ea884` | MIT, see `go/LICENSE` |

Do not edit vendored files: update them from upstream with `make skills-update`
and review the diff before committing.

<p align="center">
  <h1 align="center">planosh</h1>
  <p align="center">
    <strong>Deterministic execution plans for AI coding teams.</strong>
  </p>
  <p align="center">
    <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
    <a href="https://github.com/ASDFGHoney/planosh/stargazers"><img src="https://img.shields.io/github/stars/ASDFGHoney/planosh?style=social" alt="GitHub Stars"></a>
    <a href="https://github.com/ASDFGHoney/planosh/discussions"><img src="https://img.shields.io/github/discussions/ASDFGHoney/planosh" alt="Discussions"></a>
    <a href="README.ko.md">한국어</a>
  </p>
</p>

---

> **Developer**: "I'd have to run it through Claude to find out..."
>
> **PM**: Submitted a PR. Please review. &rarr; 40 files changed, **+4,567 -12,300**
>
> If the execution plan had been reviewed first, and that plan was deterministic? No one would need to read a single line of code.

**A plan that isn't deterministic isn't a plan.**

We ran a single `plan.sh` for **16 hours straight** &mdash; zero human intervention &mdash; and it succeeded on the first run. The plan was over-linear and over-verified, so it took longer than it needed to. But it worked. Every step passed. No one touched the keyboard.

That's the promise: **write the plan once, walk away, come back to working code.**

## Why planosh

<table>
<tr>
<td width="25%" align="center"><strong>Deterministic</strong></td>
<td width="25%" align="center"><strong>Reviewable</strong></td>
<td width="25%" align="center"><strong>Asynchronous</strong></td>
<td width="25%" align="center"><strong>Verifiable</strong></td>
</tr>
<tr>
<td>Same plan, same result. Every time.</td>
<td>Review the plan, not the 4,567 lines of generated code.</td>
<td>Run <code>bash plan.sh</code> before leaving work. Done by morning.</td>
<td>Every step has a check. Shell exit codes, not AI judgment.</td>
</tr>
</table>

## What it looks like

```bash
# -- Step 2: Google OAuth --
CURRENT_STEP=2; step 2 "Google OAuth login"
run_claude "
Implement Google OAuth login.
## Build
- User/Account/Session models (prisma)
- NextAuth Google Provider + Prisma Adapter
- /auth/signin page (Google login button)
## Don't build
- Email/password signup
- Profile editing, team features
"
verify "Build succeeds" "npm run build"
verify "Login page exists" "[ -f src/app/auth/signin/page.tsx ]"
checkpoint "feat: Google OAuth login"
```

`run_claude` combines the prompt (`-p`) with a harness (`--append-system-prompt`) to invoke Claude. `verify` judges pass/fail by exit code. That's it.

## How it works

planosh constrains AI execution with 3 layers:

```
+---------------------------------------------------+
| Layer 1: System Prompt (--append-system-prompt)    |
| HOW -- coding conventions, architecture rules,     |
| forbidden patterns                                 |
+---------------------------------------------------+
| Layer 2: User Prompt (-p)                          |
| WHAT -- what to build, what not to build,          |
| preconditions                                      |
+---------------------------------------------------+
| Layer 3: Verification (verify)                     |
| CHECK -- build, file existence, test pass          |
+---------------------------------------------------+
```

When WHAT and HOW are constrained simultaneously, the solution space narrows dramatically. The 5-stage pipeline (spec &rarr; plan &rarr; tasks &rarr; sessions &rarr; code) compresses to 1 stage (prompt &rarr; code).

For the full design, see the [Design Document](docs/DESIGN.md).

## Getting started

planosh is a concept, not a package to install. To start using it:

**1. Write a plan.sh by hand** &mdash; any shell script with `claude -p` calls works.

**2. Or use the reference skills** (Claude Code):

```bash
# Copy skills into your project
cp -r skills/planosh/ .claude/skills/
cp -r skills/planosh-calibrate/ .claude/skills/

# Generate a plan from a PRD
/planosh path/to/prd.md

# Calibrate for determinism (parallel runs, divergence detection)
/planosh-calibrate --runs=3
```

**3. Run it:**

```bash
bash plan.sh
```

## Contributing

planosh is a proposal, not a finished framework. We're building patterns and best practices together. **You don't need to contribute code &mdash; share your plans.**

### Share your plan.sh

The `.plan/` directory in this repo is the community's best practice collection. Contribute yours via PR:

```
.plan/
+-- your-plan-name/
    +-- plan.sh
    +-- harness-global.md
    +-- harness-step-N.md (optional)
    +-- README.md
```

Your README should include:

| Field | What to write |
|---|---|
| **Project** | What you built |
| **Steps** | How many, what each step does |
| **Determinism rate** | Identical results across N runs (e.g., 3/3) |
| **Key findings** | What harness/patterns were effective, what divergence you found |

### Report divergence

"I ran the same plan.sh and got different results" is the best starting point for improving a harness. Report these in [Issues](https://github.com/ASDFGHoney/planosh/issues).

### Discuss patterns

Share your discoveries, experiments, and ideas in [Discussions](https://github.com/ASDFGHoney/planosh/discussions).

### What we hope to see

- Cases where harness + verify loops achieved **100% determinism**
- Patterns that knocked out 50-file, 20-step projects in a single `plan.sh` run
- **Harness templates** for specific stacks (Next.js, Rails, Flutter, ...)
- Verification patterns that guarantee quality without AI judgment
- Harness gaps found through calibration and their fixes

As cases accumulate, patterns emerge. As patterns collect, they become a framework. planosh is the seed.

## In the Wild

Built something with planosh? [Open a PR](https://github.com/ASDFGHoney/planosh/pulls) to add it here.

<!-- 
- [project-name](link) -- brief description, determinism rate
-->

*Your project could be the first.*

## Further reading

- [Design Document](docs/DESIGN.md) &mdash; Problem definition, 3-layer constraint model, calibration loops, full plan.sh example
- [Example PRDs](docs/) &mdash; Retro webapp, C compiler, Markdown-to-slides

## License

[MIT](LICENSE)

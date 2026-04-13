<p align="center">
  <h1 align="center">planosh</h1>
  <p align="center">
    <strong>Determinism first. Everything else follows.</strong>
  </p>
  <p align="center">
    <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
    <a href="https://github.com/ASDFGHoney/planosh/stargazers"><img src="https://img.shields.io/github/stars/ASDFGHoney/planosh?style=social" alt="GitHub Stars"></a>
    <a href="https://github.com/ASDFGHoney/planosh/discussions"><img src="https://img.shields.io/github/discussions/ASDFGHoney/planosh" alt="Discussions"></a>
    <a href="README.ko.md">한국어</a>
  </p>
</p>

---

**The plan-code gap is what actually breaks AI coding teams. Every spec-driven tool claims to close it. None of them do.**

Markdown specs don't turn into code. They turn into *interpretations* of code, and every AI session produces a different one. Worse, **AI defaults to "build more" whenever it meets ambiguity**. Ask for a login and you get MFA, email verification, rate limiting, three abstract layers &mdash; none of which you asked for. **40 files changed, +4,567 -12,300. That's how the PR gets born.**

SDD tools know this. They don't expose it. Over-implementation is the feature that makes vibe coding demos go viral &mdash; "look, it built the whole thing from one sentence." Showing that the output diverges every run, or that half the generated code was never asked for, would kill the magic. So the gap stays hidden behind a friendlier UI, the spec stays prose, and the structure that maps identical input to divergent output stays intact.

If you've ever tried to bring one of these frameworks into a team that ships to production, you already know. The demo is impressive. Then you run it on real code, open the PR, and it hits you.

The next wave says "harness." Rules files, system prompts, coding conventions &mdash; bolt a generic harness onto the agent and determinism follows. **It doesn't.** A harness constrains *style*: naming, folder layout, lint rules. It cannot constrain *decisions*: which files to create, which libraries to pick, how to decompose a feature into functions. Those decisions live at the plan level, and a generic harness has no opinion about them. Run the same prompt with the same harness twice &mdash; you'll get two structurally different codebases that both pass the style rules. **The harness is green. The output diverged.**

Generic harnesses solve the wrong layer. Determinism doesn't come from constraining how code looks. It comes from constraining what gets built, in what order, with what boundaries. That's a plan, not a harness.

---

planosh's answer is one thing. **Determinism.**

Same plan, same code. Every time. Everything else follows from this one principle:

- **When there's no room for interpretation, there's no over-implementation.** Over-implementation is a symptom of low determinism, not a separate problem.
- **When results converge, unattended execution stops being a gamble.** Kick off `plan.sh` before leaving work. Come back to working code.
- **When the plan predicts the output, teams review the plan instead of the code.** You don't need to open a 4,000-line PR.
- **When non-developers can review plans, the entire team can build with AI.** Determinism pulls people who can't read code into the development loop.

Determinism is measurable &mdash; run the same plan N times in parallel and diff the outputs. Divergence marks the places that aren't deterministic yet. Find the divergence, close it, measure again. That loop is what planosh is.

---

**The evidence.** We ran one `plan.sh` for **16 hours unattended**, migrating a production app (1+ year in the wild) from Flutter to React Native. No one touched the keyboard. The result wasn't production-ready &mdash; there were gaps. But compared to Claude Code's batch mode and Speckit-generated specs running the same migration, the completeness wasn't even in the same league. And this was **without calibration** &mdash; no parallel runs, no divergence detection, no harness tightening.

Writing the plan took 2 days. **That ratio is the whole point.** Invest in the plan, not in babysitting the execution. Calibration would have closed the remaining gaps before the run even started.

## Existing SDD tools vs. planosh

| Existing approaches | planosh |
|---|---|
| AI over-implements while interpreting the spec | The plan is deterministic, so interpretation never happens |
| Generic harness constrains style, not decisions | **Plan-specific harness constrains what gets built, in what order** |
| Measures "did it work?" | **Measures "did it converge?"** (same plan &rarr; same code, N runs) |
| You review 4,000 lines of PR | You review the plan. Skip the code. |
| Unattended execution is a gamble | **Determinism makes 16-hour unattended runs real** |
| "Deterministic" is a marketing claim | Determinism is a number you measure and shrink |

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

**2. Or use the reference skills** (Claude Code plugin):

```bash
# Install as a Claude Code plugin
claude plugin add ASDFGHoney/planosh

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

That said, there's active development ahead on two fronts:

- **plan.sh readability** &mdash; plan.sh is a shell script, and shell scripts get ugly fast. We're working on making plans easier to read and write without sacrificing determinism.
- **`planosh run`** &mdash; a CLI that spins up N isolated testbeds with shim-git and runs `plan.sh` in parallel. This is the execution environment that makes calibration (`--mode calibrate`) and internal parallelism (`--mode split`) real. See [#1](https://github.com/ASDFGHoney/planosh/issues/1) and [#2](https://github.com/ASDFGHoney/planosh/issues/2).

Both are open for contribution.

### Share your plan.sh

The `.plan/` directory in this repo is the community's best practice collection. Contribute yours via PR:

```
.plan/
+-- your-plan-name/
    +-- plan.sh
    +-- harness-for-plan.md
    +-- harness-for-step-N.md (optional)
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

## Best Practices

The `best-practices/` directory collects planosh's **execution pattern best practices**. (Where `.plan/` is the community's **domain plan collection**, this is the meta layer — how to design and run plan.sh itself.)

*No entries yet. Contribute yours via PR.*

## Further reading

- [Design Document](docs/DESIGN.md) &mdash; Problem definition, 3-layer constraint model, calibration loops, full plan.sh example
- [Example PRDs](docs/) &mdash; Retro webapp, C compiler, Markdown-to-slides
- [Best Practices](best-practices/) &mdash; Execution pattern reference implementations

## License

[MIT](LICENSE)

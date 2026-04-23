---
name: suggest
description: Pure architecture advice — no generation. Given requirements (greenfield or brownfield), recommends module selections, patterns, and design tradeoffs grounded in Trabuco's catalog. Use when the user wants to THINK before committing, evaluate options, or understand what Trabuco would do for their problem.
user-invocable: true
allowed-tools: [mcp__trabuco__suggest_architecture, mcp__trabuco__list_modules, mcp__trabuco__list_providers]
argument-hint: "[requirements]"
---

# Architecture advice, no generation

This skill advises. It does not create files. For generation, use `/trabuco:new-project`, `/trabuco:design-system`, or `/trabuco:add-module`.

## Flow

1. **Delegate to the specialist if possible**: invoke the `trabuco-architect` subagent. It loads the `trabuco_expert` MCP prompt and has the full reasoning framework. Only handle in-skill if the subagent isn't available.

2. **Gather the requirement if missing**: 2–3 tight questions, not a survey. Focus on what differentiates architectures:
   - Data shape (relational / document / cache / mix)
   - Traffic shape (sync REST, async events, scheduled jobs, long-running tasks, real-time streams)
   - AI agent involvement (will users interact with an LLM-backed agent?)
   - Scale posture (small internal tool / B2B service / large-scale SaaS)

3. **Get the recommendation**: call `mcp__trabuco__suggest_architecture` with the requirement. It returns ranked patterns + a recommended module config.

4. **Present as reasoned advice, not a list dump**:
   - Top recommendation with rationale grounded in the returned patterns
   - What tradeoff you're making (e.g., "Kafka gives durability but costs operational complexity vs RabbitMQ")
   - What you'd change if the requirements shifted (e.g., "if latency matters more than throughput, I'd drop EventConsumer and stick with REST + async jobs")

5. **Cite sources**: when you make a claim about what a module does, reference `trabuco://modules` or `list_modules` output. When you cite a pattern, reference `trabuco://patterns`. Don't hallucinate.

6. **Offer to act**: "If you want me to generate this, use `/trabuco:new-project`" — but only if they ask. Don't push.

## Rules

- **No generation from this skill**. Period. The user picks when to commit.
- **Be honest about fit**. If Trabuco isn't the right tool (read `trabuco://limitations`), say so — "for a React frontend, Trabuco won't help; pair it with a Next.js or Vaadin service."
- **Don't over-architect**. If a single-module service is enough, say that. Resist the urge to recommend 8 modules when 3 suffice.

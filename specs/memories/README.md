# Memories

Retrospectives from completed features. After each feature ships,
write a short retro tracing agent mistakes to specific missing rules.

## When to Write a Retro

After any feature where the agent made a mistake that a rule should have
prevented. If everything went smoothly, no retro needed.

## Retro Format

Files named: `YYYY-MM-DD-feature-name-retro.md`

```markdown
# Retro: <feature-name> — <date>

## Mistake (what the agent did)
<!-- Paste the specific transcript or describe the error -->

## Root Cause (why the rules didn't prevent it)
<!-- Was there no rule? Was the rule too vague? Did the agent ignore it? -->

## Rule Updated
<!-- Which file changed and what was the diff -->

## Before/After
<!-- Show the old rule text and the new rule text -->
```

## After Writing

1. Commit the retro to `specs/memories/`.
2. Commit the updated skill/rules file.
3. The retro provides traceability: "we added this rule because of
   this specific transcript on this date."

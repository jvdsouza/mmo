# Using Ralph for MMO Development

## What is Ralph?

Ralph is an autonomous AI development framework that enables continuous development cycles. It will help us iterate on the combat system automatically.

## Setup

### 1. Ralph is Already Cloned
Ralph is located at: `/home/user/ralph`

### 2. Configure Environment Variables

```bash
# In your shell or .bashrc
export ANTHROPIC_API_KEY="your-api-key-here"
export RALPH_MAX_LOOPS=20  # Max iterations (default: 100)
export RALPH_RATE_LIMIT=100  # API calls per hour
```

### 3. Our Development Plan

The plan is in `@fix_plan.md` at the project root. This guides Ralph through:
- Task 1-10: Build combat system step by step
- Clear success criteria for each task
- Exit signals when complete

## Running Ralph

### Manual Method (Recommended for Learning)

```bash
# Work through the plan manually
cd /home/user/mmo
# Read @fix_plan.md
# Implement each task
# Test
# Commit
# Next task
```

### Autonomous Method (Using Ralph)

```bash
cd /home/user/ralph

# Run Ralph pointing to our MMO project
./ralph.sh --project-dir /home/user/mmo --plan @fix_plan.md
```

Ralph will:
1. Read `@fix_plan.md`
2. Execute each task sequentially
3. Test after each implementation
4. Commit working code
5. Stop when Phase 1 complete

## Monitoring Ralph

Ralph provides live monitoring:

```bash
# Terminal 1: Run Ralph
./ralph.sh --project-dir /home/user/mmo

# Terminal 2: Monitor progress
tmux attach -t ralph-session
```

View:
- Current task being worked on
- API usage (calls remaining)
- Errors/issues
- Git commits

## Development Workflow

### Hybrid Approach (Recommended)

1. **Start manually** - Implement Task 1-2 yourself to understand the pattern
2. **Let Ralph continue** - Once comfortable, let Ralph handle Task 3-10
3. **Review Ralph's work** - Check commits, test functionality
4. **Iterate** - Adjust `@fix_plan.md` if needed

### Benefits

✅ **Continuous development** - Ralph works through tasks automatically
✅ **Rate limit aware** - Respects API limits
✅ **Smart completion** - Stops when done, not after arbitrary loops
✅ **Git integration** - Commits working code
✅ **Error recovery** - Retries on failures

## Current Plan Status

Check progress:
```bash
cd /home/user/mmo
cat @fix_plan.md | grep "🔲\|✅"
```

### Task Checklist

- ✅ Task 0: Architecture (DONE)
- 🔲 Task 1: Server interfaces
- 🔲 Task 2: Player stats
- 🔲 Task 3: Melee abilities
- 🔲 Task 4: Combat manager
- 🔲 Task 5: Client interfaces
- 🔲 Task 6: Client UI
- 🔲 Task 7: Ability manager
- 🔲 Task 8: Network integration
- 🔲 Task 9: Testing
- 🔲 Task 10: Documentation

## Customizing the Plan

Edit `@fix_plan.md` to:
- Add new tasks
- Change priorities
- Adjust success criteria
- Add implementation details

Ralph will adapt to changes.

## Safety Features

Ralph includes safeguards:
- **Max loops**: Stops after N iterations
- **Rate limiting**: 100 calls/hour default
- **Circuit breaker**: Stops on repeated failures
- **User prompts**: Asks before continuing after limits

## Tips

1. **Start small**: Run Ralph on one task first
2. **Review commits**: Always check what Ralph built
3. **Test frequently**: Verify functionality after each task
4. **Adjust the plan**: Be specific about what you want
5. **Commit manually**: You can work alongside Ralph

## Troubleshooting

**Ralph doesn't start:**
- Check API key is set
- Verify Ralph directory exists
- Check permissions on ralph.sh

**Ralph loops indefinitely:**
- Check exit signals in @fix_plan.md
- Ensure "PHASE_1_COMPLETE: false" at bottom
- Add clearer completion criteria

**API limits hit:**
- Ralph pauses automatically
- Resumes after rate limit window
- Adjust RALPH_RATE_LIMIT if needed

**Code doesn't compile:**
- Ralph should detect and fix
- If stuck, stop and review
- Adjust plan with more specific instructions

## Alternative: Manual Development

You don't need to use Ralph! The `@fix_plan.md` serves as:
- Clear roadmap
- Task breakdown
- Success criteria
- Implementation guide

Follow it manually if preferred.

## Current Status

We are at: **Task 1 - Server Combat Interfaces**

Next steps:
1. Implement `server/combat/ability.go`
2. Implement `server/combat/validators.go`
3. Test compilation
4. Commit
5. Move to Task 2

## When to Use Ralph

**Use Ralph for:**
- Repetitive tasks (implementing similar abilities)
- Well-defined tasks with clear specs
- Refactoring
- Documentation updates

**Do manually:**
- Architecture decisions
- Complex game design
- Performance optimization
- Creative work

## Summary

Ralph accelerates development by:
- Working through tasks autonomously
- Following our detailed plan
- Respecting API limits
- Committing working code

But you're always in control - review, test, and adjust as needed!

---

**Ready?** Let's either:
1. Start implementing Task 1 manually, or
2. Configure and run Ralph to work through the plan

Your choice!

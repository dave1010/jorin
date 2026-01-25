# Code Review Checklist: Smells to Catch

Each point should be checked individually.
Making an _improvement_ for any point is great but
only mark as complete (`[x]`) when you have confirmed
that it is _fully_ complete across the codebase and
there is nothing left to do on that point.

## Functions & Methods

**Too Long**
- Function over 50 lines? Split it
- More than 3 levels of nesting? Extract
- Scrolling to understand? Too long

**Too Many Parameters**
- More than 3-4 params? Use a config object
- Boolean flags? Split into separate functions
- Long param list that keeps growing? Missing abstraction

**Doing Too Much**
- Function name has "and" in it? Does two things
- Can't name it without "Manager" or "Handler"? Poorly scoped
- Has multiple reasons to change? Violates SRP

## Naming

**Vague Names**
- `data`, `info`, `item`, `thing`, `temp`, `obj` - be specific
- `handleClick`, `processData` - handle/process how?
- `flag`, `check`, `status` - what about it?

**Inconsistent Conventions**
- `getUserData` and `fetchUserInfo` - pick one pattern
- `isActive` vs `activeStatus` - booleans should ask questions
- `user_id` and `userId` in same codebase

## Conditionals

**Nested Ifs**
- More than 2 levels? Use guard clauses
- Giant if-else chains? Use polymorphism or lookup tables
- Same condition checked multiple places? Extract to named function

**Boolean Blindness**
- `setStatus(true)` - true means what?
- `calculate(data, false, true)` - use named params or enums
- Returning `true/false` for different error states? Use proper types

**Magic Numbers**
- `if (status === 3)` - extract to named constant
- `setTimeout(fn, 86400000)` - what's that number?
- Array index assumptions `[0]` without length check

## Data Structures

**Primitive Obsession**
- Passing around raw strings for emails, URLs, IDs? Wrap them
- Using arrays where you mean tuples? Use objects
- Parallel arrays? Use array of objects

**Data Clumps**
- Same 3-4 variables always passed together? Make an object
- Functions with same parameter subset? Missing entity

## Error Handling

**Silent Failures**
- Empty catch blocks
- Errors logged but not handled
- Returning null/undefined without caller expecting it

**Error Swallowing**
- Catching generic Exception/Error
- Try-catch around huge blocks
- Not re-throwing after logging

## DRY Violations

**Copy-Paste Code**
- Same logic in multiple places with tiny variations? Extract and parameterize
- Similar switch statements scattered? Polymorphism time
- Repeated validation logic? Centralize it

## Comments

**Commented-Out Code**
- Delete it, that's what git is for

**Redundant Comments**
- `// increment i` above `i++`
- Comments explaining what instead of why
- Outdated comments contradicting code

**Need Comments to Understand**
- If code needs comments to be clear, code is unclear
- Extract to well-named functions instead

## Classes & Objects

**God Object**
- Class doing everything
- 20+ methods or properties
- Name too generic: Manager, Service, Helper, Util

**Inappropriate Intimacy**
- Class accessing another's internals
- Friend classes that change together
- Exposing implementation details

**Lazy Class**
- Class with one method? Just a function
- Wrapper adding no value? Remove it

## Dependencies

**Tight Coupling**
- Importing concrete implementations instead of interfaces
- Directly instantiating dependencies instead of injecting
- Can't test without spinning up database/network

**Circular Dependencies**
- A imports B imports A
- Often indicates missing abstraction layer

## Testing Smells

**Test Knows Too Much**
- Testing private methods
- Assertions on internal state
- Brittle tests that break on refactor

**Mystery Guest**
- Test depends on external state
- Can't run in isolation
- Setup in different file

## Performance Red Flags

**N+1 Queries**
- Loop making DB calls
- Fetching related data one-by-one

**Premature Optimization**
- Complex caching before profiling
- Micro-optimizations that hurt readability

**String Concatenation in Loops**
- Building strings with `+=` in tight loops
- Use StringBuilder/array join

## State Management

**Global State**
- Mutable globals
- Singleton abuse
- Hidden dependencies on external state

**Temporal Coupling**
- Methods must be called in specific order
- State spread across multiple operations
- Init methods that must be called

## Code Organization

**Shotgun Surgery**
- One change requires touching many files
- Related code scattered
- Missing cohesion

**Feature Envy**
- Method using another class's data more than its own
- Should probably live in the other class

**Long Parameter Object Building**
- Chaining 10 setters before object is usable
- Use Builder pattern or constructor

## When You're Unsure

If you can't easily:
- Name it clearly
- Explain what it does in one sentence
- Test it in isolation
- Understand it after 6 months

It needs refactoring.

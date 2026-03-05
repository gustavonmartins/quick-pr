# Development Workflow - TDD with AI Assistance

## Overview
This project follows Test-Driven Development (TDD) with AI-assisted implementation. The workflow is designed to ensure code quality through testing first, implementation second, and refactoring last.

## TDD Cycle

### Phase 0: Feature Request & Test Planning
1. **User requests a feature** to be added
2. **Claude suggests tests** - proposes what tests should be written
3. **STOP and wait for approval** - User reviews test suggestions
4. If approved, proceed to Phase 1

**Rules for Test Suggestions:**
- Suggest tests that cover the feature requirements
- Include edge cases and error conditions
- Describe what each test will verify
- Get user approval before writing any code

### Phase 1: RED - Write Tests First
1. **After test approval**, Claude writes the actual test code
2. Tests should fail initially (RED phase)
3. Commit tests before implementation

**Rules for Test Creation:**
- Write tests exactly as approved
- Tests should be clear and maintainable
- Use table-driven tests where appropriate
- Tests must fail initially (proving they test real functionality)

### Phase 2: GREEN - Make Tests Pass
1. **After user approval**, Claude (or agents) implements code
2. **Loop until all verifications pass**:
   - Write minimal code to make tests pass
   - Run tests: `go test ./...`
   - Verify build: `go build`
   - Run linter: `go vet ./...`
   - If any check fails, fix implementation
   - Repeat until all checks pass

**Verification Checklist (must all pass):**
- ✅ Tests pass (`go test ./...`)
- ✅ Build succeeds (`go build`)
- ✅ Linter passes (`go vet ./...`)

**Rules for Implementation:**
- Write only enough code to make tests pass
- Don't add features not covered by tests
- **CRITICAL: ABSOLUTELY NEVER MODIFY TESTS DURING THIS PHASE**
  - Tests are written in Phase 1 and are IMMUTABLE in Phase 2
  - If a test seems wrong or hard to satisfy, ask the user first
  - Only modify implementation code, never test code
  - This is a core TDD principle that must never be violated
- Focus on getting to GREEN quickly
- Run ALL verifications before reporting completion

### Phase 3: REFACTOR - Improve Code Quality
1. **User requests refactoring** when ready
2. Claude refactors code for maintainability
3. **CRITICAL: DO NOT modify tests during refactoring**
4. **Loop until all tests pass**:
   - Refactor code (improve structure, readability, performance)
   - Run tests: `go test ./...`
   - If tests fail, fix refactoring
   - Repeat until all tests pass

**Rules for Refactoring:**
- **NEVER change tests in this phase**
- Tests are the contract - they must continue to pass
- Improve code structure, naming, organization
- Extract functions, reduce duplication
- Optimize performance if needed
- All original tests must still pass

## Commands

### Run All Verifications (Phase 2)
```bash
go test ./... && go build && go vet ./...
```

### Run Tests
```bash
go test ./...
```

### Run Tests with Coverage
```bash
go test ./... -cover
```

### Run Tests Verbosely
```bash
go test ./... -v
```

### Run Specific Test
```bash
go test ./... -run TestName
```

### Build Project
```bash
go build
```

### Run Linter
```bash
go vet ./...
```

### Format Code
```bash
go fmt ./...
```

## File Organization
- **Tests**: `*_test.go` files alongside implementation
- **Requirements**: `REQUIREMENTS.md` - source of truth for features
- **Workflow**: This file - how we work

## Communication Protocol

### When Claude Suggests Tests (Phase 0):
```
For [feature name], I suggest the following tests:

1. Test [name]: [what it verifies]
2. Test [name]: [what it verifies]
3. Test [name]: [edge case or error condition]

Do you approve these tests?
```

### When Claude Writes Tests (Phase 1):
```
I've written the tests in [file]:
- Test 1: [description]
- Test 2: [description]

Tests are currently failing (RED). Ready to implement?
```

### When Implementation is Complete:
```
All tests passing:
- [Test results summary]

Ready for next feature or refactoring.
```

### When Refactoring is Complete:
```
Refactoring complete. All tests still passing:
- [What was refactored]
- [Test results]
```

## Important Notes
- **User initiates** each feature request
- **Test suggestions require approval** before writing code (Phase 0)
- **Tests are written once** in Phase 1 (RED) after approval
- **CRITICAL: Tests are NEVER modified in Phase 2 (GREEN) or Phase 3 (REFACTOR)**
  - Once tests are written, they are the immutable contract
  - Only implementation code changes in Phase 2 and 3
  - Violating this rule breaks the TDD workflow
- **User approval required** at key checkpoints
- **Each phase completes** before moving to next phase
- **All tests must pass** before considering work complete

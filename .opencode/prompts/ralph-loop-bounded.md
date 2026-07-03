Use only for bounded execution loops.

Task contract:

- Implement one explicit task only.
- Do not expand scope.
- Continue only while the same task is incomplete.
- Stop when acceptance criteria pass.
- Emit `<promise>DONE</promise>` only after checks pass or a blocker is documented.

Forbidden:

- architecture exploration loops
- broad refactors
- vague improvement tasks
- dependency upgrades unless the task explicitly requires them

# Current Task

Task: QPREFIX-IMPL-009
Status: READY_FOR_CODEX

Authoritative specification:
`specs/QPREFIX-IMPL-009-PERFORMANCE-RELEASE-GATE.md`

Latest spec revision commit:
`8db5ac41d38914ce741d272bc5a9ddf8190f2064`

Accepted production candidate:
`f9c7f21e65915bd3eafcd5b12590b570c22a7d6f`

QPREFIX-IMPL-008 re-review result:
PASS under the original Fast Q-prefix constitution, with one clarification now made explicit in the specs.

Authoritative invariant:
`rows = QPrefixWidth(Level) = min(Level+1, 4)`

Therefore production authority is:
- Level 0 -> q0
- Level 1 -> q01
- Level 2 -> q012
- Level >=3 -> q0123

Important:
"internal q0123" is only shorthand for internal states whose current logical Level is >=3.
It is NOT a fixed internal Bootstrap width.

QPREFIX-IMPL-008 production BootstrapMany already derives input rows from input Level and output rows from core-output Level, so its production path conforms.

Low-level explicit-row helpers may still support narrower widths for legacy/transition compatibility tests. That capability is not the Q-prefix v2 production policy.

Release-gate structural validation must now prove exact equality at every production stage:
`authoritativeRows == QPrefixWidth(logicalLevel)`
—not merely `authoritativeRows <= 4`.

Explicitly verify transitions such as:
- 3->2: q0123 -> q012
- 2->1: q012 -> q01
- 1->0: q01 -> q0

Continue QPREFIX-IMPL-009 using the revised spec.
Do not change production source unless a hard release bug is found and reported first.

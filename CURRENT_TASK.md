# Current Task

Task: QPREFIX-IMPL-004
Status: COMPLETE

Accepted Secondary commit:
`18f4da53f03acd065d18550a8f2106725462896f`

Review correction:
The temporary `QPREFIX-IMPL-004-REVIEW-FIX` blocker was withdrawn.

Authoritative Fast invariant:
- maintained Q-prefix residues uniquely determine a small centered integer lift X whenever the strict capacity invariant holds;
- dormant q4+ rows are not required to recover X;
- for targetLevel >= 3, a valid q0123-bounded Fast lift already lies inside the centered interval of the larger logical target modulus, so canonical contraction equals same-lift contraction;
- for targetLevel < 3, canonical contraction may genuinely change representative and remains an explicit semantic boundary.

Updated architecture clarification:
`specs/QPREFIX-IMPL-004-RESCALE-LEVEL-TRANSITIONS.md`
at Primary commit
`9e6dfc16234f55692bbf747176f51a37bb26936e`.

QPREFIX-IMPL-004 acceptance:
- q0123 fixed-width CRT/Rescale capability accepted;
- public Rescale/RescaleTo remain legacy-authority activated;
- logical q_ell divisor behavior accepted;
- SameLift/Canonical Level-transition semantics accepted under the Fast bounded-lift invariant;
- C2S q3-poison isolation and full regressions passed.

Next task has not yet been routed.

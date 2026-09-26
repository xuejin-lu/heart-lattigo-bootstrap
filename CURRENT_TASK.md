# Current Task

Task: FAST-STORAGE-008
Status: COMPLETE

Accepted Secondary implementation:
`1aafc442595da9af41fdc97b03b91ccb2431fe3a`

Classification:
`FAST_STORAGE_008_WIDTH2_VALID_NOT_COMPETITIVE`

Accepted result:
- explicit non-mutating private-F storage contraction 3->2 is implemented by prefix-row retention under strict centered-capacity proof;
- width-2 and width-3 plaintext mirrors and LinearTransform execution are supported;
- actual LogN13 compressed C2S passes strict width-2 and width-3 capacity for all four factors;
- width-2 and width-3 authoritative lifts match exactly after every factor group;
- exported maintained logical q rows match the existing logical Fast path exactly;
- production Bootstrap/C2S routing and fixed production width-3 policy remain unchanged.

Independent Web review:
- candidate commit is based directly on the authorized FAST-STORAGE-008 task pointer;
- contraction performs no INTT/CRT/NTT reconstruction and preserves metadata/bounds exactly;
- width-specific capacity checks use the actual storage product;
- real C2S tests hard-assert expected capacity instead of relying only on log-and-return;
- no production Bootstrap/C2S source file is modified;
- no blocking correctness defect found.

Reported capacities:
- S2 = 1329227995775244468652735166391779329;
- S3 = 1532495540835573786841920346258506432323010736431562753;
- all four real LogN13 factors fit width 2.

Reported benchmark medians:
Factor 0:
- Logical Fast: 309222 ns/op, 1986 B/op, 77 allocs/op;
- private-F width 3: 869347 ns/op, 4202882 B/op, 214 allocs/op;
- private-F width 2: 607639 ns/op, 2891522 B/op, 194 allocs/op.

Full four-factor chain:
- Logical Fast: 3788556 ns/op, 20091632 B/op, 902 allocs/op;
- private-F width 3: 10589903 ns/op, 23124178 B/op, 1452 allocs/op;
- private-F width 2: 7307347 ns/op, 15911634 B/op, 1334 allocs/op.

Decision:
- F2 is approximately 30% faster than F3 and passes the first experiment signal;
- F2 full chain is approximately 1.93x the Logical Fast time and fails the <=1.5x production-interest gate;
- therefore width 2 is mathematically valid but not yet performance-competitive.

Important root-cause signal:
- isolated Logical Fast LinearTransform reuses evaluator-owned scratch;
- current private-F LinearTransform still allocates rotation/baby/inner/outer working state per call;
- factor-0 B/op differs by roughly 2 KB vs 2.89 MB, so one bounded scratch-parity experiment remains justified before closing the C2S private-F direction.

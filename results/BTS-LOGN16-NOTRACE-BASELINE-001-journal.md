# BTS-LOGN16-NOTRACE-BASELINE-001 execution journal

- Authorized Bootstrap budget: 2 total; at most 1 Standard and 1 Fast. A pre-Bootstrap resume is allowed only when process exit and non-entry are verified.
- BTS input/source: unchanged `external/BTS`; original defaults; source manifest SHA-256 `53205b37abd1fb57a960394256b958f5d5e5b1a66f527bfb7cbadabfe3cf0cda`.
- Standard: source archived from clean pinned commit `5dbffbdea05394de2ca3a432ed5318aa832e3f40` (tree `ba959dd386565d13bf1147a847e5dfe2c0361f90`, archive SHA-256 `5cff4c901bca6cfd4ab873d37ebc239fcf18bc71541db3d521b6bbd6d36060cd`); independent binary compiled successfully to `/private/tmp/bts-logn16-notrace-001-resume.CT7bjX/bts-standard`, SHA-256 `ad5bfdcec4ab380ebe824459ea834aa911d9effb555c38dce58cbd80b26e3823`.
- Fast: active clean `fast-qprefix` checkout at `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f`; independent binary compiled successfully to `/private/tmp/bts-logn16-notrace-001-resume.CT7bjX/bts-fast`, SHA-256 `dbd62f0dc9b275f11211ac6fa744ba50a89dfe0309545b995152d7e58a9dbd53`.
- Both binaries use the same BTS module/source with separate temporary `-modfile` replacements and `GOWORK=off`; no build tags or BTS/backend source edits.

## Attempts

| Backend | State | Reserved command | Result |
|---|---|---|---|
| Standard | COMPLETE (sole call consumed) | `/private/tmp/bts-logn16-notrace-001-resume.CT7bjX/bts-standard` | NewBTS 44.316 s; Bootstrap 2m19.653s; exit 0; `L3→L18`, Scale `2^32`; max abs error `1.708e-3` at slot 7606; mean abs error `1.172e-4`. Raw log: `BTS-LOGN16-NOTRACE-STANDARD-001-resume-2.log`. Never rerun. |
| Fast | PANIC (sole call consumed) | `/private/tmp/bts-logn16-notrace-001-resume.CT7bjX/bts-fast` | `NewBTS` 1m0.538s; input `L3`, Scale `2^32`; Bootstrap panicked during C2S at `ckksEval.DecomposeNTT` → `ring.INTT`/`ring.inttCoreLazy` (`len(p1)=0`). Process exit 2. Raw log: `BTS-LOGN16-NOTRACE-FAST-001.log`. Never rerun. |

## Final checkpoint

- Bootstrap calls: Standard `1/1` completed; Fast `1/1` entered and panicked; no retries or additional Bootstrap/diagnostic calls.
- Standard raw log: `BTS-LOGN16-NOTRACE-STANDARD-001-resume-2.log`, SHA-256 `440524a8789cf3e6987d1bd6e66556e74cd9ccd093183742968b2cf0e90c9431`.
- Fast raw log: `BTS-LOGN16-NOTRACE-FAST-001.log`, SHA-256 `433f0b13764785d09cadd5cdaaaf31ddb7add41a735614970b2bf9b2baee2ef7`.
- The first Standard process log (`BTS-LOGN16-NOTRACE-STANDARD-001.log`) has only the parameter line; the session ended before `[NewBTS]` and before Bootstrap. It is retained as partial evidence and is not treated as a Bootstrap attempt.
- Outcome: `NEEDS_WEB_REVIEW`; Fast did not produce a decoded output, so there is no paired numerical acceptance or speed ratio.

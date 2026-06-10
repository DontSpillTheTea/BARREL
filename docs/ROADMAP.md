# Roadmap

**Status:** Azure-hosted BARREL is live as an AI-native review assistant prototype, with provider verification and quality hardening still in progress.

**Note:** `task` is the supported daily command interface. `make setup` exists for initial bootstrap only.

## Completed and current

### Platform and product foundation

- Azure-hosted web and API application path established
- AI-native parser plumbing integrated into the async label analysis flow
- lightweight object/blob review storage established
- Review History and detail viewer implemented
- single-image analysis implemented through async job submission
- zip batch submission and per-image job processing implemented
- evaluator login flow integrated

### Current architecture direction

- `ai_native` is the default parser
- `azure_vision_ocr` is optional debug/baseline evidence only
- deterministic BARREL checks run on top of structured AI output
- review evidence is stored in object/blob-style storage, not Postgres

## In progress

- robust AI endpoint quota and provider verification
- local-first smoke coverage for AI-native flows before Azure deployment claims
- batch queue UI hardening and review ergonomics
- field confidence and scoring quality tuning
- clearer provider error surfacing when Azure OpenAI quota or region policy blocks deployment

## Future work

- custom domain and polished deployment ergonomics
- richer expected-vs-observed diffing in the detail view
- reviewer export and reporting flows
- broader and richer regulatory rule catalog
- stronger confidence calibration and evidence rendering
- optional database adoption only if object/blob storage becomes insufficient for history/query needs

## Verification policy

- use local tests first
- use Azure smoke second
- do not run heavy Azure validation by default
- do not claim success without smoke tests
- do not keep region-hopping Azure OpenAI without proving quota and policy facts first

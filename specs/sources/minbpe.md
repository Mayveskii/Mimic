```yaml
repo: Mayveskii/minbpe
url: https://github.com/Mayveskii/minbpe
language: Python
status: partial
last_sync: "2025-05-17"

description: |
  Fork of karpathy/minbpe (3K+ stars). Minimal, clean implementation of Byte Pair Encoding
  with progressive inheritance (Tokenizer → BasicTokenizer → RegexTokenizer → GPT4Tokenizer),
  pair-merge compression, versioned parsing rules, and special token registration. Reference
  implementation for understanding BPE internals.

advantages:
  - id: mb_progressive_inheritance
    what: 4-level class hierarchy: Tokenizer(base encode/decode) → BasicTokenizer(BPE merge) → RegexTokenizer(regex pre-split) → GPT4Tokenizer(GPT-4 pattern + special tokens); each level adds exactly one capability
    evidence: "minbpe/base.py — Tokenizer with encode/decode; minbpe/basic.py — BasicTokenizer adds BPE; minbpe/regex.py — RegexTokenizer adds pre-split; minbpe/gpt4.py — GPT4Tokenizer adds special tokens"

  - id: mb_pair_merge_compression
    what: BPE core: get_stats() counts byte pairs → merge() replaces most frequent pair with new token → repeat until no pair above threshold; deterministic given same input
    evidence: "minbpe/basic.py — get_stats() returns Counter of adjacent pairs; merge() replaces pair in token list with new integer id"

  - id: mb_versioned_parsing_rules
    what: Version-specific regex pre-split patterns: GPT2_SPLIT_PATTERN vs GPT4_SPLIT_PATTERN; different tokenization behavior controlled by pattern constant
    evidence: "minbpe/regex.py — GPT2_SPLIT_PATTERN regex; minbpe/gpt4.py — GPT4_SPLIT_PATTERN regex; different contractions and number handling"

  - id: mb_special_token_registration
    what: register_special_tokens() maps string tokens (e.g., <|endoftext|>) to reserved integer IDs; special tokens never split by BPE, matched as whole units in encode
    evidence: "minbpe/gpt4.py — register_special_tokens() method; encode() with special token regex matching before BPE; special_tokens dict"

applications:
  - advantage_id: mb_progressive_inheritance
    implemented_in: core/tokenizer.c
    mechanism: "C struct hierarchy: TokenizerBase{encode,decode} → BasicTokenizer{bpe_merge} → RegexTokenizer{pre_split} → GPT4Tokenizer{special_tokens}; each level extends vtable"
    invariant: "Each level adds exactly one capability. No capability removed at deeper level. Base encode/decode always functional."
    status: planned

  - advantage_id: mb_pair_merge_compression
    implemented_in: core/bpe.c
    mechanism: "Count adjacent pairs → find max frequency pair → merge into new id → rebuild pair counts → repeat until vocab_size reached or no pair > 1"
    invariant: "Merge is deterministic for same input + same merges list. New token id = len(vocab) at merge time. Order of merges matters."
    status: planned

  - advantage_id: mb_versioned_parsing_rules
    implemented_in: core/pattern.c
    mechanism: "Compile version-specific regex pattern → pre-split input into chunks → BPE each chunk independently → concatenate results"
    invariant: "GPT2 pattern splits contractions differently from GPT4. Pre-split boundaries never crossed by BPE. Pattern applied before any merge."
    status: planned

  - advantage_id: mb_special_token_registration
    implemented_in: core/special.c
    mechanism: "Register special_tokens map: string → reserved id; during encode, scan for special token strings first → replace with reserved id → BPE remaining text"
    invariant: "Special token strings never split by BPE. Reserved ids outside normal merge range. Duplicate special token → error."
    status: planned

control:
  - advantage_id: mb_progressive_inheritance
    verification: "Unit test: create BasicTokenizer → verify BPE works but no pre-split; create GPT4Tokenizer → verify all features"
    update_trigger: "Re-analyze when minbpe releases new version"
    last_verified: never

  - advantage_id: mb_pair_merge_compression
    verification: "Unit test: 'aaab' with vocab extension → verify merge of 'aa' into new token; verify deterministic output"
    update_trigger: "Re-analyze when minbpe releases new version"
    last_verified: never

  - advantage_id: mb_versioned_parsing_rules
    verification: "Unit test: 'don't' → verify GPT2 pattern splits differently from GPT4 pattern on contractions"
    update_trigger: "Re-analyze when minbpe releases new version"
    last_verified: never

  - advantage_id: mb_special_token_registration
    verification: "Unit test: register <|endoftext|> → encode text containing it → verify single token id, not split by BPE"
    update_trigger: "Re-analyze when minbpe releases new version"
    last_verified: never
```

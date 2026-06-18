# Awesome GonkaGate

Browse community projects built with GonkaGate and see what to include in a pull request to Awesome GonkaGate.

Open [Awesome GonkaGate](https://github.com/GonkaGate/awesome-gonkagate) when you want real community projects built with GonkaGate or you are ready to submit your own. Need an official runnable example instead of a community repo? Start with [GonkaGate examples](https://github.com/GonkaGate/gonkagate-examples).

## What to include in a pull request

When your project is ready, open a pull request in that repository.

1. Explain what the project does and who it helps.
2. Make clear what another developer can reuse: a full app, starter, reference pattern, or one narrow integration.
3. List the exact GonkaGate settings: model ID, provider choice, required environment variables, and any other configuration needed to reproduce the result.
4. Add a short `how to run` path and one verification step that another developer can repeat without guesswork.

## What makes a useful entry

- The scope is obvious: full app, starter, reference pattern, or narrow integration example.
- The repo shows expected output, a screenshot, or CLI output when that makes the happy path easier to verify.
- Manual steps, caveats, retries, auth assumptions, and unsupported paths are called out instead of left implicit.
- Another developer can understand the setup without opening several extra pages.

## Check a repo quickly before you try it

- Check for exact runtime and framework versions, not only screenshots.
- Look for the model ID, provider settings, and any GonkaGate-specific configuration required to reproduce the result.
- Prefer repos with a short `how to run` path and one verification step.
- Treat missing caveats as risk. Auth setup, retries, and unsupported paths should be visible before you invest time.

## Hold off on submitting if…

- The repo mostly repeats an official example without adding new context.
- It depends on private setup knowledge or undocumented manual steps.
- It cannot show a reproducible happy path or expected output.
- It is mostly screenshots or marketing copy with no runnable proof.

## See also

- Need step-by-step setup for a specific wrapper or coding tool? Use [Framework and Tool Guides](/en/docs/guides/community/frameworks-and-integrations-overview).
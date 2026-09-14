# Contributing

v-eval is a research and design prototype. The most useful early contributions are realistic evaluation cases, corrections to the research, and feedback on whether the proposed report helps a developer decide. Please read the [code of conduct](CODE_OF_CONDUCT.md) first.

## Where things live

- `docs/requirements.md`: the consolidated requirements with their sources. Requirements change only through a maintainer decision.
- `docs/decisions/`: one decision record per decision in [MADR](https://adr.github.io/madr/) format, `NNNN-title.md`, indexed in `docs/decisions/README.md`. Copy `docs/decisions/adr-template.md` to propose a new one.
- `docs/architecture/`: the designs that follow from the decisions.
- `docs/research/`: research reports with sources and access dates.
- `docs/`: the learning handbook, starting from `docs/README.md`.
- `skills/`: the skill in the Agent Skills format; `templates/`: the contract template; `examples/`: worked examples.
- `tools/docs/`: maintainer-only documentation tooling with tests (source register generator, local link check). Exempt from the product's no-interpreter rule by [decision 0022](docs/decisions/0022-repository-maintenance-tooling.md); never shipped and never on the skill's required path. After adding or removing a citation anywhere in the docs, regenerate the source register with `python tools/docs/vdocs.py register .`; CI fails with `register stale` otherwise.

## Contributing an evaluation case

Open an issue with the "Evaluation case" template. Include the intended outcome, artifact and revision, explicit criteria, available evidence, expected per-criterion results, why those results are justified, and any gaming traces the case contains. Share only material you are authorized to publish; prefer minimal synthetic reproductions when source work is private. Distinguish independently verified evidence from candidate-authored claims. Cases accepted into the pilot set are labeled by the protocol in [decision 0015](docs/decisions/0015-pilot-cases-and-labeling.md); the anchor set is never published.

## Contributing research

Cite original papers or official project documentation with URLs and access dates. Record access dates for mutable claims such as licenses and features. Label facts, inferences, single-source claims, and unverified items. Distinguish measured findings from product recommendations. Do not copy third-party benchmarks or datasets into this repository without checking their terms. Use the "Research correction" issue template for errors.

## Proposing a design change

Open an issue with the "Design proposal" template describing the user problem, the requirement IDs and decision records affected, why an existing tool or the current design is insufficient, and the evidence. If accepted, the change lands as a new or superseding decision record plus the architecture update.

## Documentation conventions

- Keep the repository's hedged, evidence-first tone. Do not claim that something is implemented, validated, or superior unless the evidence is in the repository.
- Relative links only, checked by the documentation workflow (markdownlint and a link check run on every pull request).
- Every script and program must run on macOS, Linux, and Windows, and the skill must not depend on one host's tool names; see [decision 0020](docs/decisions/0020-portability-constraints.md).
- Pull requests use the template and must show real evidence for any behavioral claim.

## License

Original contributions are provided under the repository's [MIT license](LICENSE), inbound equals outbound. Keep the citations and evidence a reviewer needs to assess a claim.

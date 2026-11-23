# IOS XE Supplemental Validated Artifacts

These sanitized artifacts were validated against live IOS XE telemetry sources. BGP/environment validation was completed on `2026-06-01`; PTP ingestion-staging validation was completed on `2026-06-13`, and clock CSV value validation was checked on `2026-06-15`. A Catalyst 9300 current recapture on `2026-06-25` validates the current environment aggregate transform and the PTP correction-stats session identity.

Validation sources:

```text
sanitized IOS XE validation tenant
sanitized PTP validation tenant
```

Validation status:

- The managed XE collector was deployed with the combined default IOS XE + supplemental BGP/environment telemetry configuration.
- Live source data was observed for default XE CPU, memory, interface data plus the new BGP neighbor and environment sensor object types.
- The BGP and environment dictionary templates were validated successfully through the PCA UI/Weld staging path.
- The Catalyst 9300 recapture produced transformed environment rows with device-specific sensor names and no fixed collector field whitelist.
- The PTP object types were published successfully through the PCA ingestion-staging path, producing tenant-generated ingestion-specification exports.
- PTP clock raw CSV exports showed correct nanosecond normalization after the final microsecond-unit publish.
- The user confirmed the configuration works with a live data source.

## Contents

- `validation-summary.json`: compact summary of validated object counts, directions, units, and sample counts.
- `tenant-dictionaries/`: sanitized live tenant dictionary API responses for the BGP/environment object types and the four PTP object types, plus the PTP repo-vs-live comparison summary.
- `tenant-staging/`: sanitized staging snapshots for the BGP/environment object types and published PTP ingestion-specification exports. BGP/environment snapshots are retained as validation evidence only; they are not promoted as package ingestion specs.
- `tenant-profile/`: sanitized ingestion-profile snapshot at capture time. The new object types may still need profile metric enablement depending on tenant state.
- `collector-deployment/`: deployed managed-agent `telemetry.dataTransformation` configuration and validation checks.
- `golden-samples/`: live telemetry capture bundles and filtered samples for BGP, environment, and PTP inspection.
- `repo-artifacts/`: repository-authored dictionary templates matching the live behavioral contract.
- `../pca-ingestion-specifications-configuration/`: forward-format ingestion specifications for package use.
- `../pca-ingestion-dictionaries-configuration/`: transition-format dictionary templates for package use.

## Validated Dictionary Contract

`cisco-cisco-telemetry-xe-bgp-neighbor`:

- `33` metrics
- `21` dimensions
- directions: `-1`
- units: `value`, `sec`, `bytes`
- preserves `totalDropped` for raw metric `xe_bgp_neighbors_connection_total_dropped`

`cisco-cisco-telemetry-xe-environment-sensor`:

- `41` metrics
- `11` dimensions
- directions: `-1`
- units: `deg_c`, `volt`, `value`
- one aggregate environment session per router/source

PTP dictionary templates generated from the published ingestion specifications, compared against the live validation-lab dictionaries on `2026-06-15`:

```text
cisco-cisco-telemetry-xe-ptp-clock              11 metrics / 22 dimensions
cisco-cisco-telemetry-xe-ptp-parent              5 metrics / 16 dimensions
cisco-cisco-telemetry-xe-ptp-port               10 metrics / 16 dimensions
cisco-cisco-telemetry-xe-ptp-correction-stats    2 metrics / 13 dimensions
```

PTP nanosecond metrics use explicit backtick casts and microsecond units, for example:

```sql
cast(`xe_ptp_clock_mean_path_delay_ns` as DOUBLE) / 1000
```

The live validation-lab API comparison found no metric or dimension diffs between the repo templates and tenant dictionaries. The only expected metadata difference is `dictionaryType`: repo templates use `global`, while the live tenant stores the published dictionaries as `custom`.

## Notes

The tenant dictionary artifacts are live `custom` dictionaries. The product templates under `pca-ingestion-dictionaries-configuration/` may retain template-oriented metadata such as `dictionaryType`.

During the ingestion-spec transition, keep both artifact families:

- tenant-generated ingestion specifications are the forward default where available;
- dictionary templates remain required for object types that do not yet have tenant-generated ingestion-spec exports and for compatibility with older workflows and reviews.

Do not treat older broad dictionary-collection snapshots as validated truth. The validated tenant dictionary files here were fetched by exact dictionary ID after the successful publish.

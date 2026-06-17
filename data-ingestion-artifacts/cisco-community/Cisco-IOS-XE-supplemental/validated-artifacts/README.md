# IOS XE Supplemental Validated Artifacts

These artifacts were validated against live IOS XE telemetry sources. The first BGP/environment validation source was the Open Fiber PCA lab tenant on `2026-06-01`; PTP ingestion-staging validation was completed in the United Medley lab on `2026-06-13` and clock CSV value validation was checked on `2026-06-15`.

Tenant:

```text
https://openfiber.analytics.accedian.io
https://united.medley.npav.accedian.net
```

Validation status:

- The managed XE collector was deployed with the combined default IOS XE + supplemental BGP/environment telemetry configuration.
- Live source data was observed for default XE CPU, memory, interface data plus the new BGP neighbor and environment sensor object types.
- The BGP and environment tenant dictionaries were published successfully through the PCA UI/Weld staging path.
- The PTP object types were published successfully through the PCA ingestion-staging path.
- PTP clock raw CSV exports showed correct nanosecond normalization after the final microsecond-unit publish.
- The user confirmed the configuration works with a live data source.

## Contents

- `validation-summary.json`: compact summary of validated object counts, directions, units, and sample counts.
- `tenant-dictionaries/`: exact live tenant dictionary API responses for the BGP/environment object types and the four United Medley PTP object types, plus the PTP repo-vs-live comparison summary.
- `tenant-staging/`: live staging specification snapshots for the BGP/environment object types and published PTP ingestion-specification exports.
- `tenant-profile/`: ingestion-profile snapshot at capture time. The new object types may still need profile metric enablement depending on tenant state.
- `collector-deployment/`: deployed managed-agent `telemetry.dataTransformation` configuration and validation checks.
- `golden-samples/`: live telemetry capture bundle and filtered samples for BGP/environment inspection.
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

PTP dictionary templates generated from the published ingestion specifications, compared against the live United Medley dictionaries on `2026-06-15`:

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

The live United Medley API comparison found no metric or dimension diffs between the repo templates and tenant dictionaries. The only expected metadata difference is `dictionaryType`: repo templates follow the product/community package convention with `global`, while the live tenant stores the published dictionaries as `custom`.

## Notes

The tenant dictionary artifacts are live `custom` dictionaries. The package templates under `pca-ingestion-dictionaries-configuration/` are cisco-community repo artifacts and may retain template-oriented metadata such as `dictionaryType`.

During the ingestion-spec transition, keep both artifact families:

- ingestion specifications are the forward default;
- dictionary templates remain required for compatibility with older workflows and reviews.

Do not treat older broad dictionary-collection snapshots as validated truth. The validated tenant dictionary files here were fetched by exact dictionary ID after the successful publish.

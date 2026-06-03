# IOS XE Supplemental Validated Artifacts

These artifacts were validated against a live IOS XE telemetry source on 2026-06-01. The first validation source was the Open Fiber PCA lab tenant.

Tenant:

```text
https://openfiber.analytics.accedian.io
```

Validation status:

- The managed XE collector was deployed with the combined default IOS XE + supplemental BGP/environment telemetry configuration.
- Live source data was observed for default XE CPU, memory, interface data plus the new BGP neighbor and environment sensor object types.
- The BGP and environment tenant dictionaries were published successfully through the PCA UI/Weld staging path.
- The user confirmed the configuration works with a live data source.

## Contents

- `validation-summary.json`: compact summary of validated object counts, directions, units, and sample counts.
- `tenant-dictionaries/`: exact live tenant dictionary API responses for the two new object types.
- `tenant-staging/`: exact live staging specification snapshots for the two new object types.
- `tenant-profile/`: ingestion-profile snapshot at capture time. The new object types may still need profile metric enablement depending on tenant state.
- `collector-deployment/`: deployed managed-agent `telemetry.dataTransformation` configuration and validation checks.
- `golden-samples/`: live telemetry capture bundle and filtered samples for BGP/environment inspection.
- `repo-artifacts/`: repository-authored dictionary templates matching the live behavioral contract.

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

## Notes

The tenant dictionary artifacts are live `custom` dictionaries. The package templates under `pca-ingestion-dictionaries-configuration/` are cisco-community repo artifacts and may retain template-oriented metadata such as `dictionaryType`.

Do not treat older broad dictionary-collection snapshots as validated truth. The validated tenant dictionary files here were fetched by exact dictionary ID after the successful publish.

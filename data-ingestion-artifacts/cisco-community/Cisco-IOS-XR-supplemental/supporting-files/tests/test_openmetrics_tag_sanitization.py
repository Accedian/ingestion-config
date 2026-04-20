import json
import re
import tomllib
import unittest
from functools import lru_cache
from fnmatch import fnmatch
from pathlib import Path


ARTIFACT_DIR = Path(__file__).resolve().parents[2]
TELEGRAF_CONF = ARTIFACT_DIR / "telemetry-collector-configuration" / "telegraf.conf"
L57_GOLDEN_BATCH = (
    ARTIFACT_DIR
    / "supporting-files"
    / "golden-samples"
    / "l57-2026-03-15"
    / "l57-validation-metrics.jsonl"
)

PROMETHEUS_LABEL_RE = re.compile(r"^[A-Za-z_][A-Za-z0-9_]*$")

GENERIC_SLASH_RENAME_PATTERN = "^([^/]+)/(.+)$"
GENERIC_SLASH_RENAME_REPLACEMENT = "${1}_${2}"

EXPECTED_SLASH_TAG_RENAMES = {
    "global/process_instance_node": "global_process_instance_node",
    "local_end_point_address/ip_address_type": "local_end_point_address_ip_address_type",
    "local_end_point_address/ipv6": "local_end_point_address_ipv6",
    "loss/received_block_color": "loss_received_block_color",
    "pem_info_array/node_name": "pem_info_array_node_name",
    "pem_info_array/node_status": "pem_info_array_node_status",
    "pem_info_array/node_type": "pem_info_array_node_type",
}


@lru_cache(maxsize=1)
def load_telegraf_config():
    return tomllib.loads(TELEGRAF_CONF.read_text())


@lru_cache(maxsize=1)
def load_golden_batch():
    metrics = []
    for line in L57_GOLDEN_BATCH.read_text().splitlines():
        if line.strip():
            metrics.append(json.loads(line))
    return metrics


def telegraf_replacement_to_python(replacement: str) -> str:
    return re.sub(r"\$\{(\d+)\}", r"\\g<\1>", replacement)


@lru_cache(maxsize=1)
def configured_tag_rename_rules():
    config = load_telegraf_config()
    rules = []
    for processor in config.get("processors", {}).get("regex", []):
        namepass = processor.get("namepass", ["*"])
        for rule in processor.get("tag_rename", []):
            rules.append(
                {
                    "namepass": namepass,
                    "pattern": rule["pattern"],
                    "replacement": rule["replacement"],
                }
            )
    return rules


@lru_cache(maxsize=1)
def configured_field_rename_rules():
    config = load_telegraf_config()
    rules = []
    for processor in config.get("processors", {}).get("regex", []):
        namepass = processor.get("namepass", ["*"])
        for rule in processor.get("field_rename", []):
            rules.append(
                {
                    "namepass": namepass,
                    "pattern": rule["pattern"],
                    "replacement": rule["replacement"],
                }
            )
    return rules


def apply_configured_tag_renames(metric_name: str, tag_name: str) -> str:
    renamed = tag_name
    for rule in configured_tag_rename_rules():
        if not any(fnmatch(metric_name, pattern) for pattern in rule["namepass"]):
            continue
        renamed = re.sub(
            rule["pattern"],
            telegraf_replacement_to_python(rule["replacement"]),
            renamed,
        )
    return renamed


def apply_configured_field_renames(metric_name: str, field_name: str) -> str:
    renamed = field_name
    for rule in configured_field_rename_rules():
        if not any(fnmatch(metric_name, pattern) for pattern in rule["namepass"]):
            continue
        renamed = re.sub(
            rule["pattern"],
            telegraf_replacement_to_python(rule["replacement"]),
            renamed,
        )
    return renamed


class TestXrSupplementalOpenMetricsTagSanitization(unittest.TestCase):
    def test_golden_batch_contains_the_problem_signal(self):
        offending_tags = set()
        offending_fields = set()
        for metric in load_golden_batch():
            for tag_name in metric.get("tags", {}):
                if not PROMETHEUS_LABEL_RE.match(tag_name):
                    offending_tags.add(tag_name)
            for field_name in metric.get("fields", {}):
                if not PROMETHEUS_LABEL_RE.match(field_name):
                    offending_fields.add(field_name)

        self.assertIn(
            "loss/received_block_color",
            offending_tags,
            "The l57 golden batch should preserve the IPM slash-bearing tag that triggered this regression.",
        )
        self.assertIn(
            "latency/average_latency",
            offending_fields,
            "The l57 golden batch should preserve slash-bearing field names so OpenMetrics-safe renames stay covered.",
        )

    def test_expected_slash_tag_rename_rules_exist(self):
        slash_rules = [
            rule
            for rule in configured_tag_rename_rules()
            if rule["pattern"] == GENERIC_SLASH_RENAME_PATTERN
        ]
        self.assertEqual(4, len(slash_rules))
        self.assertTrue(
            all(rule["replacement"] == GENERIC_SLASH_RENAME_REPLACEMENT for rule in slash_rules)
        )

        actual = {
            tag_name: apply_configured_tag_renames("test_metric", tag_name)
            for tag_name in EXPECTED_SLASH_TAG_RENAMES
        }
        self.assertEqual(EXPECTED_SLASH_TAG_RENAMES, actual)

    def test_golden_batch_projects_to_openmetrics_safe_names(self):
        violations = []
        for metric in load_golden_batch():
            metric_name = metric["name"]
            for tag_name in metric.get("tags", {}):
                renamed = apply_configured_tag_renames(metric_name, tag_name)
                if not PROMETHEUS_LABEL_RE.match(renamed):
                    violations.append(("tag", metric_name, tag_name, renamed))
            for field_name in metric.get("fields", {}):
                renamed = apply_configured_field_renames(metric_name, field_name)
                if not PROMETHEUS_LABEL_RE.match(renamed):
                    violations.append(("field", metric_name, field_name, renamed))

        self.assertEqual(
            [],
            violations,
            "Configured tag and field renames should leave every golden-sample name compatible with OpenMetrics output.",
        )


if __name__ == "__main__":
    unittest.main()

import json
import sys

def to_camel_case(value):
    words = value.split("_")
    return words[0].lower() + "".join(word.capitalize() for word in words[1:])

def clean_analytics_name(name, remove_prefix=False):
    # Remove trailing '_value'
    if name.endswith("_value"):
        name = name[:-6]
    words = name.split("_")
    if remove_prefix and len(words) > 1:
        words = words[1:]  # Remove first word
    return to_camel_case("_".join(words))

def process_metrics(metrics, remove_prefix=False):
    for metric in metrics:
        if "analyticsName" in metric and isinstance(metric["analyticsName"], str):
            metric["analyticsName"] = clean_analytics_name(
                metric["analyticsName"], remove_prefix=remove_prefix
            )
        if "unit" not in metric:
            metric["unit"] = "value"

def process_dictionary_type(attributes):
    if "dictionaryType" in attributes and attributes["dictionaryType"] == "custom":
        attributes["dictionaryType"] = "global"

def process_json(obj, remove_prefix=False):
    if isinstance(obj, dict):
        for key, value in obj.items():
            if key == "metrics" and isinstance(value, list):
                process_metrics(value, remove_prefix=remove_prefix)
            elif key == "attributes" and isinstance(value, dict):
                process_dictionary_type(value)
                process_json(value, remove_prefix=remove_prefix)
            else:
                process_json(value, remove_prefix=remove_prefix)
    elif isinstance(obj, list):
        for item in obj:
            process_json(item, remove_prefix=remove_prefix)

def main():
    if len(sys.argv) < 2:
        print("Usage: python3 process_json.py <input_file> [output_file] [--remove-prefix]")
        sys.exit(1)

    # Detect optional flag
    remove_prefix = False
    args = sys.argv[1:]
    if "--remove-prefix" in args:
        remove_prefix = True
        args.remove("--remove-prefix")

    input_file = args[0]
    output_file = args[1] if len(args) > 1 else None

    try:
        with open(input_file, "r") as f:
            data = json.load(f)
    except FileNotFoundError:
        print(f"Error: File '{input_file}' not found.")
        sys.exit(1)
    except json.JSONDecodeError:
        print(f"Error: File '{input_file}' is not a valid JSON file.")
        sys.exit(1)

    process_json(data, remove_prefix=remove_prefix)

    if output_file:
        try:
            with open(output_file, "w") as f:
                json.dump(data, f, indent=2)
            print(f"Processing complete! Updated JSON written to '{output_file}'")
        except Exception as e:
            print(f"Error writing to output file: {e}")
            sys.exit(1)
    else:
        print(json.dumps(data, indent=2))

if __name__ == "__main__":
    main()
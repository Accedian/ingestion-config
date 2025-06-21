import json
import sys

# Function to convert a string to real camel case
def to_camel_case(value):
    # Split the value by underscores
    words = value.split("_")
    # Convert the first word to lowercase, and capitalize subsequent words
    camel_case = words[0].lower() + "".join(word.capitalize() for word in words[1:])
    return camel_case

# Function to process "analyticsName" values
def process_analytics_name(value):
    return to_camel_case(value)

# Recursively process the JSON structure
def process_json(obj):
    if isinstance(obj, dict):
        for key, value in obj.items():
            if key == "analyticsName" and isinstance(value, str):
                obj[key] = process_analytics_name(value)
            else:
                process_json(value)
    elif isinstance(obj, list):
        for index, item in enumerate(obj):
            process_json(item)

# Main function
def main():
    # Check if the input file is provided
    if len(sys.argv) < 2:
        print("Usage: python3 process_json.py <input_file> [output_file]")
        sys.exit(1)

    # Get the input and optional output file paths
    input_file = sys.argv[1]
    output_file = sys.argv[2] if len(sys.argv) > 2 else None

    # Load the JSON file
    try:
        with open(input_file, "r") as f:
            data = json.load(f)
    except FileNotFoundError:
        print(f"Error: File '{input_file}' not found.")
        sys.exit(1)
    except json.JSONDecodeError:
        print(f"Error: File '{input_file}' is not a valid JSON file.")
        sys.exit(1)

    # Process the JSON data
    process_json(data)

    # Output the results
    if output_file:
        # Write to the output file
        try:
            with open(output_file, "w") as f:
                json.dump(data, f, indent=2)
            print(f"Processing complete! Updated JSON written to '{output_file}'")
        except Exception as e:
            print(f"Error writing to output file: {e}")
            sys.exit(1)
    else:
        # Print to the console
        print(json.dumps(data, indent=2))

# Entry point
if __name__ == "__main__":
    main()
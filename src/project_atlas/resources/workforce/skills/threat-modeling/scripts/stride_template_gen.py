#!/usr/bin/env python3
import argparse
import json


def generate_template(component_name):
    template = {
        "component": component_name,
        "stride_assessment": {
            "Spoofing": {"applicable": True, "threats": [], "mitigations": []},
            "Tampering": {"applicable": True, "threats": [], "mitigations": []},
            "Repudiation": {"applicable": True, "threats": [], "mitigations": []},
            "Information Disclosure": {"applicable": True, "threats": [], "mitigations": []},
            "Denial of Service": {"applicable": True, "threats": [], "mitigations": []},
            "Elevation of Privilege": {"applicable": True, "threats": [], "mitigations": []}
        }
    }
    return template

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Generate a STRIDE threat modeling template for a component.")
    parser.add_argument("component", help="The name of the component being modeled")
    parser.add_argument("-o", "--output", help="Output file", default=None)
    args = parser.parse_args()

    template = generate_template(args.component)

    if args.output:
        with open(args.output, "w") as f:
            json.dump(template, f, indent=4)
        print(f"Template generated at {args.output}")
    else:
        print(json.dumps(template, indent=4))

"""
This module generates the NOTICE.txt and dependencies.csv files
"""

from __future__ import print_function

import argparse
import csv
import json
import os
import subprocess
import tempfile

DEFAULT_BUILD_TAGS = "darwin,linux,windows"

# Get the beats repo root directory, making sure it's downloaded first.
subprocess.run(["go", "mod", "download", "github.com/elastic/beats/..."], check=True)
BEATS_DIR = (
    subprocess.check_output(
        ["go", "list", "-m", "-f", "{{.Dir}}", "github.com/elastic/beats/..."],
    )
    .decode("utf-8")
    .strip()
)

# notice_overrides holds additional overrides entries for go-licence-detector.
notice_overrides = [
    {"name": "github.com/elastic/beats/v7", "licenceType": "Elastic"},
    {"name": "github.com/build-security/beats/v7", "licenceType": "Elastic"},
    {"name": "github.com/golang/glog", "licenceType": "Apache-2.0"},
    {"name": "github.com/spdx/tools-golang", "licenceFile": "LICENSE.code", "licenceType": "Apache-2.0"},
    {"name": "github.com/aquasecurity/trivy-policies", "licenceType": "MIT"},
    {"name": "github.com/csaf-poc/csaf_distribution/v3", "licenceType": "Apache-2.0"},
    {"name": "github.com/xi2/xz", "licenceType": "MIT"},
]

# Additional third-party, non-source code dependencies, to add to the CSV output.
additional_third_party_deps = [
    {
        "name": "Red Hat Universal Base Image minimal",
        "version": "8",
        "url": "https://catalog.redhat.com/software/containers/ubi8/ubi-minimal/5c359a62bed8bd75a2c3fba8",
        "license": "Custom;https://www.redhat.com/licenses/EULA_Red_Hat_Universal_Base_Image_English_20190422.pdf",
        "sourceURL": "https://oss-dependencies.elastic.co/red-hat-universal-base-image-minimal/8/ubi-minimal-8-source"
        ".tar.gz",
    },
]


def read_go_deps(main_packages, build_tags):
    """
    read_go_deps returns a list of module dependencies in JSON format.
    Main modules are excluded; only dependencies are returned.

    Unlike `go list -m all`, this function excludes modules that are only
    required for running tests.
    """
    go_list_args = ["go", "list", "-deps", "-json"]
    if build_tags:
        go_list_args.extend(["-tags", build_tags])
    output = subprocess.check_output(go_list_args + main_packages).decode("utf-8")
    modules_map = {}
    decoder = json.JSONDecoder()
    while True:
        output = output.strip()
        if not output:
            break
        pkg, end = decoder.raw_decode(output)
        output = output[end:]
        if "Standard" in pkg:
            continue
        module = pkg["Module"]
        if "Main" not in module:
            modules_map[module["Path"]] = module
    return sorted(modules_map.values(), key=lambda module: module["Path"])


def go_license_detector(notice_out, deps_out, modules_list):
    """
    go_license_detector runs go-licence-detector to generate the NOTICE.txt and dependencies.csv files.
    @param notice_out: Path to NOTICE.txt file to write.
    @param deps_out: Path to dependencies.csv file to write.
    @param modules_list: List of modules to include in the NOTICE.txt and dependencies.csv files.
    @return: None
    """
    modules_json = "\n".join(map(json.dumps, modules_list))

    beats_deps_template_path = os.path.join(
        BEATS_DIR,
        "dev-tools",
        "notice",
        "dependencies.csv.tmpl",
    )
    beats_notice_template_path = os.path.join(
        BEATS_DIR,
        "dev-tools",
        "notice",
        "NOTICE.txt.tmpl",
    )
    beats_overrides_path = os.path.join(
        BEATS_DIR,
        "dev-tools",
        "notice",
        "overrides.json",
    )
    beats_rules_path = os.path.join(BEATS_DIR, "dev-tools", "notice", "rules.json")

    with (
        open(beats_notice_template_path) as beats_notice_template,
        open(beats_overrides_path) as beats_overrides,
        tempfile.TemporaryDirectory() as tmpdir,
    ):
        notice_template_contents = beats_notice_template.read()
        overrides_contents = beats_overrides.read()

        # Create notice overrides.json by combining the overrides from beats with cloudbeat specific ones.
        with open(
            os.path.join(tmpdir, "overrides.json"),
            "w",
            encoding="utf8",
        ) as overrides_file:
            overrides_file.write(overrides_contents)
            overrides_file.write("\n")
            for entry in notice_overrides:
                overrides_file.write("\n")
                json.dump(entry, overrides_file)
            overrides_file.close()

            # Replace "Elastic Beats" with "Elastic Cloudbeat" in the NOTICE.txt template.
            with open(
                os.path.join(tmpdir, "NOTICE.txt.tmpl"),
                "w",
                encoding="utf8",
            ) as notice_template_file:
                notice_template_file.write(
                    notice_template_contents.replace("Elastic Beats", "Elastic Cloudbeat"),
                )
                notice_template_file.close()

                args_list = [
                    "go",
                    "run",
                    "-modfile=go.mod",
                    "go.elastic.co/go-licence-detector",
                    "-includeIndirect",
                    "-overrides",
                    overrides_file.name,
                    "-rules",
                    beats_rules_path,
                    "-noticeTemplate",
                    notice_template_file.name,
                    "-depsTemplate",
                    beats_deps_template_path,
                    "-noticeOut",
                    notice_out,
                    "-depsOut",
                    deps_out,
                ]
                subprocess.run(
                    args_list,
                    check=True,
                    input=modules_json.encode("utf-8"),
                )


def write_notice_file(notice_out_name, modules_list):
    """
    write_notice_file writes the NOTICE.txt file.
    @param notice_out_name: Path to NOTICE.txt file to write.
    @param modules_list: List of modules to include in the NOTICE.txt file.
    @return: None
    """
    go_license_detector(notice_out_name, "", modules_list)


def write_csv_file(csv_filename, modules_list):
    """
    write_csv_file writes a CSV file containing the third-party dependencies.
    @param csv_filename: Path to CSV file to write.
    @param modules_list: List of modules to include in the CSV file.
    @return: None
    """
    with tempfile.TemporaryDirectory() as tmpdir:
        tmp_deps_path = os.path.join(tmpdir, "dependencies.csv")
        go_license_detector("", tmp_deps_path, modules_list)
        with open(tmp_deps_path) as csvfile:
            reader = csv.DictReader(csvfile)
            fieldnames = reader.fieldnames
            rows = list(reader)
        with open(csv_filename, "w") as csvfile:
            writer = csv.DictWriter(csvfile, fieldnames)
            writer.writeheader()
            for row in rows:
                writer.writerow(row)
            for dep in additional_third_party_deps:
                writer.writerow(dep)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Generate the NOTICE file from package dependencies",
    )
    parser.add_argument(
        "main_package",
        nargs="*",
        default=["."],
        help="List of main Go packages for which dependencies should be processed",
    )
    parser.add_argument("--csv", dest="csvfile", help="Output to a csv file")
    parser.add_argument(
        "--build-tags",
        default=DEFAULT_BUILD_TAGS,
        help="Comma-separated list of build tags to pass to 'go list -deps'",
    )
    args = parser.parse_args()
    modules = read_go_deps(args.main_package, args.build_tags)

    if args.csvfile:
        write_csv_file(args.csvfile, modules)
        print(args.csvfile)
    else:
        notice_filename = os.path.abspath("NOTICE.txt")
        write_notice_file(notice_filename, modules)
        print(notice_filename)

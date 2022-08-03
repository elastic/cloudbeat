import _ from "underscore";
import fs from 'fs';
import config from 'config';
import xlsx from 'node-xlsx';
import path from 'path';
import YAML from 'yaml';

const {v5: uuid} = require('uuid');

const output_folder: string = config.get("output_folder");
const benchmarks_folder: string = config.get("benchmarks_folder");

function generateOutputFolder(): void {
    console.log("Creating output folder:", output_folder);
    if (fs.existsSync(output_folder)) fs.rmSync(output_folder, {recursive: true});
    fs.mkdirSync(output_folder);
}

function parseReferences(references: string): string[] {
    if (!references) {
        return [];
    }

    return references.replaceAll(":http", "\nhttp").split("\n");
}

// Generate a mapping between section # and its title.
// For example: { 'section #': '1', title: 'Control Plane Components' }
function parseAllSections(data: BenchmarksData[]): SectionMetadata[] {
    return data.filter((el) => (el["section #"] && el["title"] && !el["recommendation #"]))
        .map(i => _.pick(i, ["section #", "title"]));
}

function identifySection(rule_section: string, sections: SectionMetadata[]): string {
    for (const section of sections) {
        if (section["section #"] == rule_section) {
            return section.title;
        }
    }
    // we should never get here!
    console.log("FATAL: Could not find section for rule", rule_section);
    process.exit(-1);
}

// Adds newline character after code-blocks that are at the end of a property
function fixCodeBlocks(val: string): string {
    if (val.endsWith("```")) {
        return val.concat("\n");
    }
    return val;
}

function normalizeResults(data: BenchmarksData[], benchmark_metadata: BenchmarkMetadata,
                          profile_applicability: string): RuleSchema[] {
    const sections = parseAllSections(data);

    let result = data.filter((it) => {
        return Boolean(it["recommendation #"]);
    });

    return result.map((it) => {
            const rule_name = it["title"];
            console.log("Parsing:", benchmark_metadata.name, rule_name);
            const refs = parseReferences(it["references"])
            return {
                "id": uuid(`${benchmark_metadata.name} ${rule_name}`, config.get("uuid_seed")),
                "name": rule_name,
                "rule_number": it["recommendation #"],
                "profile_applicability": `* ${profile_applicability}`,
                "description": it["description"],
                // @ts-ignore
                "rationale": fixCodeBlocks(it["rational statement"] || it["rationale statement"] || ""),
                "audit": fixCodeBlocks(it["audit procedure"] || ""),
                "remediation": fixCodeBlocks(it["remediation procedure"] || ""),
                "impact": it["impact statement"] || "",
                // "default_value": "By default, profiling is enabled.\n", // TODO
                "references": refs,
                "section": identifySection(it["section #"], sections),
                "benchmark": {"name": benchmark_metadata.name, "version": benchmark_metadata.version},
            }
        }
    );
}

function parseSpreadsheet(tab: SpreadsheetTab, benchmark_metadata: BenchmarkMetadata): RuleSchema[] {
    const profile_applicability = tab.name;
    const results: BenchmarksData[] = [];
    const keys = tab.data[0].map(el => el.toLowerCase()); // Different benchmarks have different casing in the columns titles
    for (let idx = 1; idx < tab.data.length; idx++) {
        const values = tab.data[idx];
        // `keys` is an array that holds all the column names
        // `values` is an array that holds the cell values
        // The following line will push into results an object that is look something like:
        // {key1: value1, key2: value2, key3: value3, ...}
        const benchmark_data = keys.map((v: string, i: number) => ({
            [v]: values[i]?.replace(/\r\n/g, "\n")
        }));
        // @ts-ignore
        results.push(Object.assign.apply({}, benchmark_data));
    }

    return normalizeResults(results, benchmark_metadata, profile_applicability);
}

function parseBenchmark(file: string, benchmark_metadata: BenchmarkMetadata): RuleSchema[] {
    const excel = xlsx.parse(file, {raw: false, type: 'file', cellText: true});
    // Assumption, we treat only tabs that start with the word "Level" (as in the string "Level 1 - Master Node")
    return excel.filter(tab => Boolean(tab.name.indexOf("Level") == 0))
        .map(tab => parseSpreadsheet(tab as SpreadsheetTab, benchmark_metadata))
        .flat();
}

function parseBenchmarks(folder: string): BenchmarkSchema[] {
    const files = fs.readdirSync(folder);
    return files.map(file => {
        const file_path = folder + "/" + file;
        const filename = path.parse(file).name;
        const tokens = filename.split("_");
        const pivot = tokens.indexOf("Benchmark");
        const benchmark: BenchmarkMetadata = {
            "name": tokens.slice(0, pivot).join(" "), // assuming the "Benchmark" word separates the benchmark name and the benchmark version
            "version": tokens.slice(-1)[0]            // assuming the version is always the last token in the string
        }

        return {
            "filename": filename,
            "metadata": benchmark,
            "rules": parseBenchmark(file_path, benchmark)
        }
    });
}

function generateOutputFiles(benchmarks: BenchmarkSchema[]): void {
    const result: any = {
        "policies": {}
    };

    for (const benchmark of benchmarks) {
        console.log("Parsed total of", benchmark.rules.length, "rules in benchmark", benchmark.filename);
        result.policies[benchmark.filename] = {};
        for (let rule of benchmark.rules) {
            result.policies[benchmark.filename][rule.rule_number] = rule;
        }
        fs.writeFileSync(output_folder + "/" + benchmark.filename + ".yaml", YAML.stringify(benchmark.rules));
    }
    fs.writeFileSync(output_folder + "/" + config.get("output_filename"), YAML.stringify(result));
}

function main(): void {
    // Make sure output folder exists an is empty
    generateOutputFolder();

    const parsed_benchmarks = parseBenchmarks(benchmarks_folder)
    generateOutputFiles(parsed_benchmarks);
    console.log("Done!");
}

main()
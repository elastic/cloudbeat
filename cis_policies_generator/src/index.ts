import _ from "underscore";
import fs from 'fs';
import config from 'config';
import xlsx from 'node-xlsx';
import path from 'path';
import YAML from 'yaml';
const program = require('commander');
import {FixBrokenReferences} from "./fixBrokenReferences";
import {generateRuleTemplates} from "./generateRuleTemplates";
import {BENCHMARK_TYPES} from "./constants";

const {v5: uuid} = require('uuid');
const benchmarks_folder: string = config.get("benchmarks_folder");

program
    .name('cis-policies-generator')
    .description('CIS Benchmark parser CLI')
    .version('0.0.1')
    .addHelpText('after', `
    Example calls:
        $ npm start -- -m
        $ npm start -- -t -b 'cis_eks'`);

program
    .option('-t, --templates', 'generate csp rule templates and place them in the integration dir')
    .option('-b, --benchmark <string>', 'benchmark to be used for the rules template generation', 'cis_k8s')
    .option('-m, --rulesMeta', 'generate rules metadata for any provided benchmark in the input dir')

program.parse();
const options = program.opts();
const shouldGenRuleTemplates = options.templates;
const shouldGenRulesMeta = options.rulesMeta;
const rulesTemplateBenchmark = options.benchmark;

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
            const rule_number = it["recommendation #"];
            const rule_section = identifySection(it["section #"], sections);

            console.log("Parsing:", benchmark_metadata.name, rule_name);
            const refs = parseReferences(it["references"])
            return {
                "id": uuid(`${benchmark_metadata.name} ${rule_name}`, config.get("uuid_seed")),
                "name": rule_name,
                "rule_number": rule_number,
                "profile_applicability": `* ${profile_applicability}`,
                "description": it["description"],
                "version": "1.0",
                // @ts-ignore
                "rationale": fixCodeBlocks(it["rational statement"] || it["rationale statement"] || ""),
                "audit": fixCodeBlocks(it["audit procedure"] || ""),
                "remediation": fixCodeBlocks(it["remediation procedure"] || ""),
                "impact": it["impact statement"] || "",
                // "default_value": "By default, profiling is enabled.\n", // TODO: retrieve default_value straight from CIS
                "references": refs,
                "tags": constructRuleTags(benchmark_metadata, rule_number, rule_section),
                "section": rule_section,
                "benchmark": {
                    "name": benchmark_metadata.name,
                    "version": benchmark_metadata.version,
                    "id": getBenchmarkAttr(benchmark_metadata, "id")
                },
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
            "version": tokens.slice(-1)[0],            // assuming the version is always the last token in the string
            "filename": filename,
        }

        return {
            "metadata": benchmark,
            "rules": parseBenchmark(file_path, benchmark)
        }
    });
}

function generateRulesMetadataFiles(benchmarks: BenchmarkSchema[]): void {
    for (const benchmark of benchmarks) {
        console.log("Parsed total of", benchmark.rules.length, "rules in benchmark", benchmark.metadata.filename);
        const benchmark_id = getBenchmarkAttr(benchmark.metadata, "id")
        for (let rule of benchmark.rules) {
            const ruleNumber = rule.rule_number!.replaceAll(".", "_");
            const rule_folder = `../bundle/compliance/${benchmark_id}/rules/cis_${ruleNumber}`
            const metadata_file = rule_folder + "/data.yaml";

            if (fs.existsSync(rule_folder)) {
                _.assign(rule, getExistingValues(metadata_file));
                fs.writeFileSync(metadata_file, YAML.stringify({metadata: rule} as MetadataFile));
            }
        }
    }
}

function constructRuleTags(benchmark_metadata: BenchmarkMetadata, rule_number: string, section: string): string[] {
    const benchmark_type = getBenchmarkAttr(benchmark_metadata, "tag")
    return ["CIS", benchmark_type, "CIS " + rule_number, section].filter(Boolean)
}

function getBenchmarkAttr(benchmark_metadata: BenchmarkMetadata, field: string): string {
    for (let [type, attr] of Object.entries(BENCHMARK_TYPES)) {
        if (benchmark_metadata.name.includes(type)) {
            // @ts-ignore
            return attr[field];
        }
    }

    return ""
}

function getExistingValues(filePath: string): Partial<RuleSchema> {
    if (!fs.existsSync(filePath)) {
        return {};
    }

    const file = fs.readFileSync(filePath, 'utf8');
    const rule = YAML.parse(file).metadata as RuleSchema;

    // Remove falsy attributes if there's any
    return _.pick({
        default_value: rule.default_value,
        id: rule.id
    }, (prop: any) => prop)
}

async function main(): Promise<void> {
    if (shouldGenRulesMeta) {
        const parsed_benchmarks = parseBenchmarks(benchmarks_folder);
        await FixBrokenReferences(parsed_benchmarks);

        console.log("Generate rules metadata");
        generateRulesMetadataFiles(parsed_benchmarks);
    }

    if (shouldGenRuleTemplates) {
        console.log("Generate CSP rule templates");
        generateRuleTemplates(rulesTemplateBenchmark);
    }

    console.log("Done!");
}

main().then(r => r)
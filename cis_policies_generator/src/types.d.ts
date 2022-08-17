interface BenchmarkMetadata {
    name: string;
    version: string;
    filename: string;
}

interface BenchmarkInfo {
    name: string;
    version: string;
    id: string;
}

interface CisEksBenchmarkSchema {
    "section #": string;
    "recommendation #": string;
    "title": string;
    "description": string;
    "rationale statement": string;
    "impact statement": string;
    "remediation procedure": string;
    "audit procedure": string;
    "references": string;
}

interface CisK8sBenchmarkSchema {
    "section #": string;
    "recommendation #": string;
    "title": string;
    "description": string;
    "rational statement": string;
    "impact statement": string;
    "remediation procedure": string;
    "audit procedure": string;
    "references": string;
}

type BenchmarksData = CisK8sBenchmarkSchema | CisEksBenchmarkSchema;

interface SectionMetadata {
    "section #": string;
    "title": string;
}

interface BenchmarkSchema {
    metadata: BenchmarkMetadata;
    rules: RuleSchema[];
}

interface SpreadsheetTab {
    name: string;
    data: string[][];
}

interface HttpCache {
    [key: string]: number
}

interface MetadataFile {
    metadata: RuleSchema
}

interface RuleSchema {
    audit: string;
    rule_number?: string;
    benchmark: BenchmarkInfo;
    default_value?: string;
    description: string;
    id: string;
    impact: string;
    name: string;
    profile_applicability: string;
    rationale: string;
    references: string[];
    remediation: string;
    section: string;
    tags: string[];
    version: string;
}

interface CspRuleTemplate {
    attributes: CspRuleTemplateAttr;
    id: string;
    type: string;
    migrationVersion: object;
    coreMigrationVersion: string;
}

interface CspRuleTemplateAttr {
    enabled: boolean;
    muted: boolean;
    metadata: RuleSchema;
}

interface BenchmarkAttributes {
    id: string,
    tag: string
}
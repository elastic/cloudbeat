"use strict";

const pdfjsLib = require("pdfjs-dist/legacy/build/pdf.js");
const Worker = require("./worker.js");
const _ = require("lodash");
const YAML = require("yaml");
const fs = require("fs");
const path = require("path")
const papa = require("papaparse");
const { Command } = require('commander');

const cspPath = "../csp-security-policies"
const BENCHMARK = "Kubernetes"; //"Kubernetes"; // EKS
//   "./CIS_Amazon_Elastic_Kubernetes_Service_(EKS)_Benchmark_v1.0.1.pdf"; //"./CIS_Kubernetes_Benchmark_v1.6.0.pdf";
const w = new Worker({benchmark: BENCHMARK});
const program = new Command();

program
  .name('cis-benchmark-pdf-serialization')
  .description('CLI CIS Benchmark parser')
  .version('0.0.1');

program
  .option('-p, --pdfpath <string>')
  .option('-r, --reportpath <string>')

program.parse();
const options = program.opts();
const pdfPath = options.pdfpath;
const reportPath = options.reportpath;

try {
    (async function () {
        const missingRules = [];
        const loadingTask = pdfjsLib.getDocument(pdfPath);
        const doc = await loadingTask.promise;
        const numPages = doc.numPages;
        console.log("# Document Loaded");
        console.log("Number of Pages: " + numPages);
        const B_PAGE = 17; //13; // 254; // 260; //17;
        const E_PAGE = 262; // 132; //255; // 261; //18;
        for (let pageNum = B_PAGE; pageNum <= E_PAGE; pageNum++) {
            const page = await doc.getPage(pageNum);
            const content = await page.getTextContent();
            // console.log("# Page " + pageNum);
            w.addItems(content.items, pageNum);
            page.cleanup();
        }
        const ruleObjs = w.Run();
        // if (fs.existsSync("./yamls")) fs.rmSync("./yamls", { recursive: true });
        // fs.mkdirSync("./yamls");
        ruleObjs.forEach((ruleObj) => {
            // console.log(ruleObj);
            const r_number = ruleObj.metadata.tags[2]
                .split(" ")[1]
                .replaceAll(".", "_");
            // ruleObj = _.omit(ruleObj, "metadata.tags");
            const yaml_doc = new YAML.Document();
            // console.log(doc.toString());
            if (isRuleImplemented(r_number)) {
                if (!fs.existsSync(path.join(cspPath, `compliance/cis_k8s/rules/cis_${r_number}/data.yaml`))) {
                    console.log("JENIA NEW RULE", r_number);
                    yaml_doc.contents = ruleObj;
                    // console.log("content")
                    // console.log(yaml_doc.contents)
                    fs.writeFileSync(path.join(cspPath, `compliance/cis_k8s/rules/cis_${r_number}/data.yaml`), yaml_doc.toString());
                    console.log("after write")
                }
            } else {
                missingRules.push({number: r_number, name: ruleObj.metadata.name})
            }
        });
        
        if (reportPath) {
            const filename = "missing_rules.csv"
            const columns = ["number", "name"]
            let csv = papa.unparse({ data: missingRules, fields: columns});
            if (csv == null) return;
            
            fs.writeFile(path.join(reportPath, filename), csv, (err) => {
                if (err) throw err;
                console.log('The file has been saved!');
            });
        }
    })();
} catch (error) {
    console.error(err);
}

function isRuleImplemented(r_number) {
    return fs.existsSync(path.join(cspPath, `/compliance/cis_k8s/rules/cis_${r_number}`))
}
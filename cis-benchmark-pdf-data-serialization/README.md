# cis-benchmark-pdf-data-serialization
 
The cis-benchmark serialization tool is an internal tool meant for creating the cis benchmark rule metadata.

## Getting started 

### Prerequisites
- node.js version v17.6.0 and above

### Running the project

Install node dependencies `npm install`

Go the `main.js` and edit the following:
- Download the `CIS_Kubernetes_Benchmark_v1.6.0.pdf` file.
- Set the `pdfPath` to your `CIS_Kubernetes_Benchmark_v1.6.0.pdf` file path.
- **Optional** set the B_Page/E_page to the first/last rule page you want to convert.

Project CLI: `node main.js --help`
```
Usage: cis-benchmark-pdf-serialization [options]

CLI CIS Benchmark parser

Options:
  -V, --version              output the version number
  -p, --pdfpath <string>
  -r, --reportpath <string>
  -h, --help                 display help for command
  
Example:
  node main.js -p <BENCHMARK_PATH> -r <DIR_TO_STORE_REPORT>
```

### Guidelines on how to process generated rules

There is no generation of markdown for the field values (remains a gap).
This means that the metadata needs to be manually converted after generation.

Here is the list of common things that needs to be addressed when manually editing the generated data:
- In some cases there are problems with spaces when concating strings (heuristic needs to be improved).
  Sometimes the name field strings are getting concated with a missing space between words.
  Like this for example: 
  Ensure that the --kubelet-client-certificate and --kubelet-client-keyarguments are set as appropriate (Automated).
  Notice that `--kubelet-client-keyarguments` doesn't have a space and should be `--kubelet-client-key arguments`.
  This can happen not only in names, but across other values as well.
- profile_applicability field needs to be converted from `• Level 1 - Master Node` to `* Level 1 - Master Node`
  or any other fields that have `•`, like bullets within the values should be converted to `*`.
- Notice that there are several ways to assign values to a property in YAMLs (`|`, `>`).
  `|` is mainly used for multiline values and is essential when you have codeblock inside the value.
  `>` is mainly used for singleline values and it adds `\n` for every row end at the YAML, this option also supports inline code.
- Code blocks must have newline/`\n` prior to every \`\`\` and newline/`\n` after every \`\`\`.
- Some rules might have bold values inside the value of the rule that will need to be converted with `**VALUE**`.
  Example of such case are rules 1.2.15, 2.4, 2.5, 2.6, notice that they have `**Note**`.
- Some rules do not have all of the properties and if that's the case then you will need to add empty properties, because ignoring that will cause the tests to fail.
  One of such cases is rule 5.1.3 that doesn't have properties `default_value`, `references`.
- There are also instances of EKS rules having `Audit1`, `Audit2` properties which are being parsed but not picked by the script.

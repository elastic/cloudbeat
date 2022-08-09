"use strict";

const _ = require("lodash");
const uuidv5 = require("uuid").v5;

const START_OF_COMMON_RULE_REGEX = new RegExp(`^[0-9]\.[0-9]+\.[0-9]+ `);
const START_OF_KUBERNETES_RULE_REGEX = new RegExp(
  `^[0-9]\.[0-9]+\.[0-9]+ |^2\.[0-9]+ `
);
const END_OF_RULE_REGEX = new RegExp("^CIS Controls:$");
const MONOBLOCK_FONT_REGEX = new RegExp("^g_d0_f5$|^g_d0_f6$");
const RULE_PROP_FONT_REGEX = new RegExp("^g_d0_f4$");
const RULE_PROP_KEY_REGEX = new RegExp(":$");

const Y_CAP = 1000;
const WORKER_NAMESPACE = "5d8d0dd5-acd2-4c46-b565-aa1fb03617af";

const PICK_RULE_PROPERTIES = [
  "id",
  "name",
  "profile_applicability",
  "description",
  "rationale",
  "audit",
  "remediation",
  "impact",
  "default_value",
  "references",
  "section",
  "tags",
  "benchmark",
];

const KUBERNETES_RULE_CATEGORIES = {
  1.1: "Master Node Configuration Files",
  1.2: "API Server",
  1.3: "Controller Manager",
  1.4: "Scheduler",
  2: "etcd",
  3.1: "Authentication and Authorization",
  3.2: "Logging",
  4.1: "Worker Node Configuration Files",
  4.2: "Kubelet",
  5.1: "RBAC and Service Accounts",
  5.2: "Pod Security Policies",
  5.3: "Network Policies and CNI",
  5.4: "Secrets Management",
  5.5: "Extensible Admission Control",
  5.7: "General Policies",
};

const EKS_RULE_CATEGORIES = {
  2.1: "Logging",
  3.1: "Worker Node Configuration Files",
  3.2: "Kubelet",
  4.1: "RBAC and Service Accounts",
  4.2: "Pod Security Policies",
  4.3: "CNI Plugin",
  4.4: "Secrets Management",
  4.5: "Extensible Admission Control",
  4.6: "General Policies",
  5.1: "Image Registry and Image Scanning",
  5.2: "Identity and Access Management (IAM)",
  5.3: "AWS Key Management Service (KMS)",
  5.4: "Cluster Networking",
  5.5: "Authentication and Authorization",
  5.6: "Other Cluster Configurations",
};

const KUBERNETES_BENCHMARK_METADATA = {
  name: "CIS Kubernetes V1.20",
  version: "v1.0.0",
};

const EKS_BENCHMARK_METADATA = {
  name: "CIS Amazon Elastic Kubernetes Service (EKS) Benchmark",
  version: "v1.0.1",
};

const BENCHMARK_TYPE_TO_INFO = {
  Kubernetes: {
    CATEGORIES: KUBERNETES_RULE_CATEGORIES,
    BENCHMARK_METADATA: KUBERNETES_BENCHMARK_METADATA,
    START_OF_RULE_REGEX: START_OF_KUBERNETES_RULE_REGEX,
  },
  EKS: {
    CATEGORIES: EKS_RULE_CATEGORIES,
    BENCHMARK_METADATA: EKS_BENCHMARK_METADATA,
    START_OF_RULE_REGEX: START_OF_COMMON_RULE_REGEX,
  },
};

class Worker {
  constructor(params) {
    this.benchmark = params.benchmark;
    this.items = [];
    this.ruleItems = [];
  }

  addItems(items, page) {
    this.items = this.items.concat(
      _.forEach(
        _.filter(
          // Removing first 5 items that indicate page
          _.drop(items, 5),
          (item) => item.width && item.height && !_.isEmpty(item.str)
        ),
        (value, key) => {
          value.page = page;
        }
      )
    );
  }

  groupItemsByRule() {
    const lines = _.reduce(
      // Sort by pages and item's Y in PDF
      _.groupBy(this.items, (item) => -item.page * Y_CAP + item.transform[5]),
      (result, value, key) => {
        result.push(value);
        return result;
      },
      []
    );
    this.ruleItems = _.reduce(
      lines,
      (result, value, key) => {
        if (
          _.find(value, (obj) =>
            BENCHMARK_TYPE_TO_INFO[this.benchmark].START_OF_RULE_REGEX.test(
              obj.str
            )
          )
        ) {
          result.push([value]);
        } else if (_.find(value, (obj) => END_OF_RULE_REGEX.test(obj.str))) {
          result[result.length - 1].push(value);
        } else {
          const last = result.length && result[result.length - 1];
          if (last && !END_OF_RULE_REGEX.test(last[last.length - 1].str)) {
            result[result.length - 1].push(value);
          }
        }
        return result;
      },
      []
    );
  }

  parseRuleToObject(items) {
    return _.pick(
      _.defaults(
        _.reduce(
          _.reduce(
            items.slice(_.findIndex(items, (item) => this.isRuleKeyLine(item))),
            (result, value, key) => {
              if (this.isRuleKeyLine(value)) {
                result.push([value]);
              } else {
                // Ignore everything not under rule
                if (result.length) result[result.length - 1].push(value);
              }
              return result;
            },
            []
          ),
          (result, value, key) => {
            result[this.rulePropLineToKey(value[0])] = this.ruleLinesToValue(
              value.slice(1)
            );
            return result;
          },
          {}
        ),
        this.buildCustomProp(items)
      ),
      PICK_RULE_PROPERTIES
    );
  }

  ruleLinesToValue(lines) {
    return _.reduce(
      lines,
      (result, value, key) => {
        result += this.buildLine(value) + "\n";
        return result;
      },
      ""
    );
  }

  rulePropLineToKey(prop_line) {
    return this.buildLine(prop_line)
      .toLowerCase()
      .replaceAll(" ", "_")
      .replaceAll(":", "");
  }

  // buildKeyLine(line) {
  //   let key = line.map((ln) => ln.str).join(" ");
  //   // Rule 1.2.4 devides CIS Controls property to two lines (it is a one liner in the PDF)
  //   // if (String(key) === "CISControls:") key = "CIS Controls:";
  //   // console.log(line);
  //   return key;
  // }

  buildLine(line) {
    let l_width = 0;
    return _.reduce(
      line,
      (result, value, key) => {
        // Decide according to width calculation difference between items
        if (l_width && value.transform[4] - l_width > 1) {
          result += ` ${value.str}`;
        } else {
          result += value.str;
        }
        l_width = value.width + value.transform[4];
        return result;
      },
      ""
    );
    // return line.map((ln) => ln.str).join(" ");
  }

  isRuleKeyLine(line) {
    return (
      RULE_PROP_KEY_REGEX.test(this.buildLine(line)) &&
      RULE_PROP_FONT_REGEX.test(line[0].fontName)
    );
  }

  buildCustomProp(items) {
    const rn_index = _.findIndex(items, (item) =>
      BENCHMARK_TYPE_TO_INFO[this.benchmark].START_OF_RULE_REGEX.test(
        item[0].str
      )
    );
    const t_name = this.buildLine(
      _.flatten(
        items.slice(
          rn_index,
          _.findIndex(items, (item) => this.isRuleKeyLine(item), rn_index)
        )
      )
    );
    const r_number = t_name.split(" ")[0];
    const name = t_name.split(" ").slice(1).join(" ");
    const section =
      BENCHMARK_TYPE_TO_INFO[this.benchmark].CATEGORIES[
        r_number.substr(0, r_number.lastIndexOf("."))
      ];
    return {
      name,
      id: uuidv5(`${this.benchmark} ${t_name}`, WORKER_NAMESPACE),
      section,
      tags: ["CIS", this.benchmark, "CIS " + r_number, section],
      benchmark: BENCHMARK_TYPE_TO_INFO[this.benchmark].BENCHMARK_METADATA,
    };
  }

  Run() {
    // console.log(this.items);
    this.groupItemsByRule();
    // // return this.parseRuleToObject(this.ruleItems[0]);
    return this.ruleItems.map((rule) => ({
      metadata: this.parseRuleToObject(rule),
    }));
    // // this.ruleItems.forEach((rule) => this.parseRuleToObject(rule));
    // // return this.parseRuleToObject(this.ruleItems[0]);
  }
}

module.exports = Worker;

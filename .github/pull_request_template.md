<!-- Thanks for submitting a pull request! We appreciate you spending the time to work on these changes. Please provide enough information so that others can review your pull request. The three fields below are mandatory. -->

### Summary of your changes
<!--
Please provide a detailed description of the changes introduced by this Pull Request.
Provide a description of the main changes, as well as any additional information the code reviewer should be aware of before beginning the review process.
-->

### Related Issues
<!--
- Related: https://github.com/elastic/security-team/issues/
- Fixes: https://github.com/elastic/security-team/issues/
-->

### Checklist

- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] I have added the necessary README/documentation (if appropriate)

#### Introducing a new rule?

- [ ] Generate rule metadata using [this script](https://github.com/elastic/csp-security-policies/tree/main/dev#generate-rules-metadata)
- [ ] Add relevant unit tests
- [ ] Generate relevant rule templates using [this script](https://github.com/elastic/csp-security-policies/tree/main/dev#generate-rule-templates), and open a PR in [elastic/packages/cloud_security_posture](https://github.com/elastic/integrations/tree/main/packages/cloud_security_posture)

> **Note** Make sure to bump cloudbeat [`version/policy.go`](https://github.com/elastic/cloudbeat/blob/main/version/policy.go#L20) after merging this PR and drafting a new release.

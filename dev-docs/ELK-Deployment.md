# Elastic Stack Deployment
[What is the ELK Stack?](https://www.elastic.co/what-is/elk-stack)
There are recommended ways to deploy the stack for development and testing purposes.

## Elastic Package Deployment

[elastic-package](https://github.com/elastic/elastic-package) - a tool that spins up en entire elastic stack locally.

(you may need to [authenticate](https://docker-auth.elastic.co/github_auth))

For example, spinning up 8.6.0 stack locally:

- Load stack environment variables
  ```zsh
  eval "$(elastic-package stack shellinit --shell $(basename $SHELL))"
  ```
- Spin up the 8.6.0 stack
  ```zsh
  elastic-package stack up --version 8.6.0 -v -d
  ```

## Elastic Cloud Deployment
This is the recommended way to deploy the stack for development and testing purposes. As it dosn't require any local resources, configuration, and maintenance.
You can just spin up a new deployment, and connect to it:

Spin up Elastic stack using [cloud](https://cloud.elastic.co/home) or [staging](https://staging.found.no/home) (which will be deleted after 14 days)

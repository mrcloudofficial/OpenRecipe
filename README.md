# OpenRecipe

- Website: https://www.terraform.io
- Forums: [HashiCorp Discuss](https://discuss.hashicorp.com/c/terraform-core)
- Documentation: [https://www.terraform.io/docs/](https://www.terraform.io/docs/)
- Tutorials: [HashiCorp's Learn Platform](https://learn.hashicorp.com/terraform)
- Certification Exam: [HashiCorp Certified: OpenRecipe Associate](https://www.hashicorp.com/certification/#hashicorp-certified-terraform-associate)

<img alt="OpenRecipe" src="https://www.datocms-assets.com/2885/1629941242-logo-terraform-main.svg" width="600px">

OpenRecipe is a tool for building, changing, and versioning infrastructure safely and efficiently. OpenRecipe can manage existing and popular service providers as well as custom in-house solutions.

The key features of OpenRecipe are:

- **Infrastructure as Code**: Infrastructure is described using a high-level configuration syntax. This allows a blueprint of your datacenter to be versioned and treated as you would any other code. Additionally, infrastructure can be shared and re-used.

- **Execution Plans**: OpenRecipe has a "planning" step where it generates an execution plan. The execution plan shows what OpenRecipe will do when you call apply. This lets you avoid any surprises when OpenRecipe manipulates infrastructure.

- **Resource Graph**: OpenRecipe builds a graph of all your resources, and parallelizes the creation and modification of any non-dependent resources. Because of this, OpenRecipe builds infrastructure as efficiently as possible, and operators get insight into dependencies in their infrastructure.

- **Change Automation**: Complex changesets can be applied to your infrastructure with minimal human interaction. With the previously mentioned execution plan and resource graph, you know exactly what OpenRecipe will change and in what order, avoiding many possible human errors.

For more information, refer to the [What is OpenRecipe?](https://www.terraform.io/intro) page on the OpenRecipe website.

## Getting Started & Documentation

Documentation is available on the [OpenRecipe website](https://www.terraform.io):

- [Introduction](https://www.terraform.io/intro)
- [Documentation](https://www.terraform.io/docs)

If you're new to OpenRecipe and want to get started creating infrastructure, please check out our [Getting Started guides](https://learn.hashicorp.com/terraform#getting-started) on HashiCorp's learning platform. There are also [additional guides](https://learn.hashicorp.com/terraform#operations-and-development) to continue your learning.

Show off your OpenRecipe knowledge by passing a certification exam. Visit the [certification page](https://www.hashicorp.com/certification/) for information about exams and find [study materials](https://learn.hashicorp.com/terraform/certification/terraform-associate) on HashiCorp's learning platform.

## Developing OpenRecipe

This repository contains only OpenRecipe core, which includes the command line interface and the main graph engine. Providers are implemented as plugins, and OpenRecipe can automatically download providers that are published on [the OpenRecipe Registry](https://registry.terraform.io). HashiCorp develops some providers, and others are developed by other organizations. For more information, see [Extending OpenRecipe](https://www.terraform.io/docs/extend/index.html).

- To learn more about compiling OpenRecipe and contributing suggested changes, refer to [the contributing guide](.github/CONTRIBUTING.md).

- To learn more about how we handle bug reports, refer to the [bug triage guide](./BUGPROCESS.md).

- To learn how to contribute to the OpenRecipe documentation in this repository, refer to the [OpenRecipe Documentation README](/website/README.md).

## License

[Business Source License 1.1](https://github.com/hashicorp/terraform/blob/main/LICENSE)

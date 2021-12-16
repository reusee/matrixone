- Status: draft
- Start Date: 2021-12-15
- Authors: Yuesheng Li
- Implementation PR:
- Issue for this RFC:

# Summary
This proposal describes a framework to test the Matrixcube and Matrixone with chaos method.

# Motivation
Chaos testing or chaos engineering is a powerful tool to improve a system's resilience to different faults.
With simulating a large amount of scenario in the development cycle, more potential bugs can be exposed rather than in the production environment that may cause severe consequences.

# Technical Design

The framework includes a helper library and multiple project specific testing programs.

## The Library

The library includes several modules coresponding to different stages of the testing process.

### Random configuration generation

To introduce randomness to the testing process, a configuraion is generated for every run.
A configuration includes arguments of the running environment and actions to apply.

Arguments includes configurations of the system, usually provided as a file or structure.
Hardware or container specifications are also included in arguments. 

Actions includes normal operations that the system can do. 
Faults also take place in actions as individual steps, instead of being randomly injected at running time.

Every configuration has a unique id.
Configurations can be saved to files, and committed to source repos. 
They can be run repeatly, be updated as project evolves. 

### Setup and execution

This module setups and executes tests according to configurations.

Multiple testing instances can run in parallel.

### Model validation

During the execution, logs and history will be collected.
In this module, logs and history will be feed to different model validation tools.

### Report generation

If one testing instance failed in validation stage, a report about the instance will be generated.
The report includes the id of configuration being used, the target program version (usually the commit hash), and reasons to validation failures.

## Testing Programs

The library described above provides building blocks for testing.
A standalone program must be written to glue the library and project-specific informations.

These should be provided in the program
* configuration items of the target project
* basic actions the project can do
* compound actions to be schedule randomly in configuration generation stage
* how to setup the testing environment
* what validation model to use

# Drawbacks

TODO

# Rationale / Alternatives

This framework is heavily inspired by these methods or projects

* fuzz testing, for generate random input before execution
* jepsen, for history collection and model validation

# Unresolved Questions

Possible improvements to this framework

* Distributed multi-machine execution
* Move more components from testing program to the library, enabling more code reusing.
* 

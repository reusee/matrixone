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

Configurations can be saved to files, and committed to source repos. 
They can be run repeatly, be updated as project evolves. 

### Setup and execution

The setup module setup and boot the target system according to the configuration.  

### Model validation

TODO

### Report generation

TODO

## Testing Programs

TODO

# Drawbacks

TODO

# Rationale / Alternatives

TODO

# Unresolved Questions

TODO

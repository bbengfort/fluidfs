# Changelog

The release versions that are  tagged in Github. You can see the tags through the Github web application and download the binary of the version you'd like. Note that for many operating systems you'll have to compile for your local architecture (linux and darwin supported).

The versioning uses a three part version system, "a.b.c" - "a" represents a major release that may not be backwards compatible. "b" is incremented on minor releases that may contain extra features, but are backwards compatible. "c" releases are bug fixes or other micro changes that developers should feel free to immediately update to.

**Note**: FluidFS names major releases according to either major water monuments or falls around the world!

## Version 0.2 "Skogafoss"

* **tag**: [v0.2](https://github.com/bbengfort/fluidfs/releases/tag/v0.2)
* **name**: Skogafoss
* **release**: Pending
* **commit**: [see tag](#)

This release will include linearizeable and best-effort (and possibly transaction) consistency using single Raft. This is the last release before we move toward implementing Hierarchical Consistency directly.

Version 0.2 "Skogafoss" is named after the [Skógafoss Waterfall](https://en.wikipedia.org/wiki/Sk%C3%B3gafoss) in Iceland.

## Version 0.1 "Gullfoss"

* **tag**: [v0.1](https://github.com/bbengfort/fluidfs/releases/tag/v0.1)
* **name**: Gullfoss
* **release**: February 8, 2017
* **commit**: [d698e09](https://github.com/bbengfort/fluidfs/commit/d698e09b4ae658f6017be3e46e0c5a7a7d2ee53a)

This release, codenamed Gullfoss implements a simple in-memory file system without replication as well as many of the library mechanisms required for the full system. The system is configured with a YAML configuration file and data files located in the home directory (per-user processes) along with an fstab-like configuration for mounting directories. The system implements the FUSE filesystem API and can be mounted and operated against in an in-memory fashion. Database interactions and interfaces are defined using the db module, and both BoltDB and LevelDB can be used as database drivers. Additionally, the Gullfoss version has fixed and variable length chunking for data storage and replication. Two interfaces - both a command line and web interface have been implemented for command and control.

Version 0.1 "Gullfoss" is named after the [Gullfoss Waterfall](https://en.wikipedia.org/wiki/Gullfoss) in Iceland.  

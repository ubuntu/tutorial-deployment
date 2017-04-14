# tutorial-deployment
Serve and help generating tutorial deployment for https://tutorials.ubuntu.com

Those couple of tools are used in conjonction with https://github.com/canonical-websites/tutorials.ubuntu.com to generate a tutorial website. Those can be written in markdown or google doc, [using the claat google's library](https://github.com/googlecodelabs/tools).

## Important note
We are using currently [a fork](https://github.com/didrocks/codelab-ubuntu-tools) of the claat tools as our fixes for the markdown parser are getting reviewed and merged by the google team.

As this is a more robust deployment procedure, some structural changed were needed and are in progress [there](https://github.com/didrocks/tutorials.ubuntu.com/tree/reformat-tooling). This tool works exclusively with that branch.

## Generate
The `generate` command will generate tutorials in **html**, using [Polymerjs](https://www.polymer-project.org/), to be compatible with the aforementioned tutorial source code.

It fetches in well known places the codelab list and sources (both in **google doc** or **markdown** format), the general events and categories metadata, to generate the desired ouput and API files.

Every default directories will be detected by the tool if present in the tutorial directories. Arguments and options can tweak this behavior.

## Serve
The `serve` command will generate the same codelab content generated on the fly in a temporary directory, but also install watchers on local source files (codelab markdown file or any referenced local images).

Any save on any of those files will retrigger the corresponding codelab build and API generation, serve by this local http webserver (default port is **8080**)

Changes are all done in temporary files and not destructive on the tutorial repository. Note that source and webserver paths can be overriden as for the generate command.

#!/usr/bin/env bash

# build Antlr4 parser
# given antlr4 aliased as: alias antlr4='java -jar $HOME/lib/antlr-4.7-complete.jar'
java -jar lib/antlr-4.7.1-complete.jar -Werror -o tmp/ Able.g4

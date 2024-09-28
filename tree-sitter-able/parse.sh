#!/usr/bin/env bash

npx tree-sitter generate src/grammar.js

npx tree-sitter parse $@

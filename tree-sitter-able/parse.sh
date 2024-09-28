#!/usr/bin/env bash

tree-sitter generate grammar.js

tree-sitter parse $@

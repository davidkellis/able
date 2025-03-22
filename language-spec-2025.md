# Language Specification 2025

## Identifiers

All identifiers must conform to the pattern `[a-zA-Z0-9][a-zA-Z0-9_]*`

### Naming Conventions

- Prefer snake_case for file names, package names, variable names, and function names.
- Prefer PascalCase for type names.
- Primitive built-in types are all lowercase and single word names (e.g. string, u8, i128)

## Comments

```
## comments follow the double pound sign, `##`, to the end of the line
```

## Packages

A single file, `package.yml`, defines a project's build configuration, dependencies, package metadata, etc.

The directory containing the `package.yml` file is the root of the package.

The `package.yml` file has the following structure:

```
name: my-first-able-package
version: 1.0.0
license: MIT

authors:
- David <david@conquerthelawn.com>

dependencies:
  mysql:
    github: davidkellis/able-mysql
    version: ~>0.16.0
```

## Modules

Every source file should start with a module definition of the form `module <unqualified name>`, for example `module io`; otherwise, the file is interpreted as if `module main` were specified.

### Module Path

Every source file resides within a directory, and the directory structure relative to the package root directory is considered part of a module's path.

For example, if the package root directory contains a subdirectory called foo and there is a source file named bar.able within the foo subdirectory, then the module that is declared within the bar.able source file is considered to be a submodule of the foo module, since the foo directory represents the foo module.

Module path segments are delimited with a period, as in `foo.bar.baz`.

### Importing Modules

- Module import: `import io`
- Wildcard import: `import io.*`
- Selective import: `import io.{puts, gets, SomeType}`
- Aliased import: `import internationalization as i18n.{transform as tf}`

Imports may occur in module scope or any local scope.

## Variables

```
<variable name>: <type name>                        # assigned the zero value for <type name>
<variable name>: <type name> = <value expression>   # assigned the specified initial value
<variable name> = <value expression>                # assigned the specified initial value and the type is inferred from the type of <value expression> and the use of <variable name>
```

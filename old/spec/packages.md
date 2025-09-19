# Packages

Packages form a tree of namespaces rooted at the name of the library, and the hierarchy follows the directory structure of the source files within the library.

Every package has a root package name, which is defined to be the name of the package as documented in the `package.yml` file.

All unqualified package names must be valid identifiers.

Qualified package names are package paths composed of path segments delimited with a period, as in `foo.bar.baz`. Each path segment is an unqualified package name.

If a source file declares its package with a package definition of the form `package <unqualified-name>`, then the source code within the file is deemed to be part of the fully qualified package name `<root-package-name>.<intermediate-directories...>.<unqualified-name>`.

If a source file does not declare a package definition, then the source code within the file is deemed to be part of the fully qualified package name `<root-package-name>.<intermediate-directories...>`

Every source file resides within a directory, and the directory structure relative to the package root directory is considered part of a package's path.

Directory names containing source files must be valid identifiers with the exception that hypens will be treated as underscores when the directory name is used as a package name. For example, if `hello-world/foo.able` declares a package of `package my_package`, and the `hello-world` directory is the root of the package, then the fully qualified package name would be `hello_world.my_package`.

For illustration, assume a package is organized into a directory at the path `/home/david/projects/hello-world` with the following directory structure:

```
package.yml
foo.able
bar.able
baz.able
qux/quux.able
qux/corge.able
```

such that:

- package.yml specifies a name of `name: hello_world`
- foo.able does not specify a package declaration
- bar.able has a package declaration of `package bar`
- baz.able has a package declaration of `package qux`
- qux/quux.able does not specify a package declaration
- qux/corge.able has a package declaration of `package corge`

This sample project would define the following package tree:

- `hello_world` is the root package
- `hello_world` contains all of the definitions contained within `foo.able`
- `hello_world.bar` contains all of the definitions contained within `bar.able`
- `hello_world.qux` contains all of the definitions contained within `baz.able` as well as `qux/quux.able`
- `hello_world.qux.corge` contains all of the definitions contained within `qux/corge.able`

## Package Config

A single file, `package.yml`, defines a project's build configuration, dependencies, package metadata, etc.

The directory containing the `package.yml` file is the root of the library.

The `package.yml` file has the following structure:

```
name: hello_world
version: 1.0.0
license: MIT

authors:
- David <david@conquerthelawn.com>

dependencies:
  collections:
    github: davidkellis/able-collections
    version: ~>0.16.0
```

## Importing Packages

- Package import: `import io`
- Wildcard import: `import io.*`
- Selective import: `import io.{puts, gets, SomeType}`
- Aliased import: `import internationalization as i18n.{transform as tf}`

Imports may occur in package scope or any local scope.

Importing identifiers from a package creates a new identifier binding within the binding scope of the import expression such that the locally bound names refer to the same object in memory that the imported package identifier references.

For example, if a function named `baz` is defined in the a package named `foo.bar`, and we import it with `import foo.bar.baz`, then the locally scoped `baz` identifier may be used to invoke the `foo.bar.baz` function. Additionally, the locally scoped `baz` identifier and `foo.bar.baz` are both referentially equal - they both reference the same object in memory.

## Package exports

All identifiers defined in the package scope are public and exported by default. If you wish for an identifier to be private to the package you may prefix the identifier with the `private` keyword.

For example:

```
foo = "bar"
private baz = "qux"
```

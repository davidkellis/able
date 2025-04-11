This parser is modeled after https://tomassetti.me/incremental-parsing-using-tree-sitter/ and https://github.com/gabriele-tomassetti/tree-sitter-story.

There is currently a bug in the latest version of tree-sitter, and without referencing the files in https://github.com/gabriele-tomassetti/tree-sitter-story I don't think I would have ended up with a functioning parser.

I had to copy most of the package.json and package-lock.json from https://github.com/gabriele-tomassetti/tree-sitter-story, blow away node_modules and build, and then `npm i`.

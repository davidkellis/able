// Below: a fairly typical program, in this case used to display links from HTML.

use std::{
    fmt::Display,
    io::{self, Read},
    process,
};

use clap::Parser;
use lnx::{LinkExtractor, Result};

/// A program for extracting links from HTML.
#[derive(Debug, Parser)]
struct Args {
    /// the base url used for relative links
    #[arg(short, long)]
    url: Option<String>,
    /// a style expression used to target specific links
    ///
    /// Defaults to just "a"
    #[arg(short, long)]
    style: Option<String>,
}

fn main() {
    if let Err(e) = run(Args::parse()) {
        eprintln!("{e}");
        process::exit(1);
    }
}

fn run(args: Args) -> Result<()> {
    let text = read_input()?;
    let extractor = args.style.as_deref().map_or_else(
        || LinkExtractor::new(&text),
        |style| LinkExtractor::with_style(&text, style),
    );

    match args.url.as_deref() {
        Some(url) => display(extractor.links_with_url(url)),
        None => display(extractor.links()),
    }

    Ok(())
}

fn display(links: impl IntoIterator<Item: Display>) {
    for link in links {
        println!("{link}");
    }
}

fn read_input() -> io::Result<String> {
    let mut text = String::new();
    io::stdin().lock().read_to_string(&mut text)?;
    Ok(text)
}

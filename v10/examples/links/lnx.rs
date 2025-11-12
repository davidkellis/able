// Here, the LinkExtractor type from the previous file:

pub struct LinkExtractor {
    document: Html,
    selector: Selector,
}

impl LinkExtractor {
    pub fn new(text: &str) -> Self {
        Self {
            document: Html::parse_fragment(text),
            selector: Selector::parse("a").unwrap(),
        }
    }

    /// `style` overrides the CSS selector used to identify links
    pub fn with_style(text: &str, style: &str) -> Self {
        Self {
            selector: Selector::parse(style).unwrap_or_else(|_| Selector::parse("a").unwrap()),
            ..Self::new(text)
        }
    }

    pub fn links(& self) -> impl Iterator<Item = Link<'_>> {
        let mut duplicate_filter = HashSet::new();
        self.document
            .select(&self.selector)
            .filter_map(|link| link.attr("href"))
            .filter(move |&href| duplicate_filter.insert(href))
            .map(|link| Link { link, url: "" })
    }

    pub fn links_with_url<'a>(&'a self, url: &'a str) -> impl Iterator<Item = Link<'a>> + 'a {
        self.links().map(|link| link.with_url(url))
    }
}

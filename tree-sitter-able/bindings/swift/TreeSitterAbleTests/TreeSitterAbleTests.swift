import XCTest
import SwiftTreeSitter
import TreeSitterAble

final class TreeSitterAbleTests: XCTestCase {
    func testCanLoadGrammar() throws {
        let parser = Parser()
        let language = Language(language: tree_sitter_able())
        XCTAssertNoThrow(try parser.setLanguage(language),
                         "Error loading Able grammar")
    }
}

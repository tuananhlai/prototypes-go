# Basic Calculator Implementation

This project implements a basic arithmetic calculator in Go, capable of evaluating expressions with additions, subtractions, and parentheses.

## Computer Science Concepts

This solution utilizes several fundamental computer science concepts found in compiler design and language processing.

### 1. Lexical Analysis (Tokenization)

**Definition**: Lexical analysis is the first phase of a compiler or interpreter. It converts a sequence of characters (the source code) into a sequence of meaningful units called **tokens**.

**Detailed Explanation**:
The source code is naturally a stream of characters (e.g., `(1 + 2)`). To a computer, this is just bytes. The **Lexer** (or Tokenizer) scans this stream and groups characters into **lexemes** that map to specific token types.

- **Skipping Whitespace**: Spaces are usually ignored as they don't affect code logic.
- **Categorization**: '1' is recognized as a specific type `NUMBER`. '+' is recognized as `PLUS`.
- **State**: The lexer acts as a simple state machine. For example, when it sees a digit, it enters a "reading number" state and consumes characters until it sees a non-digit.

**In this Project**:
The `tokenize` function iterates through the input string.

- If it sees a digit, it keeps reading until the number ends, creating a `tokenTypeNumber`.
- If it sees `(`, `)`, `+`, or `-`, it immediately creates the corresponding token.
- Result: `"1 + 2"` becomes `[ {Type: NUMBER, Val: "1"}, {Type: PLUS}, {Type: NUMBER, Val: "2"} ]`.

### 2. Recursive Descent Parsing

**Definition**: A top-down parsing technique where the structure of the code is constructed by mutually recursive procedures (functions).

**Detailed Explanation**:
In a recursive descent parser, every non-terminal symbol in the grammar corresponds to a function in the code.

- **"Top-Down"**: It starts from the highest-level rule (e.g., "Parse Program") and works its way down to the details (e.g., "Parse Number").
- **"Recursive"**: Functions call themselves to handle nested structures. For example, an expression can contain another expression inside parentheses. The parser handles this by calling `parseExpression` from within `parseExpression`.
- **Call Stack**: The system's call stack implicitly manages the nesting levels. When `(1 + (2 + 3))` is parsed, the parser pauses the outer expression to process the inner `(2 + 3)`, effectively pushing the state onto the stack.

**In this Project**:

- `readExpression` is the core function.
- When it encounters `(`, it calls `readParenExpr`.
- `readParenExpr` might find another `(`, calling `readParenExpr` again (or `readExpression`).
- This recursion allows for arbitrarily deep nesting of parentheses.

### 3. Abstract Syntax Tree (AST)

**Definition**: An AST is a tree representation of the abstract syntactic structure of source code.

**Detailed Explanation**:

- **Tree Structure**: Source code is linear, but logic is hierarchical. An AST captures this hierarchy.
- **"Abstract"**: It leaves out details from the real syntax that are implicit in the tree structure. For instance, grouping parentheses `((1+2))` don't typically exist as nodes in the AST; instead, the structure itself (the `1+2` node being a child of another node) implies the grouping.
- **Nodes**:
  - **Leaf Nodes**: Comparison operands or literals (e.g., logical numbers `1`, `2`).
  - **Internal Nodes**: Operators that consist of other nodes (e.g., `Add` contains `left` and `right` children).

**In this Project**:

- **Interface**: `expression` is the node type.
- **Leaf**: `numberExpr` (holds an `int`).
- **Internal**: `binaryExpr` (holds `left` and `right` `expression` nodes and an operator).
- **Evaluation**: The tree is evaluated via a **Post-Order Traversal** (visit children, then visit root). calling `.value()` on the root allows the call to ripple down to the leaves and bubble up the result.

### 4. Lookahead Parsing (Predictive Parsing)

**Definition**: A parsing validation strategy where the parser looks at the next (or upcoming) tokens to decide which grammar rule to apply.

**Detailed Explanation**:
Parsers often face ambiguity. If the current symbol is `A`, can it be followed by `B` or `C`?

- **LL(1)**: A common category of parsers (Left-to-right, Leftmost derivation) that needs **1** token of lookahead.
- By checking the `current_token`, the parser "predicts" the structure without needing to backtrack (try a path, fail, and go back).

**In this Project**:
In `readExpression`, the parser checks `p.tokens[p.cur].typ`:

- `case tokenTypeNumber`: It knows strictly to call `readNumber`.
- `case tokenTypeLParen`: It knows strictly to call `readParenExpr`.
- `case tokenTypeMinus`: It knows strictly to call `readNegativeExpr`.
  There is no guessing. The single lookahead token is sufficient to determine the entire path.

### 5. Backus–Naur Form (BNF)

**Definition**: BNF is a formal notation (metasyntax) used to describe the context-free grammar of a language.

**Detailed Explanation**:
It defines the valid syntax rules using:

- **Terminals**: Literal characters/strings (e.g., `"+"`, `"("`, digits).
- **Non-Terminals**: placeholders for patterns (e.g., `<expression>`, `<number>`).
- **Productions**: Rules defined with `::=`, meaning "can be replaced by".
- **Alternatives**: Choices separated by `|`.

**Grammar for this Calculator**:

```bnf
<parse>            ::= <expression>
<expression>       ::= <number> | <paren_expr> | <unary_minus_expr>
<paren_expr>       ::= "(" <expression_sequence> ")"
<expression_sequence> ::= <expression> { <operator> <expression> }
<unary_minus_expr> ::= "-" <expression>
<operator>         ::= "+" | "-"
<number>           ::= <digit> { <digit> }
<digit>            ::= "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9"
```

**Breakdown**:

1.  **`<expression>`**: Can be a simple number, a parenthesized group, or a negative number.
2.  **`<paren_expr>`**: Must start with `(`, contain a sequence of expressions, and end with `)`.
3.  **`<expression_sequence>`**: Handles chains like `1 + 2 - 3`. It is an expression followed by zero or more pairs of `operator` and `expression`.
4.  **`{ }`**: Represents repetition (0 or more times).

This grammar formally proves that `((1+2))` is valid, but `(1+)` is not (because `<expression>` cannot be empty after `+`).

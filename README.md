# parsing in golang

Simple lexing and parsing of lightweight languages, based on the assumption of whitespace
separation of non-symbol tokens with C++-style single- (//) and multi-line (/*...*/) comments.

The lexer can be fine-tuned/extended by adding 'intercepts', the parser can be extended by
adding rules to combine tokens.

Includes helper functions for goroutine-safe application wide stats counting and timing.

TODO:

- Extend Documentation
- Round-out Tests
- Add Examples
- Make stats/counters non-singleton so that user can make that decision
- Parser generators

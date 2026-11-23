# emptycomment

Forbids `/* */` block comments that contain only whitespace and `*` characters. Empty block comments are leftovers from deleted text or unfinished stubs and carry no information. `//` comments are never reported, because an empty `//` line separates the paragraphs of a doc comment.

```go
// Bad
/**/
/* */
/*
 *
 */

// Good
/* explanation */
//
```

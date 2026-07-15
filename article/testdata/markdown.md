# Structured Text Rendering

Markdown pages render through the same block parser as HTML, so *emphasis*,
**strong text**, `inline code`, ~~strikethrough~~ and [absolute
links](https://example.com/docs) all carry through. Relative links resolve
against the page URL: [a nearby post](../nearby-post/) becomes absolute, and
bare autolinks like https://example.com/feed.xml work too.[^1]

## Quotes and code

> And they said one to another, Go to, let us make brick, and burn them
> thoroughly. And they had brick for stone, and slime had they for morter.

```go
func main() {
	fmt.Println("hello")
}
```

## Lists and tables

- first item
- second item
  - nested item

1. ordered
2. items

| Name    | Value |
| ------- | ----- |
| rows    | 2     |
| columns | also 2 |

---

![architecture diagram](images/diagram.png)

[^1]: Footnotes fold into a section at the end of the document.

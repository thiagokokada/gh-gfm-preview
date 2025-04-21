This is a test of all Markdown possibilities:

------------------------------------------

## Anchors

- [Headings](#headings)
- [Horizontal Rules](#horizontal-rules)
- [Emphasis](#emphasis)
- [Links](#links)
- [Blockquotes](#blockquotes)
- [Indentation](#indentation)
- [Lists](#lists)
  + [Unordered](#unordered)
  + [Ordered](#ordered)
    * [Numbers in sequence](#numbers-in-sequence)
    * [Numbers not in sequence](#numbers-not-in-sequence)
- [Images](#images)
- [Tables](#tables)
- [Code](#code)
- [Math](#math)
- [Raw HTML](#raw-html)

------------------------------------------

## Headings

# h1 Heading 1
## h2 Heading 2
### h3 Heading 3
#### h4 Heading 4
##### h5 Heading 5
###### h6 Heading 6

------------------------------------------

## Horizontal Rules

___

---

***

------------------------------------------

## Emphasis

**This is bold text**

__This is bold text__

*This is italic text*

_This is italic text_

~~Strikethrough~~

------------------------------------------

## Links

[link text][1]

[link with title][2]

This is [an example](http://example.com/ "Title") inline link.

[This link](http://example.net/) has no title attribute.

------------------------------------------

## Blockquotes

> Blockquotes can also be nested...
>> ...by using additional greater-than signs right next to each other...
> > > ...or with spaces between arrows.

------------------------------------------

## Indentation

  indentation 1-1

indentation 1-2
    indentation 2-1

------------------------------------------

## Lists

### Unordered

+ Create a list by starting a line with `+`, `-`, or `*`
+ Sub-lists are made by indenting 2 spaces:
  - Marker character change forces new list start:
    * Ac tristique libero volutpat at
    + Facilisis in pretium nisl aliquet
    - Nulla volutpat aliquam velit
+ Very easy!

### Ordered

#### Numbers in sequence

1. Lorem ipsum dolor sit amet
2. Consectetur adipiscing elit
3. Integer molestie lorem at massa

#### Numbers not in sequence

1. You can use sequential numbers...
1. ...or keep all the numbers as `1.`

------------------------------------------

## Images

![Minion][3]
![Stormtroopocat][4]

Like links, Images also have a footnote style syntax:

![Alt text][5]

------------------------------------------

## Tables

| Option | Description |
| ------ | ----------- |
| data   | path to data files to supply the data that will be passed into templates. |
| engine | engine to be used for processing templates. Handlebars is the default. |
| ext    | extension to be used for dest files. |

Right aligned columns

| Option | Description |
| ------:| -----------:|
| data   | path to data files to supply the data that will be passed into templates. |
| engine | engine to be used for processing templates. Handlebars is the default. |
| ext    | extension to be used for dest files. |

------------------------------------------

## Code

Inline `code`

Indented code

    // Some comments
    line 1 of code
    line 2 of code
    line 3 of code


Block code "fences"

```
Sample text here...
```

Syntax highlighting

``` js
var foo = function (bar) {
  return bar++;
};

console.log(foo(5));
```

------------------------------------------

## Math

When $a \ne 0$, there are two solutions to $(ax^2 + bx + c = 0)$ and they are

$$ x = {-b \pm \sqrt{b^2-4ac} \over 2a} $$

------------------------------------------

## Raw HTML

<!-- https://github.com/svengreb/styleguide-markdown/blob/main/rules/raw-html.md -->
<p align="center">The winter is winter is sparkling and frozen!</p>

Sparkling <img src="https://raw.githubusercontent.com/nordtheme/assets/main/static/images/artworks/arctic/nature/dark/snowfall.svg?sanitize=true" width=16 height=16 align="center" /> snowflakes falling down in the winter!

<details>
  <summary>Winter</summary>
  <p>Sparkling and frozen!</p>
</details>

<!-- https://gist.github.com/seanh/13a93686bf4c2cb16e658b3cf96807f2#file-html_tags_you_can_use_on_github-md -->
User input with `<kbd>`: <kbd>help mycommand</kbd>.

Subscripts<sub>sub</sub> and superscripts<sup>sup</sup> with `<sub>` and `<sup>`.

You can use `<sup>` to create linked footnotes.<sup id="backToMyFootnote"><a href="#myFootnote">1</a></sup>

------------------------------------------

[1]: http://dev.nodeca.com
[2]: http://nodeca.github.io/pica/demo/ "title text!"
[3]: images/dinotocat.png
[4]: https://octodex.github.com/images/saritocat.png "The Stormtroopocat"
[5]: https://octodex.github.com/images/daftpunktocat-thomas.gif "The Dojocat"

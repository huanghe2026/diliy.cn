# Welcome to the Geoyuan blog

> `nyx@geoyuan:~# ./post.sh --title "hello world"`

So we finally stood up a place to write things down.

For a while the team lived entirely inside terminals, group chats, and the
occasional whiteboard covered in hex. That works when you're four people in a
room, but the crew is growing and the memory is leaking. This blog is the fix:
a quiet corner of `diliy.cn` where we keep what's worth keeping.

## What goes here

- **CTF writeups** — the ones we're allowed to talk about.
- **Tooling notes** — things we built, broke, and rebuilt at 3 AM.
- **Team updates** — where we are, where we're headed.

## How it's built

This whole thing is a [mdBook](https://github.com/rust-lang/mdBook) project.
Markdown in, a static site out, served straight from `diliy.cn/blogs/`. No
database, no tracking, no framework churn — just text and a dark theme that
matches the rest of the site.

To add a post:

```markdown
# My new post
Write things here.
```

Then link it from `src/SUMMARY.md` and run `mdbook build`. That's the entire
pipeline.

---

*— nyx, lead / reverse eng*

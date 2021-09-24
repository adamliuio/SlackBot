# SlackBot

A bot that helps you utilize the power of Slack.

Hacker News ([API](https://github.com/HackerNews/API))

## Commands

`/hn top 10-20`: Retrieve the top 10 to 20 stories.

- `top` can be one of top/new/best
- `10-20` means "10 to 20", it can also be a single number like "5"

## Shortcuts

There are 2 shortcuts: **Global** or **On Message**.
`ResponseUrl` `Channel` fields in `SCPayload` are empty if the shortcut is global
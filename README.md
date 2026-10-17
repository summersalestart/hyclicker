<div align="center">
  <img src="banner.png" alt="HyClicker">
</div>

[![Build](https://img.shields.io/github/actions/workflow/status/summersalestart/hyclicker/ci.yml?branch=main)](https://github.com/summersalestart/hyclicker/actions)
[![Latest Release](https://img.shields.io/github/v/release/summersalestart/hyclicker)](https://github.com/summersalestart/hyclicker/releases/latest)
[![Last Commit](https://img.shields.io/github/last-commit/summersalestart/hyclicker)](https://github.com/summersalestart/hyclicker/commits/main)
[![Downloads](https://img.shields.io/github/downloads/summersalestart/hyclicker/total)](https://github.com/summersalestart/hyclicker/releases)
[![Stars](https://img.shields.io/github/stars/summersalestart/hyclicker)](https://github.com/summersalestart/hyclicker/stargazers)

# HyClicker

Simple dark auto clicker for Windows. You give it a min/max CPS range and it clicks somewhere in between, drifting around so it doesn't look like a metronome. There's a global hotkey to toggle it even when the window isn't focused.

This is the Electron rewrite of an older tkinter version (which was deleted — it looked hideous thanks to tkinter's aliasing, and nobody used it).

## How the clicking works

It doesn't click at one fixed rate. Every 0.5–2 seconds it picks a new target CPS somewhere inside your min/max range, then fires clicks at that rate with a bit of jitter (±15%) thrown on top. Delays are clamped between 30ms and 500ms so nothing goes insane.

If you set min 8 / max 12, it'll wander around between 8 and 12 instead of sitting at one speed.

## Usage

- `CPS (1)` / `CPS (2)` — the range it drifts between (defaults: 8 and 12)
- `Start` — obvious. There's a short delay after starting so your cursor isn't the thing being clicked.
- `Keybind` — click it, press whatever key you want the toggle to be. F6 by default. Works globally.
- `Left` — click to swap to Right mouse button
- yellow dot minimizes, red dot quits, github cat is github

## Running from source

```bash
cd hyclicker-app
npm install
npm start
```

## Building

```bash
cd hyclicker-app
npm run dist
```

The exe ends up in `hyclicker-app/dist/`.

## Releases / updates

The exe checks the GitHub releases page on launch. If there's a newer version it'll show a little bar with a link to the releases page — it just opens the browser, it doesn't touch anything on disk. Download the new exe yourself like a normal person.

## The usual disclaimers

- This is unsigned, so SmartScreen might yell. The source is right here if you'd rather build it yourself.
- A lot of games ban auto clickers. Read the ToS of whatever you're clicking on, that's on you, not me.

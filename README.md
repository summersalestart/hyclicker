<div align="center">
  <img src="banner-dark.png" alt="HyClicker">
</div>

# HyClicker

Simple dark auto clicker for Windows. You give it a min/max CPS range and it clicks somewhere in between, drifting around so it doesn't look like a metronome. There's a global hotkey to toggle it even when the window isn't focused.

This started as a tkinter thing (`main.py`, still works if that's your thing) and got rebuilt as an Electron app because tkinter's canvas has no anti-aliasing and the rounded corners looked like garbage. The Electron version is the one I actually use.

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

Python version (the original):

```bash
pip install -r requirements.txt
python main.py
```

Electron version (the current one, under `hyclicker-app/`):

```bash
cd hyclicker-app
npm install
npm start
```

## Building

Python exe:

```bash
pip install pyinstaller
pyinstaller HyClicker.spec
```

Electron exe:

```bash
cd hyclicker-app
npm run dist
```

Both end up in their respective `dist/` folders.

## Releases / updates

The exe checks the GitHub releases page on launch. If there's a newer version it'll show a little bar with a link to the releases page — it just opens the browser, it doesn't touch anything on disk. Download the new exe yourself like a normal person.

## The usual disclaimers

- This is unsigned, so SmartScreen might yell. The source is right here if you'd rather build it yourself.
- A lot of games ban auto clickers. Read the ToS of whatever you're clicking on, that's on you, not me.

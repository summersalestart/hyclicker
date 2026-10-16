# HyClicker — Auto Clicker

A compact, modern dark-themed auto-clicker utility with a naturalistic,
varying click speed and a rebindable global hotkey.

## Files

| File                 | Purpose                                                        |
|----------------------|-----------------------------------------------------------------|
| `main.py`            | Entry point — creates the window and starts the app.           |
| `gui.py`             | Custom Tkinter UI matching the minimalist dark design.         |
| `engine.py`          | The click loop: naturalistic CPS timing, thread management.    |
| `keybind_manager.py` | Global hotkey registration + "press a key to rebind" capture.  |
| `theme.py`           | Colors, fonts, and geometric rounded curves.                    |
| `config.py`          | Persistent JSON configuration loader and saver.                |
| `requirements.txt`   | Third-party dependencies.                                       |

The click-timing logic (`engine.py`) and hotkey logic (`keybind_manager.py`)
have zero Tkinter dependencies — they can be reused with different interfaces
or tested independently.

## 1. Installation

Requires Python 3.8+.

```bash
python -m venv venv
venv\Scripts\activate        # Windows
# source venv/bin/activate   # macOS/Linux

pip install -r requirements.txt
```

`tkinter` ships with the standard Python installer on Windows/macOS.
On Linux you may need: `sudo apt-get install python3-tk`.

No administrator/root privileges are required to run it.

## 2. Running

```bash
python main.py
```

- **CPS (1)** = minimum clicking speed (default 8).
- **CPS (2)** = maximum clicking speed (default 12).
- **Start / Stop** — toggles naturalistic auto-clicking.
- **Keybind** — click it, then press any key to rebind the global toggle (default **F6**).
- **GitHub Icon** — opens https://github.com/summersalestart/hyclicker in your browser.
- **Left / Right** — toggles between Left and Right mouse button clicking.
- **Yellow Dot** — minimizes the window.
- **Red Dot** — closes the app cleanly.

The keybind works globally even when another game or window is active.

## 3. Building a Standalone .exe

```bash
pip install pyinstaller
pyinstaller HyClicker.spec
```

The resulting executable will be in `dist/HyClicker.exe`.

## 3b. HTML/CSS/JS Desktop Rebuild (Electron)

`hyclicker-app/` is a pixel-faithful rebuild of the same UI with HTML + CSS + JS.
All rounding uses CSS `border-radius` (GPU anti-aliased), and button hovers are
motion-only (scale/lift/shadow, no color change).

```bash
cd hyclicker-app
npm install
npm start        # run the app
npm run dist     # portable .exe via electron-builder
```

## 4. Architecture & Timing System

- **Phase-based CPS Variation**: Rather than uniform random delay, the engine shifts between subtle speed phases with human-like jitter (±15%), mimicking realistic clicking.
- **Thread Safety**: The click engine runs on a daemon thread, leaving the UI perfectly responsive.
- **Clean Teardown**: Listeners and threads terminate gracefully upon exit.

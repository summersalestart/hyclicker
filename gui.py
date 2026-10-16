import base64
import os
import sys
import tkinter as tk
import webbrowser

import config
from engine import BUTTON_MAP, ClickEngine, MAX_ALLOWED_CPS
from keybind_manager import KeybindManager
import theme
import winstyle

ICON_FILENAME = "icon.ico"
WINDOW_W = 342
WINDOW_H = 382
CORNER_RADIUS = 26

# Base64-encoded GitHub circular badge (38x38 PNG) matching the exact UI asset
GITHUB_BADGE_B64 = (
    "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"
)


def _resolve_icon_path():
    base_dir = getattr(sys, "_MEIPASS", os.path.dirname(os.path.abspath(__file__)))
    path = os.path.join(base_dir, ICON_FILENAME)
    return path if os.path.isfile(path) else None


class HyClickerApp:

    def __init__(self, root: tk.Tk):
        self.root = root
        self._drag_offset = (0, 0)
        self._is_capturing = False
        self._error_timer = None

        self.engine = ClickEngine(on_state_change=self._on_engine_state_change)

        saved = config.load_config()
        self.keybinds = KeybindManager(
            on_trigger=self._on_hotkey_triggered,
            on_capture_complete=self._on_keybind_captured,
            initial_hotkey_str=saved.get("hotkey_str"),
            initial_display_str=saved.get("display_str"),
        )

        self.mouse_button_name = saved.get("mouse_button", "Left")
        if self.mouse_button_name not in BUTTON_MAP:
            self.mouse_button_name = "Left"

        self.font_title = theme.pick_font(self.root, theme.FONT_TITLE_CANDIDATES)
        self.font_label = theme.pick_font(self.root, theme.FONT_LABEL_CANDIDATES)
        self.font_input = theme.pick_font(self.root, theme.FONT_INPUT_CANDIDATES)
        self.font_button = theme.pick_font(self.root, theme.FONT_BUTTON_CANDIDATES)
        self.font_small = theme.pick_font(self.root, theme.FONT_SMALL_CANDIDATES)
        self.font_toast = theme.pick_font(self.root, theme.FONT_TOAST_CANDIDATES)

        self._configure_root()
        self._build_ui()
        self._restore_saved_inputs(saved)

    def _configure_root(self):
        self.root.title("HyClicker")
        self.root.geometry(f"{WINDOW_W}x{WINDOW_H}")
        self.root.resizable(False, False)
        self.root.overrideredirect(True)

        # Transparency for smooth rounded corners on Windows
        self._transparent_supported = sys.platform == "win32"
        if self._transparent_supported:
            try:
                self.root.configure(bg=theme.TRANSPARENT_KEY)
                self.root.wm_attributes("-transparentcolor", theme.TRANSPARENT_KEY)
            except Exception:
                self._transparent_supported = False
                self.root.configure(bg=theme.BG)
        else:
            self.root.configure(bg=theme.BG)

        icon_path = _resolve_icon_path()
        if icon_path:
            try:
                self.root.iconbitmap(icon_path)
            except Exception:
                pass

        self.root.bind("<Map>", self._on_map)
        self.root.protocol("WM_DELETE_WINDOW", self._on_close)

    def _on_map(self, event):
        if event.widget == self.root and self.root.state() == "normal" and not self.root.overrideredirect():
            self.root.after(10, lambda: self.root.overrideredirect(True))

    def _build_ui(self):
        # Master Canvas covering entire window with rounded corners
        bg_color = theme.TRANSPARENT_KEY if self._transparent_supported else theme.BG
        self.canvas = tk.Canvas(
            self.root,
            width=WINDOW_W,
            height=WINDOW_H,
            bg=bg_color,
            highlightthickness=0,
            bd=0,
        )
        self.canvas.pack(fill="both", expand=True)

        # Outer rounded rectangle background
        self.bg_polygon = theme.draw_rounded_rect(
            self.canvas,
            0,
            0,
            WINDOW_W,
            WINDOW_H,
            CORNER_RADIUS,
            fill=theme.BG,
            outline="",
        )

        # Allow dragging anywhere on background
        self._draggable_items = []
        self._bind_draggable(self.bg_polygon)

        # Invisible header drag handle (top strip). Placed above the
        # background but below title/dots so buttons stay clickable.
        self.header_drag = self.canvas.create_rectangle(
            0, 0, WINDOW_W, 68, fill=theme.BG, outline=""
        )
        self._bind_draggable(self.header_drag)

        # ----------------- HEADER -----------------
        # Title "HyClicker" (top left)
        self.title_text = self.canvas.create_text(
            16,
            33,
            text="HyClicker",
            fill=theme.TEXT_PRIMARY,
            font=self.font_title,
            anchor="w",
        )
        self.canvas.tag_bind(self.title_text, "<ButtonPress-1>", self._start_drag)
        self.canvas.tag_bind(self.title_text, "<B1-Motion>", self._do_drag)
        self._draggable_items.append(self.title_text)

        # Minimize dot (Yellow, center x=275.5, y=31.5, r=12.5)
        self._build_yellow_dot(263, 19, 288, 44)

        # Close dot (Red, center x=312, y=32, r=12.5)
        self._build_red_dot(300, 20, 325, 45)

        # Divider dotted line at y=68
        self.divider = self.canvas.create_line(
            16,
            68,
            WINDOW_W - 16,
            68,
            fill=theme.DIVIDER,
            dash=(3, 3),
        )
        self._bind_draggable(self.divider)

        # ----------------- CPS SECTION -----------------
        # CPS (1) Label (center x=104, y=102)
        self.cps1_label = self.canvas.create_text(
            104,
            102,
            text="CPS (1)",
            fill=theme.TEXT_PRIMARY,
            font=self.font_label,
            anchor="center",
        )
        self._bind_draggable(self.cps1_label)

        # CPS (2) Label (center x=238, y=102)
        self.cps2_label = self.canvas.create_text(
            238,
            102,
            text="CPS (2)",
            fill=theme.TEXT_PRIMARY,
            font=self.font_label,
            anchor="center",
        )
        self._bind_draggable(self.cps2_label)

        # CPS (1) Pill Input (center x=104, y=148, w=111, h=46, r=23)
        self.cps1_entry = self._build_pill_input(
            cx=104,
            cy=148,
            w=111,
            h=46,
            placeholder="Value",
            default_numeric="8",
        )

        # CPS (2) Pill Input (center x=238, y=148, w=111, h=46, r=23)
        self.cps2_entry = self._build_pill_input(
            cx=238,
            cy=148,
            w=111,
            h=46,
            placeholder="Value",
            default_numeric="12",
        )

        # ----------------- KEYBIND SECTION -----------------
        # "Keybind" Label (center x=171, y=193.5)
        self.keybind_label = self.canvas.create_text(
            171,
            193,
            text="Keybind",
            fill=theme.TEXT_PRIMARY,
            font=self.font_label,
            anchor="center",
        )
        self._bind_draggable(self.keybind_label)

        # Keybind Pill Button (center x=171, y=235, w=116, h=46, r=23)
        self._build_keybind_pill(cx=171, cy=235, w=116, h=46)

        # ----------------- START BUTTON -----------------
        # Start button (center x=171, y=325, w=112, h=46, r=23)
        self._build_start_button(cx=171, cy=325, w=112, h=46)

        # ----------------- BOTTOM BAR -----------------
        # GitHub Octocat badge (bottom-left, center x=31.5, y=353.5)
        self._build_github_button(cx=31.5, cy=353.5)

        # Mouse Button pill ("Left" / "Right", bottom-right, center x=294, y=356, w=64, h=26)
        self._build_mouse_button_pill(cx=294, cy=356, w=64, h=26)

        # Inline non-intrusive notification / error label (centered below keybind, auto-fading)
        self.toast_text = self.canvas.create_text(
            171,
            280,
            text="",
            fill=theme.DOT_RED,
            font=self.font_toast,
            anchor="center",
        )
        self._bind_draggable(self.toast_text)

    # -------------------------------------------------------------
    # WIDGET BUILDERS
    # -------------------------------------------------------------
    def _build_yellow_dot(self, x1, y1, x2, y2):
        dot = self.canvas.create_oval(
            x1, y1, x2, y2, fill=theme.DOT_YELLOW, outline=""
        )
        self.canvas.tag_bind(
            dot, "<Enter>", lambda e: self.canvas.itemconfig(dot, fill=theme.DOT_YELLOW_HOVER)
        )
        self.canvas.tag_bind(
            dot, "<Leave>", lambda e: self.canvas.itemconfig(dot, fill=theme.DOT_YELLOW)
        )
        self.canvas.tag_bind(
            dot, "<Button-1>", lambda e: self._on_minimize()
        )
        self.canvas.tag_bind(
            dot, "<Enter>", lambda e: self.canvas.configure(cursor="hand2"), add="+"
        )
        self.canvas.tag_bind(
            dot, "<Leave>", lambda e: self.canvas.configure(cursor=""), add="+"
        )

    def _build_red_dot(self, x1, y1, x2, y2):
        dot = self.canvas.create_oval(
            x1, y1, x2, y2, fill=theme.DOT_RED, outline=""
        )
        self.canvas.tag_bind(
            dot, "<Enter>", lambda e: self.canvas.itemconfig(dot, fill=theme.DOT_RED_HOVER)
        )
        self.canvas.tag_bind(
            dot, "<Leave>", lambda e: self.canvas.itemconfig(dot, fill=theme.DOT_RED)
        )
        self.canvas.tag_bind(
            dot, "<Button-1>", lambda e: self._on_close()
        )
        self.canvas.tag_bind(
            dot, "<Enter>", lambda e: self.canvas.configure(cursor="hand2"), add="+"
        )
        self.canvas.tag_bind(
            dot, "<Leave>", lambda e: self.canvas.configure(cursor=""), add="+"
        )

    def _build_pill_input(self, cx, cy, w, h, placeholder="Value", default_numeric="8"):
        x1, y1 = cx - w / 2, cy - h / 2
        x2, y2 = cx + w / 2, cy + h / 2
        r = h / 2

        poly = theme.draw_rounded_rect(
            self.canvas, x1, y1, x2, y2, r, fill=theme.PILL_BG, outline=""
        )

        var = tk.StringVar(value=placeholder)
        entry = tk.Entry(
            self.canvas,
            textvariable=var,
            font=self.font_input,
            justify="center",
            relief="flat",
            bd=0,
            bg=theme.PILL_BG,
            fg=theme.TEXT_MUTED,
            insertbackground=theme.TEXT_PRIMARY,
            highlightthickness=0,
        )
        win = self.canvas.create_window(cx, cy, window=entry, width=w - 24, height=h - 14)

        state = {
            "var": var,
            "entry": entry,
            "poly": poly,
            "win": win,
            "placeholder": placeholder,
            "default": default_numeric,
        }

        def on_focus_in(e):
            if var.get() == placeholder:
                var.set("")
                entry.configure(fg=theme.TEXT_INPUT)
            self.canvas.itemconfig(poly, fill=theme.PILL_FOCUS)
            entry.configure(bg=theme.PILL_FOCUS)

        def on_focus_out(e):
            val = var.get().strip()
            if not val:
                var.set(placeholder)
                entry.configure(fg=theme.TEXT_MUTED)
            else:
                entry.configure(fg=theme.TEXT_INPUT)
            self.canvas.itemconfig(poly, fill=theme.PILL_BG)
            entry.configure(bg=theme.PILL_BG)
            self._save_inputs_to_config()

        entry.bind("<FocusIn>", on_focus_in)
        entry.bind("<FocusOut>", on_focus_out)
        self.canvas.tag_bind(poly, "<Button-1>", lambda e: entry.focus_set())

        return state

    def _build_keybind_pill(self, cx, cy, w, h):
        x1, y1 = cx - w / 2, cy - h / 2
        x2, y2 = cx + w / 2, cy + h / 2
        r = h / 2

        self.keybind_poly = theme.draw_rounded_rect(
            self.canvas, x1, y1, x2, y2, r, fill=theme.PILL_BG, outline=""
        )
        self.keybind_text = self.canvas.create_text(
            cx,
            cy,
            text=self.keybinds.display_string,
            fill=theme.TEXT_MUTED,
            font=self.font_input,
            anchor="center",
        )

        for tag in (self.keybind_poly, self.keybind_text):
            self.canvas.tag_bind(tag, "<Enter>", self._on_keybind_enter)
            self.canvas.tag_bind(tag, "<Leave>", self._on_keybind_leave)
            self.canvas.tag_bind(tag, "<Button-1>", lambda e: self._on_keybind_clicked())

    def _on_keybind_enter(self, e):
        self.canvas.configure(cursor="hand2")
        if not self._is_capturing:
            self.canvas.itemconfig(self.keybind_poly, fill=theme.PILL_HOVER)

    def _on_keybind_leave(self, e):
        self.canvas.configure(cursor="")
        if not self._is_capturing:
            self.canvas.itemconfig(self.keybind_poly, fill=theme.PILL_BG)

    def _build_start_button(self, cx, cy, w, h):
        x1, y1 = cx - w / 2, cy - h / 2
        x2, y2 = cx + w / 2, cy + h / 2
        r = h / 2

        self.start_poly = theme.draw_rounded_rect(
            self.canvas, x1, y1, x2, y2, r, fill=theme.START_BG, outline=""
        )
        self.start_text = self.canvas.create_text(
            cx,
            cy,
            text="Start",
            fill=theme.TEXT_PRIMARY,
            font=self.font_button,
            anchor="center",
        )

        for tag in (self.start_poly, self.start_text):
            self.canvas.tag_bind(tag, "<Enter>", self._on_start_enter)
            self.canvas.tag_bind(tag, "<Leave>", self._on_start_leave)
            self.canvas.tag_bind(tag, "<ButtonPress-1>", self._on_start_press)
            self.canvas.tag_bind(tag, "<ButtonRelease-1>", self._on_start_release)

    def _on_start_enter(self, e):
        self.canvas.configure(cursor="hand2")
        fill = theme.STOP_HOVER if self.engine.is_running() else theme.START_HOVER
        self.canvas.itemconfig(self.start_poly, fill=fill)

    def _on_start_leave(self, e):
        self.canvas.configure(cursor="")
        fill = theme.STOP_BG if self.engine.is_running() else theme.START_BG
        self.canvas.itemconfig(self.start_poly, fill=fill)

    def _on_start_press(self, e):
        fill = theme.STOP_PRESSED if self.engine.is_running() else theme.START_PRESSED
        self.canvas.itemconfig(self.start_poly, fill=fill)

    def _on_start_release(self, e):
        self._on_start_enter(e)
        self._on_toggle_clicked()

    def _build_github_button(self, cx, cy):
        try:
            self.github_photo = tk.PhotoImage(data=GITHUB_BADGE_B64)
            self.github_img = self.canvas.create_image(
                cx, cy, image=self.github_photo, anchor="center"
            )
            self.canvas.tag_bind(
                self.github_img, "<Enter>", lambda e: self.canvas.configure(cursor="hand2")
            )
            self.canvas.tag_bind(
                self.github_img, "<Leave>", lambda e: self.canvas.configure(cursor="")
            )
            self.canvas.tag_bind(
                self.github_img, "<Button-1>", lambda e: self._on_github_clicked()
            )
        except Exception:
            pass

    def _build_mouse_button_pill(self, cx, cy, w, h):
        x1, y1 = cx - w / 2, cy - h / 2
        x2, y2 = cx + w / 2, cy + h / 2
        r = h / 2

        self.mouse_poly = theme.draw_rounded_rect(
            self.canvas, x1, y1, x2, y2, r, fill=theme.PILL_BG, outline=""
        )
        self.mouse_text = self.canvas.create_text(
            cx,
            cy,
            text=self.mouse_button_name,
            fill=theme.TEXT_MUTED,
            font=self.font_small,
            anchor="center",
        )

        for tag in (self.mouse_poly, self.mouse_text):
            self.canvas.tag_bind(
                tag, "<Enter>", lambda e: (
                    self.canvas.configure(cursor="hand2"),
                    self.canvas.itemconfig(self.mouse_poly, fill=theme.PILL_HOVER)
                )
            )
            self.canvas.tag_bind(
                tag, "<Leave>", lambda e: (
                    self.canvas.configure(cursor=""),
                    self.canvas.itemconfig(self.mouse_poly, fill=theme.PILL_BG)
                )
            )
            self.canvas.tag_bind(
                tag, "<Button-1>", lambda e: self._toggle_mouse_button()
            )

    # -------------------------------------------------------------
    # LOGIC & INTERACTIONS
    # -------------------------------------------------------------
    def _restore_saved_inputs(self, saved):
        if "cps1" in saved and saved["cps1"]:
            self.cps1_entry["var"].set(saved["cps1"])
            self.cps1_entry["entry"].configure(fg=theme.TEXT_INPUT)
        if "cps2" in saved and saved["cps2"]:
            self.cps2_entry["var"].set(saved["cps2"])
            self.cps2_entry["entry"].configure(fg=theme.TEXT_INPUT)

    def _save_inputs_to_config(self):
        v1 = self.cps1_entry["var"].get().strip()
        v2 = self.cps2_entry["var"].get().strip()
        config.save_config({
            "cps1": v1 if v1 != "Value" else "",
            "cps2": v2 if v2 != "Value" else "",
            "mouse_button": self.mouse_button_name,
        })

    def _toggle_mouse_button(self):
        self.mouse_button_name = "Right" if self.mouse_button_name == "Left" else "Left"
        self.canvas.itemconfig(self.mouse_text, text=self.mouse_button_name)
        self.engine.configure(self.engine.min_cps, self.engine.max_cps, self.mouse_button_name)
        self._save_inputs_to_config()

    def _on_github_clicked(self):
        webbrowser.open_new_tab("https://github.com/summersalestart/hyclicker")

    def _show_toast(self, message, is_error=True):
        if self._error_timer:
            self.root.after_cancel(self._error_timer)
            self._error_timer = None
        color = theme.DOT_RED if is_error else theme.START_BG
        self.canvas.itemconfig(self.toast_text, text=message, fill=color)
        self._error_timer = self.root.after(2500, self._clear_toast)

    def _clear_toast(self):
        self.canvas.itemconfig(self.toast_text, text="")
        self._error_timer = None

    def _validate_and_apply(self) -> bool:
        raw1 = self.cps1_entry["var"].get().strip()
        raw2 = self.cps2_entry["var"].get().strip()

        # Fallback to default speeds if left as "Value" or empty
        val1_str = self.cps1_entry["default"] if (not raw1 or raw1 == "Value") else raw1
        val2_str = self.cps2_entry["default"] if (not raw2 or raw2 == "Value") else raw2

        try:
            min_cps = float(val1_str)
            max_cps = float(val2_str)
        except ValueError:
            self._show_toast("CPS values must be numbers")
            return False

        if min_cps <= 0 or max_cps <= 0:
            self._show_toast("CPS must be greater than 0")
            return False

        if min_cps > max_cps:
            self._show_toast("CPS (1) cannot exceed CPS (2)")
            return False

        if max_cps > MAX_ALLOWED_CPS:
            self._show_toast(f"Max CPS capped at {MAX_ALLOWED_CPS}")
            return False

        self._clear_toast()
        self.engine.configure(min_cps, max_cps, self.mouse_button_name)
        return True

    def _on_toggle_clicked(self):
        if self.engine.is_running():
            self.engine.stop()
        else:
            if self._validate_and_apply():
                self.engine.start()

    def _on_hotkey_triggered(self):
        self.root.after(0, self._on_toggle_clicked)

    def _on_keybind_clicked(self):
        if self._is_capturing:
            return
        self._is_capturing = True
        self.canvas.itemconfig(self.keybind_text, text="...", fill=theme.TEXT_PRIMARY)
        self.canvas.itemconfig(self.keybind_poly, fill=theme.PILL_FOCUS)
        self.keybinds.begin_capture()

    def _on_keybind_captured(self, display_str):
        def update():
            self._is_capturing = False
            self.canvas.itemconfig(self.keybind_text, text=display_str, fill=theme.TEXT_MUTED)
            self.canvas.itemconfig(self.keybind_poly, fill=theme.PILL_BG)

        self.root.after(0, update)
        config.save_config({
            "hotkey_str": self.keybinds.hotkey_string,
            "display_str": display_str,
        })

    def _on_engine_state_change(self, is_running: bool):
        self.root.after(0, self._update_start_button, is_running)

    def _update_start_button(self, is_running: bool):
        if is_running:
            self.canvas.itemconfig(self.start_poly, fill=theme.STOP_BG)
            self.canvas.itemconfig(self.start_text, text="Stop")
        else:
            self.canvas.itemconfig(self.start_poly, fill=theme.START_BG)
            self.canvas.itemconfig(self.start_text, text="Start")

    # -------------------------------------------------------------
    # WINDOW MANAGEMENT & DRAGGING
    # -------------------------------------------------------------
    def _bind_draggable(self, item):
        self.canvas.tag_bind(item, "<ButtonPress-1>", self._start_drag)
        self.canvas.tag_bind(item, "<B1-Motion>", self._do_drag)
        if item not in self._draggable_items:
            self._draggable_items.append(item)

    def _start_drag(self, event):
        self._drag_offset = (
            event.x_root - self.root.winfo_x(),
            event.y_root - self.root.winfo_y(),
        )

    def _do_drag(self, event):
        x = event.x_root - self._drag_offset[0]
        y = event.y_root - self._drag_offset[1]
        self.root.geometry(f"+{x}+{y}")

    def _on_minimize(self):
        self.root.overrideredirect(False)
        winstyle.apply_dark_titlebar(self.root)
        self.root.iconify()

    def _on_close(self):
        # Shut down background listeners/threads first, then destroy the
        # window on the next event-loop tick: destroying Tk from inside a
        # canvas event binding can abort the interpreter (Tcl "alloc:
        # invalid block" crash on exit).
        try:
            self.engine.shutdown()
        except Exception:
            pass
        try:
            self.keybinds.shutdown()
        except Exception:
            pass
        try:
            self.root.after(0, self._destroy)
        except Exception:
            self._destroy()

    def _destroy(self):
        try:
            self.root.destroy()
        except Exception:
            try:
                self.root.quit()
            except Exception:
                pass

import math
import tkinter.font as tkfont

# Colors matching the provided UI
BG = "#191919"
PANEL = "#242424"
PANEL_BORDER = "#2b2b2b"

PILL_BG = "#242424"
PILL_HOVER = "#2a2a2a"
PILL_FOCUS = "#2e2e2e"

TEXT_PRIMARY = "#ffffff"
TEXT_SECONDARY = "#dddddd"
TEXT_MUTED = "#3a3a3a"
TEXT_INPUT = "#ffffff"

DOT_YELLOW = "#f0d155"
DOT_YELLOW_HOVER = "#fae07a"

DOT_RED = "#e34f4f"
DOT_RED_HOVER = "#f06c6c"

DIVIDER = "#303030"

START_BG = "#00bf63"
START_HOVER = "#02d66f"
START_PRESSED = "#00a855"

STOP_BG = "#e34f4f"
STOP_HOVER = "#f06c6c"
STOP_PRESSED = "#c73838"

TRANSPARENT_KEY = "#010203"

def exact_rounded_rect_points(x1, y1, x2, y2, r, n_points=12):
    r = min(r, (x2 - x1) / 2, (y2 - y1) / 2)
    points = []
    corners = [
        (x2 - r, y1 + r, -math.pi / 2, 0),
        (x2 - r, y2 - r, 0, math.pi / 2),
        (x1 + r, y2 - r, math.pi / 2, math.pi),
        (x1 + r, y1 + r, math.pi, 3 * math.pi / 2),
    ]
    for cx, cy, start_ang, end_ang in corners:
        for i in range(n_points + 1):
            ang = start_ang + (end_ang - start_ang) * (i / n_points)
            points.append(cx + r * math.cos(ang))
            points.append(cy + r * math.sin(ang))
    return points

def draw_rounded_rect(canvas, x1, y1, x2, y2, r, **kwargs):
    points = exact_rounded_rect_points(x1, y1, x2, y2, r)
    return canvas.create_polygon(points, smooth=False, **kwargs)

def pick_font(root, candidates):
    try:
        families = set(tkfont.families(root))
    except Exception:
        return candidates[-1]
    for candidate in candidates:
        if candidate[0] in families:
            return candidate
    return candidates[-1]

FONT_TITLE_CANDIDATES = [
    ("Segoe UI", -36),
    ("Bahnschrift", -35),
    ("Arial", -35),
]

FONT_LABEL_CANDIDATES = [
    ("Segoe UI", -22),
    ("Bahnschrift", -22),
    ("Arial", -22),
]

FONT_INPUT_CANDIDATES = [
    ("Segoe UI", -28),
    ("Bahnschrift", -28),
    ("Arial", -28),
]

FONT_BUTTON_CANDIDATES = [
    ("Segoe UI", -26),
    ("Bahnschrift", -26),
    ("Arial", -26),
]

FONT_SMALL_CANDIDATES = [
    ("Segoe UI", -20),
    ("Bahnschrift", -20),
    ("Arial", -20),
]

FONT_TOAST_CANDIDATES = [
    ("Segoe UI", -13),
    ("Bahnschrift", -13),
    ("Arial", -13),
]

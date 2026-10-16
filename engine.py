import random
import threading
import time

from pynput.mouse import Controller, Button

MIN_DELAY = 0.03
MAX_DELAY = 0.5
PHASE_MIN_DURATION = 0.5
PHASE_MAX_DURATION = 2.0
JITTER_FRACTION = 0.15
MAX_ALLOWED_CPS = 50

START_GRACE_PERIOD = 0.35

BUTTON_MAP = {
    "Left": Button.left,
    "Right": Button.right,
    "Middle": Button.middle,
}

class ClickEngine:

    def __init__(self, on_state_change=None):
        self._thread = None
        self._running = threading.Event()
        self._stop_flag = threading.Event()
        self._lock = threading.Lock()

        self.min_cps = 8.0
        self.max_cps = 12.0
        self.button = Button.left

        self._mouse = Controller()
        self._on_state_change = on_state_change
        self._last_phase_cps = None

    def configure(self, min_cps: float, max_cps: float, button_name: str = "Left"):
        with self._lock:
            self.min_cps = min_cps
            self.max_cps = max_cps
            self.button = BUTTON_MAP.get(button_name, Button.left)

    def start(self):
        if self.is_running():
            return
        self._stop_flag.clear()
        self._running.set()
        self._thread = threading.Thread(target=self._run_loop, daemon=True)
        self._thread.start()
        self._notify(True)

    def stop(self):
        if not self.is_running():
            return
        self._stop_flag.set()
        self._running.clear()
        self._notify(False)

    def toggle(self):
        if self.is_running():
            self.stop()
        else:
            self.start()

    def is_running(self) -> bool:
        return self._running.is_set()

    def shutdown(self):
        self._stop_flag.set()
        self._running.clear()
        if self._thread and self._thread.is_alive():
            self._thread.join(timeout=1.0)

    def _notify(self, is_running: bool):
        if self._on_state_change:
            try:
                self._on_state_change(is_running)
            except Exception:
                pass

    def _pick_next_phase_cps(self, min_cps, max_cps, avoid_value=None):
        value = random.uniform(min_cps, max_cps)
        span = max_cps - min_cps
        if avoid_value is not None and span > 0.5:
            attempts = 0
            while abs(value - avoid_value) < span * 0.15 and attempts < 5:
                value = random.uniform(min_cps, max_cps)
                attempts += 1
        return value

    def _run_loop(self):
        phase_end_time = 0.0
        target_cps = None

        slept = 0.0
        step = 0.02
        while slept < START_GRACE_PERIOD and not self._stop_flag.is_set():
            t = min(step, START_GRACE_PERIOD - slept)
            time.sleep(t)
            slept += t

        while not self._stop_flag.is_set():
            now = time.monotonic()

            with self._lock:
                min_cps = max(0.1, self.min_cps)
                max_cps = max(min_cps, self.max_cps)
                button = self.button

            if target_cps is None or now >= phase_end_time:
                target_cps = self._pick_next_phase_cps(
                    min_cps, max_cps, avoid_value=self._last_phase_cps
                )
                self._last_phase_cps = target_cps
                phase_duration = random.uniform(PHASE_MIN_DURATION, PHASE_MAX_DURATION)
                phase_end_time = now + phase_duration

            ideal_delay = 1.0 / target_cps
            jitter = random.uniform(-JITTER_FRACTION, JITTER_FRACTION)
            delay = max(MIN_DELAY, min(MAX_DELAY, ideal_delay * (1.0 + jitter)))

            try:
                self._mouse.click(button, 1)
            except Exception:
                break

            slept = 0.0
            step = 0.02
            while slept < delay and not self._stop_flag.is_set():
                t = min(step, delay - slept)
                time.sleep(t)
                slept += t

        self._running.clear()

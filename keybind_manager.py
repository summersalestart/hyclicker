from pynput import keyboard

_MODIFIER_ATTRS = {
    "ctrl": ["ctrl", "ctrl_l", "ctrl_r"],
    "alt": ["alt", "alt_l", "alt_r", "alt_gr"],
    "shift": ["shift", "shift_l", "shift_r"],
    "cmd": ["cmd", "cmd_l", "cmd_r"],
}
_MODIFIER_ORDER = ["ctrl", "alt", "shift", "cmd"]

def _build_modifier_lookup():
    lookup = {}
    for canonical, attr_names in _MODIFIER_ATTRS.items():
        for attr_name in attr_names:
            key_obj = getattr(keyboard.Key, attr_name, None)
            if key_obj is not None:
                lookup[key_obj] = canonical
    return lookup

_MODIFIER_LOOKUP = _build_modifier_lookup()

def normalize_modifier(key):
    return _MODIFIER_LOOKUP.get(key)

def normalize_base_key(key):
    if isinstance(key, keyboard.KeyCode):
        vk = getattr(key, "vk", None)
        if vk is not None and 65 <= vk <= 90:
            letter = chr(vk).lower()
            return letter, letter.upper()
        if vk is not None and 48 <= vk <= 57:
            digit = chr(vk)
            return digit, digit
        if key.char is not None and key.char.isprintable():
            return key.char.lower(), key.char.upper()
        return None, None
    name = str(key).split(".")[-1]
    return f"<{name}>", name.replace("_", " ").title()

class KeybindManager:

    def __init__(self, on_trigger, on_capture_complete=None,
                 initial_hotkey_str=None, initial_display_str=None):
        self._on_trigger = on_trigger
        self._on_capture_complete = on_capture_complete

        self._hotkey_str = initial_hotkey_str or "<f6>"
        self._display_str = initial_display_str or "F6"
        self._hotkey_listener = None
        self._capture_listener = None
        self._held_modifiers = set()

        self._start_global_listener()

    @property
    def display_string(self) -> str:
        return self._display_str

    @property
    def hotkey_string(self) -> str:
        return self._hotkey_str

    def begin_capture(self):
        self._stop_capture_listener()
        self._held_modifiers = set()

        def on_press(key):
            mod = normalize_modifier(key)
            if mod:
                self._held_modifiers.add(mod)
                return

            hotkey_token, display_token = normalize_base_key(key)
            if hotkey_token is None:
                return

            mods_in_order = [m for m in _MODIFIER_ORDER if m in self._held_modifiers]
            hotkey_str = "+".join([f"<{m}>" for m in mods_in_order] + [hotkey_token])
            display_str = "+".join([m.upper() for m in mods_in_order] + [display_token])

            self._rebind(hotkey_str, display_str)
            if self._on_capture_complete:
                self._on_capture_complete(display_str)
            return False

        def on_release(key):
            mod = normalize_modifier(key)
            if mod:
                self._held_modifiers.discard(mod)

        self._capture_listener = keyboard.Listener(
            on_press=on_press, on_release=on_release
        )
        self._capture_listener.start()

    def shutdown(self):
        self._stop_capture_listener()
        self._stop_global_listener()

    def _rebind(self, hotkey_str: str, display_str: str):
        self._hotkey_str = hotkey_str
        self._display_str = display_str
        self._start_global_listener()

    def _start_global_listener(self):
        self._stop_global_listener()
        try:
            self._hotkey_listener = keyboard.GlobalHotKeys(
                {self._hotkey_str: self._on_trigger}
            )
            self._hotkey_listener.start()
        except Exception:
            self._hotkey_listener = None

    def _stop_global_listener(self):
        if self._hotkey_listener is not None:
            try:
                self._hotkey_listener.stop()
            except Exception:
                pass
            self._hotkey_listener = None

    def _stop_capture_listener(self):
        if self._capture_listener is not None:
            try:
                self._capture_listener.stop()
            except Exception:
                pass
            self._capture_listener = None

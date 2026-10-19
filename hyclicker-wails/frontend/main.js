const $ = (id) => document.getElementById(id);
const cps1 = $("cps1"), cps2 = $("cps2");
const keybindBtn = $("keybindBtn"), startBtn = $("startBtn");
const toast = $("toast"), mouseBtn = $("mouseBtn"), rocketBtn = $("rocketBtn");
const updateBar = $("updateBar"), updateBtn = $("updateBtn"), updateText = $("updateText");

const app = window.go.main.App;

let running = false;
let capturing = false;

function showToast(msg) {
  toast.textContent = msg;
  clearTimeout(showToast._t);
  showToast._t = setTimeout(() => (toast.textContent = ""), 2500);
}

function setRunning(v) {
  running = v;
  startBtn.textContent = v ? "Stop" : "Start";
  startBtn.classList.toggle("stop", v);
}

async function restore() {
  try {
    const cfg = await app.GetConfig();
    if (cfg.min_cps) cps1.value = cfg.min_cps;
    if (cfg.max_cps) cps2.value = cfg.max_cps;
    if (cfg.button) mouseBtn.textContent = cfg.button;
    if (cfg.hotkey_name) keybindBtn.textContent = cfg.hotkey_name;
    rocketBtn.classList.toggle("on", !!cfg.rocket);
  } catch {}
}
restore();

function parseCps() {
  const raw1 = (cps1.value || "").trim() || "8";
  const raw2 = (cps2.value || "").trim() || "12";
  const a = parseFloat(raw1), b = parseFloat(raw2);
  if (Number.isNaN(a) || Number.isNaN(b)) { showToast("CPS values must be numbers"); return null; }
  if (a <= 0 || b <= 0) { showToast("CPS must be greater than 0"); return null; }
  if (a > b) { showToast("CPS (1) cannot exceed CPS (2)"); return null; }
  if (b > 50) { showToast("Max CPS capped at 50"); return null; }
  return { a, b };
}

async function toggle() {
  if (running) {
    await app.Toggle(); // toggle handles stop; keep state in sync below
    setRunning(false);
    return;
  }
  const p = parseCps();
  if (!p) return;
  try {
    const err = await app.Configure(p.a, p.b, mouseBtn.textContent);
    if (err) { showToast(String(err)); return; }
  } catch (e) { showToast(String(e)); return; }
  await app.Toggle();
  setRunning(true);
}

startBtn.addEventListener("click", toggle);

window.runtime.EventsOn("clicker-state", (v) => setRunning(!!v));
window.runtime.EventsOn("request-toggle", () => toggle());
window.runtime.EventsOn("toast", (msg) => showToast(String(msg)));
window.runtime.EventsOn("update-available", (v) => {
  updateText.textContent = `Update available (v${v})`;
  updateBtn.textContent = "Get it";
  updateBar.hidden = false;
});
window.runtime.EventsOn("rocket-state", (v) => rocketBtn.classList.toggle("on", !!v));
rocketBtn.addEventListener("click", () => {
  app.SetRocket(!rocketBtn.classList.contains("on")).then((v) => rocketBtn.classList.toggle("on", !!v));
});
window.runtime.EventsOn("hotkey-changed", (name) => { keybindBtn.textContent = name; showToast(`${name} was free; bound it instead`); });

// keybind capture: click pill, then press a key
keybindBtn.addEventListener("click", () => {
  capturing = true;
  keybindBtn.textContent = "...";
  keybindBtn.classList.add("capturing");
});

window.addEventListener("keyup", (e) => {
  if (!capturing) return;
  if (e.ctrlKey || e.altKey || e.shiftKey || e.metaKey) return;
  const cfg = app.GetConfig().then((c) => {
    keybindBtn.textContent = c.hotkey_name || "F6";
  });
  capturing = false;
  keybindBtn.classList.remove("capturing");
}, true);

const NAME_MAP = {
  " ": "Space", ArrowUp: "Up", ArrowDown: "Down", ArrowLeft: "Left", ArrowRight: "Right",
  Escape: "Esc", Tab: "Tab", Enter: "Enter", Delete: "Del", Insert: "Ins",
};
window.addEventListener("keydown", async (e) => {
  if (!capturing) return;
  e.preventDefault();
  if (["Control", "Alt", "Shift", "Meta", "ControlRight", "ControlLeft", "ShiftLeft", "ShiftRight", "AltLeft", "AltRight", "MetaLeft", "MetaRight"].includes(e.key)) {
    return;
  }
  capturing = false;
  keybindBtn.classList.remove("capturing");
  if (e.key === "Escape") {
    const cfg = await app.GetConfig();
    keybindBtn.textContent = cfg.hotkey_name || "F6";
    return;
  }
  const key = NAME_MAP[e.key] || (/^F\d+$/.test(e.code) ? e.code.replace("Arrow", "").replace("Escape", "Esc") : (e.key.length === 1 ? e.key.toUpperCase() : e.code));
  const display = [e.ctrlKey && "Ctrl", e.altKey && "Alt", e.shiftKey && "Shift"].filter(Boolean).concat(key).join("+");
  try {
    const err = await app.BindHotkey(key, e.ctrlKey, e.altKey, e.shiftKey);
    if (err) { showToast(String(err)); restore(); return; }
    keybindBtn.textContent = display;
  } catch (err) {
    showToast(String(err));
  }
}, true);

const rt = window.runtime;
$("minBtn").addEventListener("click", () => rt.WindowMinimise());
$("closeBtn").addEventListener("click", () => rt.Quit());
$("githubBtn").addEventListener("click", () => rt.BrowserOpenURL("https://github.com/summersalestart/hyclicker"));
updateBtn.addEventListener("click", () => rt.BrowserOpenURL("https://github.com/summersalestart/hyclicker/releases/latest"));
mouseBtn.addEventListener("click", async () => {
  const next = mouseBtn.textContent === "Left" ? "Right" : "Left";
  mouseBtn.textContent = next;
  const cfg = await app.GetConfig();
  await app.Configure(cfg.min_cps, cfg.max_cps, next);
});
[cps1, cps2].forEach((el) =>
  el.addEventListener("change", async () => {
    const cfg = await app.GetConfig();
    await app.Configure(
      parseFloat((cps1.value || "").trim()) || cfg.min_cps,
      parseFloat((cps2.value || "").trim()) || cfg.max_cps,
      cfg.button
    ).catch(() => {});
  })
);

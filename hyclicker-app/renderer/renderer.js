const $ = (id) => document.getElementById(id);
const cps1 = $("cps1"), cps2 = $("cps2");
const keybindBtn = $("keybindBtn"), startBtn = $("startBtn");
const toast = $("toast"), mouseBtn = $("mouseBtn");
const api = window.hyclicker;

let running = false;
let capturing = false;
let hotkeyDisplay = "F6";
let hotkeyAccel = "F6";

function showToast(msg) {
  toast.textContent = msg;
  clearTimeout(showToast._t);
  showToast._t = setTimeout(() => (toast.textContent = ""), 2500);
}

async function restore() {
  try {
    const cfg = await api.loadConfig();
    if (cfg.cps1) cps1.value = cfg.cps1;
    if (cfg.cps2) cps2.value = cfg.cps2;
    if (cfg.mouse_button) mouseBtn.textContent = cfg.mouse_button;
    if (cfg.hotkeyDisplay) {
      hotkeyDisplay = cfg.hotkeyDisplay;
      hotkeyAccel = cfg.hotkeyAccel || cfg.hotkeyDisplay;
      keybindBtn.textContent = hotkeyDisplay;
      // re-register the saved hotkey (main defaults to F6 at launch)
      if (hotkeyAccel !== "F6") {
        try { await api.registerHotkey(hotkeyAccel); }
        catch { hotkeyDisplay = "F6"; hotkeyAccel = "F6"; keybindBtn.textContent = "F6"; }
      }
    }
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

function setRunning(v) {
  running = v;
  startBtn.textContent = v ? "Stop" : "Start";
  startBtn.classList.toggle("stop", v);
}

async function toggle() {
  if (running) {
    await api.stop();
    setRunning(false);
    return;
  }
  const p = parseCps();
  if (!p) return;
  try {
    await api.configure({ minCps: p.a, maxCps: p.b, button: mouseBtn.textContent });
  } catch (e) { showToast(String(e.message || e)); return; }
  await api.start();
  setRunning(true);
  api.saveConfig({ cps1: cps1.value.trim(), cps2: cps2.value.trim(), mouse_button: mouseBtn.textContent }).catch(() => {});
}

startBtn.addEventListener("click", toggle);
api.onState((v) => setRunning(!!v));
api.onRequestToggle(() => toggle());

// keybind capture: click pill, then press a key
keybindBtn.addEventListener("click", () => {
  capturing = true;
  keybindBtn.textContent = "...";
  keybindBtn.classList.add("capturing");
});

const ACCEL_MAP = {
  " ": "Space", ArrowUp: "Up", ArrowDown: "Down", ArrowLeft: "Left", ArrowRight: "Right",
  Escape: "Escape", Tab: "Tab", Enter: "Enter",
};
window.addEventListener("keydown", async (e) => {
  if (!capturing) return;
  e.preventDefault();
  capturing = false;
  keybindBtn.classList.remove("capturing");
  if (e.key === "Escape") { keybindBtn.textContent = hotkeyDisplay; return; }
  const mods = [];
  if (e.ctrlKey) mods.push("Control");
  if (e.altKey) mods.push("Alt");
  if (e.shiftKey) mods.push("Shift");
  let key = ACCEL_MAP[e.key] || (/^F\d+$/.test(e.code) ? e.code : (e.key.length === 1 ? e.key.toUpperCase() : e.code));
  const accel = [...mods, key].join("+");
  const display = [...mods.map((m) => (m === "Control" ? "Ctrl" : m)), key].join("+");
  try {
    await api.registerHotkey(accel);
    hotkeyDisplay = display; hotkeyAccel = accel;
    keybindBtn.textContent = display;
    api.saveConfig({ hotkeyDisplay: display, hotkeyAccel: accel }).catch(() => {});
  } catch {
    keybindBtn.textContent = hotkeyDisplay;
    showToast("Hotkey register failed");
  }
}, true);

$("minBtn").addEventListener("click", () => api.minimize());
$("closeBtn").addEventListener("click", () => api.close());
$("githubBtn").addEventListener("click", () => api.openGithub());
mouseBtn.addEventListener("click", () => {
  mouseBtn.textContent = mouseBtn.textContent === "Left" ? "Right" : "Left";
  api.saveConfig({ mouse_button: mouseBtn.textContent }).catch(() => {});
});
[cps1, cps2].forEach((el) =>
  el.addEventListener("change", () => {
    api.saveConfig({ cps1: cps1.value.trim(), cps2: cps2.value.trim() }).catch(() => {});
  })
);

// ---- update notice: opens GitHub releases page ----
const updateBar = $("updateBar"), updateBtn = $("updateBtn"), updateText = $("updateText");
api.onUpdateAvailable((v) => {
  updateText.textContent = `Update available (v${v})`;
  updateBtn.textContent = "Get it";
  updateBar.hidden = false;
});
updateBtn.addEventListener("click", () => api.openReleases());
api.getVersion().then((v) => { if (v) document.title = `HyClicker v${v}`; }).catch(() => {});

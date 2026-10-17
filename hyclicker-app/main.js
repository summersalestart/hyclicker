const { app, BrowserWindow, ipcMain, globalShortcut, shell } = require("electron");
const path = require("path");
const fs = require("fs");
const https = require("https");

const REPO = "summersalestart/hyclicker";

// ---- self-updater (portable-friendly, GitHub Releases based) ----
function fetchJson(url) {
  return new Promise((resolve, reject) => {
    https.get(url, { headers: { "User-Agent": "HyClicker-Updater" } }, (res) => {
      let data = "";
      res.on("data", (c) => (data += c));
      res.on("end", () => {
        if (res.statusCode !== 200) return reject(new Error(`HTTP ${res.statusCode}`));
        try { resolve(JSON.parse(data)); } catch (e) { reject(e); }
      });
    }).on("error", reject);
  });
}

function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    const follow = (u) => {
      https.get(u, { headers: { "User-Agent": "HyClicker-Updater" } }, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location)
          return follow(res.headers.location);
        if (res.statusCode !== 200) return reject(new Error(`HTTP ${res.statusCode}`));
        const out = fs.createWriteStream(dest);
        res.pipe(out);
        out.on("finish", () => out.close(resolve));
        out.on("error", reject);
      }).on("error", reject);
    };
    follow(url);
  });
}

function versionOf(tag) {
  return String(tag || "").replace(/^v/, "").trim();
}
function isNewer(candidate, current) {
  const c = candidate.split(".").map(Number), r = current.split(".").map(Number);
  for (let i = 0; i < 3; i++) if ((c[i] || 0) !== (r[i] || 0)) return (c[i] || 0) > (r[i] || 0);
  return false;
}

let updateInfo = null; // { version, assetUrl }

async function checkForUpdate() {
  try {
    const rel = await fetchJson(`https://api.github.com/repos/${REPO}/releases/latest`);
    const v = versionOf(rel.tag_name);
    if (!v || !isNewer(v, app.getVersion())) return null;
    const asset = (rel.assets || []).find((a) => /HyClicker(\s*\(1\))?\.exe$/i.test(a.name) || /^HyClicker.*\.exe$/i.test(a.name));
    if (!asset) return null;
    updateInfo = { version: v, assetUrl: asset.browser_download_url };
    win?.webContents.send("update-available", v);
    return updateInfo;
  } catch {
    return null; // offline / rate-limited: stay silent
  }
}

function applyUpdate() {
  if (!updateInfo) return;
  const exe = process.execPath;
  const exeDir = path.dirname(exe);
  const newFile = path.join(exeDir, "HyClicker-new.exe");
  const oldFile = path.join(exeDir, "HyClicker-old.exe");
  win?.webContents.send("update-progress", "Downloading...");
  downloadFile(updateInfo.assetUrl, newFile)
    .then(() => {
      // rename running exe (allowed on Windows), move new into place, relaunch
      const { spawn } = require("child_process");
      fs.renameSync(exe, oldFile);
      fs.renameSync(newFile, exe);
      spawn("cmd", ["/c", `timeout /t 2 /nobreak >nul & start "" "${exe}" & del /q "${oldFile}"`], {
        detached: true, stdio: "ignore", windowsHide: true,
      }).unref();
      app.quit();
    })
    .catch(() => {
      try { fs.unlinkSync(newFile); } catch {}
      win?.webContents.send("update-progress", "Update failed");
    });
}

let win = null;

// ---- click engine (port of engine.py) ----
const engine = {
  running: false,
  minCps: 8,
  maxCps: 12,
  button: "Left",
  timer: null,
  lastPhaseCps: null,
  targetCps: null,
  phaseEnd: 0,
};

function pickPhaseCps(min, max, avoid) {
  let v = min + Math.random() * (max - min);
  const span = max - min;
  if (avoid != null && span > 0.5) {
    for (let i = 0; i < 5; i++) {
      if (Math.abs(v - avoid) >= span * 0.15) break;
      v = min + Math.random() * (max - min);
    }
  }
  return v;
}

let nutMouse = null, nutButton = null;
async function nut() {
  if (!nutMouse) {
    const nutjs = require("@nut-tree-fork/nut-js");
    nutMouse = nutjs.mouse;
    nutButton = nutjs.Button;
    nutMouse.config.autoDelayMs = 0;
  }
  return { mouse: nutMouse, Button: nutButton };
}

async function clickOnce() {
  try {
    const { mouse, Button } = await nut();
    const b = engine.button === "Right" ? Button.RIGHT : Button.LEFT;
    await mouse.click(b);
  } catch {}
}

function scheduleNext() {
  if (!engine.running) return;
  const now = Date.now();
  const min = Math.max(0.1, engine.minCps);
  const max = Math.max(min, engine.maxCps);
  if (engine.targetCps == null || now >= engine.phaseEnd) {
    engine.targetCps = pickPhaseCps(min, max, engine.lastPhaseCps);
    engine.lastPhaseCps = engine.targetCps;
    engine.phaseEnd = now + (500 + Math.random() * 1500);
  }
  const ideal = 1000 / Math.max(0.1, engine.targetCps);
  const jitter = 1 + (Math.random() * 0.3 - 0.15);
  const delay = Math.min(500, Math.max(30, ideal * jitter));
  engine.timer = setTimeout(async () => {
    if (!engine.running) return;
    await clickOnce();
    scheduleNext();
  }, delay);
}

function startClicking() {
  if (engine.running) return;
  engine.running = true;
  engine.targetCps = null;
  // grace period so the Start click itself isn't auto-clicked
  setTimeout(() => { if (engine.running) scheduleNext(); }, 350);
  win?.webContents.send("clicker-state", true);
}
function stopClicking() {
  engine.running = false;
  clearTimeout(engine.timer);
  win?.webContents.send("clicker-state", false);
}

// ---- config ----
function configPath() {
  return path.join(app.getPath("userData"), "hyclicker.json");
}
function loadConfig() {
  try { return JSON.parse(fs.readFileSync(configPath(), "utf8")); }
  catch { return {}; }
}
function saveConfig(patch) {
  const cur = loadConfig();
  fs.writeFileSync(configPath(), JSON.stringify({ ...cur, ...patch }, null, 2));
}

// ---- window ----
function createWindow() {
  win = new BrowserWindow({
    width: 342,
    height: 382,
    resizable: false,
    frame: false,
    transparent: true,
    hasShadow: false,
    icon: path.join(__dirname, "assets", "icon.ico"),
    webPreferences: { preload: path.join(__dirname, "preload.js") },
  });
  win.loadFile(path.join(__dirname, "renderer", "index.html"));
}

function registerHotkey(accelerator) {
  globalShortcut.unregisterAll();
  // globalShortcut.register returns false (not throw) on failure
  return globalShortcut.register(accelerator, () => {
    // ask renderer to validate CPS first
    win?.webContents.send("request-toggle");
  });
}

app.whenReady().then(() => {
  createWindow();
  registerHotkey("F6");
  win.webContents.on("did-finish-load", () => checkForUpdate());
  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
  });
});

app.on("will-quit", () => globalShortcut.unregisterAll());
app.on("window-all-closed", () => {
  if (process.platform !== "darwin") app.quit();
});

// ---- IPC ----
ipcMain.handle("configure", (_e, { minCps, maxCps, button }) => {
  if (!(minCps > 0 && maxCps > 0)) throw new Error("CPS must be greater than 0");
  if (minCps > maxCps) throw new Error("CPS (1) cannot exceed CPS (2)");
  if (maxCps > 50) throw new Error("Max CPS capped at 50");
  engine.minCps = minCps; engine.maxCps = maxCps; engine.button = button;
});
ipcMain.handle("start", () => startClicking());
ipcMain.handle("stop", () => stopClicking());
ipcMain.handle("is-running", () => engine.running);
ipcMain.handle("register-hotkey", (_e, accelerator) => {
  const ok = registerHotkey(accelerator);
  if (!ok) throw new Error("Hotkey register failed");
});
ipcMain.handle("minimize", () => win?.minimize());
ipcMain.handle("close", () => { stopClicking(); win?.close(); });
ipcMain.handle("open-github", () => shell.openExternal("https://github.com/summersalestart/hyclicker"));
ipcMain.handle("load-config", () => loadConfig());
ipcMain.handle("save-config", (_e, patch) => saveConfig(patch));
ipcMain.handle("get-version", () => app.getVersion());
ipcMain.handle("apply-update", () => applyUpdate());

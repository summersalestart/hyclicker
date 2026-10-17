const { contextBridge, ipcRenderer } = require("electron");

contextBridge.exposeInMainWorld("hyclicker", {
  configure: (args) => ipcRenderer.invoke("configure", args),
  start: () => ipcRenderer.invoke("start"),
  stop: () => ipcRenderer.invoke("stop"),
  registerHotkey: (accelerator) => ipcRenderer.invoke("register-hotkey", accelerator),
  minimize: () => ipcRenderer.invoke("minimize"),
  close: () => ipcRenderer.invoke("close"),
  openGithub: () => ipcRenderer.invoke("open-github"),
  loadConfig: () => ipcRenderer.invoke("load-config"),
  saveConfig: (patch) => ipcRenderer.invoke("save-config", patch),
  onState: (cb) => ipcRenderer.on("clicker-state", (_e, v) => cb(v)),
  onRequestToggle: (cb) => ipcRenderer.on("request-toggle", () => cb()),
  onUpdateAvailable: (cb) => ipcRenderer.on("update-available", (_e, v) => cb(v)),
  getVersion: () => ipcRenderer.invoke("get-version"),
  openReleases: () => ipcRenderer.invoke("open-releases"),
});

<div align="center">

  <img src="banner.png" alt="HyClicker">

</div>

 

<p align="center">

  <a href="https://github.com/summersalestart/hyclicker/releases">

    <img src="https://img.shields.io/github/v/release/summersalestart/hyclicker" alt="Latest Release">

  </a>

  <a href="https://github.com/summersalestart/hyclicker/commits/main">

    <img src="https://img.shields.io/github/last-commit/summersalestart/hyclicker" alt="Last Commit">

  </a>

  <a href="https://github.com/summersalestart/hyclicker/releases">

    <img src="https://img.shields.io/github/downloads/summersalestart/hyclicker/total" alt="Downloads">

  </a>

  <a href="https://github.com/summersalestart/hyclicker/stargazers">

    <img src="https://img.shields.io/github/stars/summersalestart/hyclicker" alt="Stars">

  </a>

</p>

# HyClicker

Simple dark themed auto clicker with a clean and basic UI Includes a global hotkey to toggle on or off

## How it works

Every 0.5–2 seconds it picks a new target CPS somewhere inside your min/max range then fires clicks at that rate

If you set min to 8 and max to 12 it will only use cps around that range

## Usage

- CPS (1) / CPS (2) the range it drifts between

- Start short delay after starting

- Keybind press whatever key you want the toggle to be.

- Left click to swap to RMB

## Running from source

```bash

cd hyclicker-app

npm install

npm start

```

## Building

```bash

cd hyclicker-app

npm run dist

```

The exe ends up in hyclicker-app/dist/.

## Releases / updates

Detects version on every launch, if theres a old version running, you will be prompted to install the new one which you can do so from releases

## Disclaimers

- This is unsigned, smartscreen might show warnings

- A lot of games ban auto clickers, read TOS before using this app any bans issued after using hyclicker are not on me


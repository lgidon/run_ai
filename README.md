# Llamafile TUI – Local LLM Runner & Configuration Manager

A terminal-based interface (TUI) for **llamafile** – easily discover, configure, and run GGUF language models locally. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss), the launcher guides you through setup and keeps your configuration persistent.

![License: MIT](https://img.shields.io/badge/license-MIT-blue) ![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8)

## 📖 Overview

Llamafile TUI is a one-stop terminal app that:

- Finds and selects your **llamafile** executable.
- Browses your filesystem to choose a **models directory** (where `.gguf` files live).
- Configures the **server port**.
- Lists available models and lets you pick one.
- Offers a choice of **context window sizes** (8192–65536 tokens).
- Launches the model as a background HTTP server and shows the local API endpoint.
- Provides a **settings menu** to change paths, port, or reset the configuration at any time.

Everything is driven by the keyboard and styled with a modern TUI.

## ✨ Features

- **Interactive file browser** – navigate directories, switch drives (Windows supported), and confirm selections with `ENTER` or `SPACEBAR`.
- **Persistent configuration** – saves your paths and port to `config.json` so you don’t have to repeat setup.
- **Settings panel** – reconfigure binary path, models folder, port, or clear everything and restart the wizard without quitting the app.
- **Live status bar** – always know what the app is doing.
- **Model loading feedback** – shows a loading indicator while the LLM warms up.
- **Quick model switch** – stop the current model and return to the model list.
- **Global quit** – `Q` or `Ctrl+C` anywhere (except port input) gracefully shuts down the backend.

## 🚀 Getting Started

### Prerequisites

- **Go 1.21** or later installed.
- A [llamafile](https://github.com/Mozilla-Ocho/llamafile) binary (or any compatible executable) accessible on your system.
- One or more `.gguf` model files stored locally.

### Installation

```bash
git clone https://github.com/yourusername/llamafile-tui.git
cd llamafile-tui
go build -o lltui .

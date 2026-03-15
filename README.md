# DinDash 🚗💨

```text
    ____  _            ____                __
   / __ \(_)___  ____ / __ \____ ________ / /_
  / / / / / __ \/ __ `/ / / / __ `/ ___/ __ \
 / /_/ / / / / / /_/ / /_/ / /_/ (__  ) / / /
/_____/_/_/ /_/\__, /_____/\__,_/____/_/ /_/
              /____/
        1-DIN CAR AUDIO SYNCHRONIZER
```

[English] | [中文版](./README-zh.md)

> **An automated MP3 preprocessing tool specifically designed for 1-DIN car head units.**

**DinDash** is a high-efficiency synchronization tool built for legacy car audio hardware. It "flattens" complex music folder structures and resolves encoding or sorting issues that often plague non-Android 1-DIN head units.

## 🎯 Solving Two Core Frustrations

1.  **Garbled Filenames & Length Constraints:**:
    * **The Pain**: 1-DIN screens usually only display 8–12 characters and often do not support non-ASCII characters (like Chinese).
    * **The Solution**: Automatically extracts numeric sequences from the path (e.g., `09-01.mp3`), ensuring a concise display and correct sorting.
2.  **ID3 Tag Related Crashes**:
    * **The Pain**: Messy Artist/Album tags or embedded high-resolution cover art often cause head units to read slowly or even crash.
    * **The Solution**: Forcefully strips all tag metadata, setting only the Title to match the numeric filename for 100% hardware compatibility.


## 🚀 Performance

* **Manual Operation**: ~10–15 minutes (Moving, renaming, and cleaning tags manually).
* **DinDash**: **~6 seconds** (Fully automated processing).
* **macOS Friendly**: Automatically invokes `dot_clean` after execution to remove hidden `._` system files that often clutter USB drives on non-Apple devices.

## 🛠 System Requirements

This tool requires a Go environment and the command-line interface of [Kid3 - Audio Tag Editor](https://kid3.kde.org/).

### Installation (macOS)

```bash
brew install kid3
```

## 📦 Installation & Compilation

You can run the script directly using Go or compile it into a standalone binary.

- Option A: Run Directly
```bash
go run main.go -src [source_path] -dest [destination_path]
```

- Option B: Build Executable

1. Compile:
  ```bash
  go build -o dindash main.go
  ```

2. Move to System Path (Optional):
  ```bash
  mv dindash /usr/local/bin/
  ```

3. Run:
  ```bash
  dindash -src ~/Music -dest /Volumes/USB
  ```

## 📖 Usage

Use flag parameters to specify your source music folder and the destination (USB/Disk) path:

```bash
./dindash -src [source_path] -dest [destination_path]
```

### Example

```bash
./dindash -src ~/Documents/MP3 -dest /Volumes/USB_DISK/MP3_temp
```

## ⚙️ Processing Logic

1. Path Analysis: Analyzes the relative path while excluding numbers found within file extensions.
2. Numeric Renaming: Extracts all numbers from the path and joins them with hyphens (-) to create the new filename.
3. Duplicate Check: Automatically skips files that already exist at the destination to save write-cycles and time.
4. ID3 Tag Cleanup:
  - Title: Set to the new numeric filename.
  - Artist / Album: Cleared entirely to minimize file size and prevent parsing errors.
5. Disk Cleanup: Invokes macOS dot_clean to ensure the USB drive remains free of hidden system "junk" files.

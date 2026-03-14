package main

import (
	"flag"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func pathExist (path string) (bool, os.FileInfo) {
	info, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false, nil
	}

	return err == nil, info
}

func copyFile (srcPath string, destPath string) error {
	source, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func main() {
	var skipCount, copyCount, errCount int
	// 設定預設 Logger (讓輸出帶有時間與層級)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	srcFlagArgs := flag.String("src", "", "來源 MP3 資料夾路徑")
	destFlagArgs := flag.String("dest", "", "目的地路徑")

	flag.Parse()

	srcFolderPath := *srcFlagArgs
	destFolderPath := *destFlagArgs

	if destFolderPath == "" {
		flag.PrintDefaults() // 自動印出用法說明
		return
	}

	// --- 第一步：判斷 USB 是否存在 ---
	isExist, info := pathExist(destFolderPath)

	if !isExist {
		slog.Error("找不到 USB 裝置，請確認是否已插入並正確掛載。", "FOLDER", destFolderPath)
		os.Exit(1)
	}
	if info != nil && !info.IsDir() {
		slog.Error("路徑存在但不是一個有效的磁碟掛載點。", "FOLDER", destFolderPath)
		os.Exit(1)
	}

	// --- 第二步：(進階) 測試 USB 是否可寫入 ---
	testFile := filepath.Join(destFolderPath, ".write_test")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		slog.Error("無法寫入 USB！請檢查隨身碟是否開啟了唯讀開關或磁碟已滿", "ERR", err)
		os.Exit(1)
	}
	os.Remove(testFile)

	// 正規表達式：匹配所有數字
	re := regexp.MustCompile(`\d+`)

	// var musicFilePaths []string

	err = filepath.WalkDir(srcFolderPath, func(path string, dir os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// --- "." 開頭的檔案 (包含 .DS_Store 和 ._ 檔案) ---
		if strings.HasPrefix(dir.Name(), ".") {
			return nil
		}

		// 只處理 mp3 檔案
		if !dir.IsDir() && strings.ToLower(filepath.Ext(dir.Name())) == ".mp3" {
			// 1. 取得相對路徑，避免抓到來源資料夾之前的數字
			relPath, _ := filepath.Rel(srcFolderPath, path)

			// 2. 把副檔名 ".mp3" 砍掉，這樣正則就不會抓到末尾的 3
			purePath := strings.TrimSuffix(relPath, filepath.Ext(relPath))

			// 3. 提取所有數字
			numbers := re.FindAllString(purePath, -1)

			// 4. 組成新檔名 (例如: 09-01.mp3)
			var newFileName string
			if len(numbers) > 0 {
				newFileName = strings.Join(numbers, "-")

			} else {
				newFileName = "unknown-" + dir.Name()[:2]
			}

			// --- 判斷 MP3 在 USB 是否存在 ---
			destPath := filepath.Join(destFolderPath, newFileName + ".mp3")
			isExist, _ := pathExist(destPath)

			if isExist {
				skipCount++
				return nil
			}

			// 4. 執行複製
			slog.Info("正在複製", "to", newFileName+".mp3")
			err = copyFile(path, destPath)
			if err != nil {
			    slog.Error("複製失敗", "file", path, "ERR", err)
			    errCount++
			    return nil
			}
			slog.Info("複製成功", "to", newFileName+".mp3")

			cmd := exec.Command("kid3-cli", "-c", "set title '" + newFileName + "'", "-c", "set artist ''", "-c", "set album ''", destPath)

			if cmdErr := cmd.Run(); cmdErr != nil {
				slog.Error("ID3 Tag 清理失敗", "file", newFileName, "Err", cmdErr)
			} else {
				slog.Info("ID3 Tag 清理完成", "file", newFileName)
			}
			copyCount++
		}

		return nil
	})

	if err != nil {
		slog.Error("掃描過程發生錯誤", "ERR", err)
		os.Exit(1)
	}

	slog.Info("執行最後的磁碟清理...")
	exec.Command("dot_clean", "-m", destFolderPath).Run()

	slog.Info(strings.Repeat("-", 50))
	slog.Info("📊 執行統計報告",
		"總計", skipCount + copyCount + errCount,
		"新增", copyCount,
		"跳過", skipCount,
		"失敗", errCount,
	)
}

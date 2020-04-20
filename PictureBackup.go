package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"io"
)

func wantToBackupFile(testPath string) bool {
	switch path.Ext(testPath)[1:] {
	case
		"png",
		"jpg",
		"jpeg",
		"bmp",
		"gif",
		"tiff",
		"avi",
		"mpg",
		"mpeg",
		"m1v",
		"mp2",
		"mpe",
		"m3u",
		"ivf",
		"mov",
		"mp4",
		"m4v",
		"mp4v",
		"3g2",
		"3gp2",
		"3gp",
		"3gpp",
		"m2ts":
		return true
	}
	return false
}

func calculateBytesToTransfer(paths []string) (bytesToTransfer int64) {
	for _, pa := range paths {
		filepath.Walk(pa, func(p string, i os.FileInfo, e error) error {
			if e != nil {
				log.Println("[E] calculateBytesToTransfer() Error visiting path '", p, "':", e)
				return nil
			}

			if i.IsDir()  || !i.Mode().IsRegular() {
				return nil
			}			

			if wantToBackupFile(p) {
				bytesToTransfer += i.Size()
			}

			return nil
		})
	}
	return bytesToTransfer
}

func buildDestPath(srcRoot, fileToBackup, dstRootBackup string) (string, error) {
	relativeSrcPath, err := filepath.Rel(srcRoot, fileToBackup)
	if err != nil {
		return "", err
	}
	return filepath.Join(dstRootBackup, relativeSrcPath), nil
}

/* Copy from src to dst, and create dst file structure if needed */
func accommodatedCopyFile(src, dst string) (int64, error) {
        source, err := os.Open(src)
        if err != nil {
                return 0, err
        }
        defer source.Close()

        err = os.MkdirAll(filepath.Dir(dst), os.ModeDir)
        if err != nil {
        	return 0, err
        }

        destination, err := os.Create(dst)
        if err != nil {
                return 0, err
        }
        defer destination.Close()

        nBytes, err := io.Copy(destination, source)
        return nBytes, err
}

func backupPaths(srcRootPaths []string, dstRootPath string) error {

	log.Println("[i] Calculating backup size")
	bytesToTransfer := calculateBytesToTransfer(srcRootPaths)
	if 0 == bytesToTransfer {
		return fmt.Errorf("0 bytes to backup")
	}
	log.Println("[i] Backup size:", bytesToTransfer, "bytes")

	log.Println("[i] Copying files")
	var bytesTransfered int64 = 0

	for _, srcRootPath := range srcRootPaths {
		filepath.Walk(srcRootPath, func(p string, i os.FileInfo, e error) error {
			if e != nil {
				log.Println("[E] backupPaths() Error visiting path '", p, "':", e)
				return nil
			}

			if i.IsDir()  || !i.Mode().IsRegular()  {
				return nil
			}

			if !wantToBackupFile(p) {
				return nil
			}

			dstPath, err := buildDestPath(srcRootPath, p, dstRootPath)
			if err != nil {
				log.Println("[E] buildDestPath() error:", err)
				return nil
			}
			log.Println("[i] Will copy '", p, "' to '", dstPath, "'")
			nBytes, err := accommodatedCopyFile(p, dstPath)
			bytesTransfered += nBytes
			if err != nil {
				log.Println("[E] accommodatedCopyFile() error:", err)
			}
			
			return nil
		})
	}

	log.Println("[i] Done")
	return nil
}

func ensureValidDirs(dirPaths ...string) {
	for _, dp := range(dirPaths) {
        dpStat, err := os.Stat(dp)
        if err != nil {
            panic(fmt.Sprint("Can't os.Stat() '", dp, "' error: ", err))
        }

        if !dpStat.IsDir() {
        	panic(fmt.Sprint("Expected directory for path '", dp, "'"))
        }
	}
}

func main() {
	bkupPaths := []string{`C:\Users\Guy\Desktop\temp\bkuptest`}
	dstPath := `C:\Users\Guy\Desktop\temp\bkupdest`

	ensureValidDirs(bkupPaths...)
	ensureValidDirs(dstPath)
	err := backupPaths(bkupPaths, dstPath)
	if err != nil {
		log.Println("[E] backupPaths() error:", err)
	}
}

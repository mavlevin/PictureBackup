package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"io"
	"bufio"
)

const verboseDebug = false

func wantToBackupFile(testPath string) bool {
	switch path.Ext(testPath) {
	case
		".png",
		".jpg",
		".jpeg",
		".bmp",
		".gif",
		".tiff",
		".avi",
		".mpg",
		".mpeg",
		".m1v",
		".mp2",
		".mpe",
		".m3u",
		".ivf",
		".mov",
		".mp4",
		".m4v",
		".mp4v",
		".3g2",
		".3gp2",
		".3gp",
		".3gpp",
		".m2ts":
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

const totalCompletionStatusPrints = 10
var completionStatusPrints [totalCompletionStatusPrints+1]bool 
// +1 because start indexing from 1, to ignore 0% done

func logCompletionStatus(done, total int64) {
	percentDone := (float64(done) / float64(total))*100
	if verboseDebug{
		log.Println("Done", percentDone, "%")
	}

	printEveryPercent := float64(100.0) / totalCompletionStatusPrints
	completionStatusPrintNumber := int(percentDone / printEveryPercent)

	if(!completionStatusPrints[completionStatusPrintNumber]) {
		// first take care of a file that completed >1 milestones
		for i := 1; i < completionStatusPrintNumber; i++ {
			if !completionStatusPrints[i] {
			log.Printf("Done %.2f%%\n", float64(i)*printEveryPercent)
			completionStatusPrints[i] = true
			}
		}
		log.Printf("Done %.2f%%\n", percentDone)
		completionStatusPrints[completionStatusPrintNumber] = true
	}
}

func backupPaths(srcRootPaths []string, dstRootPath string) error {

	log.Println("[i] Calculating backup size")
	bytesToTransfer := calculateBytesToTransfer(srcRootPaths)
	if 0 == bytesToTransfer {
		return fmt.Errorf("0 bytes to backup")
	}
	log.Println("[+] Backup size:", bytesToTransfer, "bytes")

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
			if verboseDebug {
				log.Println("[i] Will copy '", p, "' to '", dstPath, "'")				
			}
			nBytes, err := accommodatedCopyFile(p, dstPath)
			bytesTransfered += nBytes
			if err != nil {
				log.Println("[E] accommodatedCopyFile() error:", err)
			}
			logCompletionStatus(bytesTransfered, bytesToTransfer)
			return nil
		})
	}

	log.Println("[+] Done backing up files")
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
	in := bufio.NewReader(os.Stdin)

	fmt.Println("Enter backup destination path: ")
	dstPath, err := in.ReadString('\n')
	if err != nil {
		panic(err)
	}
	dstPath = dstPath[:len(dstPath)-1]
	ensureValidDirs(dstPath)

	var bkupPaths []string
	for {
		fmt.Println("Enter a backup source path or 'done' to finish: ")
		srcPath, err := in.ReadString('\n')
		srcPath = srcPath[:len(srcPath)-1]
		if err != nil {
			panic(err)
		}
		if srcPath == "done" {
			break
		}
		bkupPaths = append(bkupPaths, srcPath)
	}
	ensureValidDirs(bkupPaths...)

	fmt.Println("Will backup from")
	for _, sp := range(bkupPaths) {
		fmt.Printf("\t%s\n", sp)
	}
	fmt.Println("To")
	fmt.Printf("\t%s\n", dstPath)

	fmt.Println("Enter 'c' to confirm")
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "c" {
		log.Println("MUST CONFIRM TO CONTINUE")
		return
	}

	err = backupPaths(bkupPaths, dstPath)
	if err != nil {
		log.Println("[E] backupPaths() error:", err)
	}
}

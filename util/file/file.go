package file

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

var (
	// ErrEmptyDirectoryName is returned by constructing the full path
	// of data directory if the passed argument is empty
	ErrEmptyDirectoryName = errors.New("data directory must not be empty")
	// ErrDotDirectoryName is returned by constructing the full path of
	// data directory if the passed argument is "."
	ErrDotDirectoryName = errors.New("data directory must not be equal to \".\"")
)

func ExistsWithInfo(path string) (exists bool, fileInfo os.FileInfo) {
	fileInfo, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false, nil
	}
	return true, fileInfo
}

func Exists(path string) bool {
	res, _ := ExistsWithInfo(path)
	return res
}

func ConcatFile(writeFile string, readFile string, removeReadFileIfSuccess bool) error {
	fi, err := os.Open(readFile)
	if err != nil {
		return err
	}
	defer fi.Close()
	fo, err := os.OpenFile(writeFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer fo.Close()
	buf := make([]byte, 8192)
	for {
		// read a chunk
		n, err := fi.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		// write a chunk
		if _, err := fo.Write(buf[:n]); err != nil {
			return err
		}
	}
	if removeReadFileIfSuccess {
		err = os.Remove(readFile)
		if err != nil {
			return nil
		}
	}
	return nil
}

// UserHome returns the current user home path
func UserHome() string {
	// os/user relies on cgo which is disabled when cross compiling
	// use fallbacks for various OSes instead
	// usr, err := user.Current()
	// if err == nil {
	// 	return usr.HomeDir
	// }
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}

	return os.Getenv("HOME")
}

// ResolveResourceDirectory searches locations for a research directory and returns absolute path
func ResolveResourceDirectory(path string) string {
	workDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}

	_, rtFilename, _, _ := runtime.Caller(1)
	rtDirectory := filepath.Dir(rtFilename)

	pathAbs, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	fmt.Println("abs path:", pathAbs)

	fmt.Printf("runtime.Caller= %s \n", rtFilename)
	//fmt.Printf("Filepath Raw= %s \n")
	fmt.Printf("Filepath Directory= %s \n", filepath.Dir(path))
	fmt.Printf("Filepath Absolute Directory= %s \n", pathAbs)

	fmt.Printf("Working Directory= %s \n", workDir)
	fmt.Printf("Runtime Filename= %s \n", rtFilename)
	fmt.Printf("Runtime Directory= %s \n", rtDirectory)

	//dir1 := filepath.Join(workDir, filepath.Dir(path))
	//fmt.Printf("Dir1= %s \n", dir1)

	dirs := []string{
		pathAbs, //try direct path first
		filepath.Join(workDir, filepath.Dir(path)), //default
		//filepath.Join(rt_directory, "./", filepath.Dir(path)),
		filepath.Join(rtDirectory, "./", filepath.Dir(path)),
		filepath.Join(rtDirectory, "../", filepath.Dir(path)),
		filepath.Join(rtDirectory, "../../", filepath.Dir(path)),
		filepath.Join(rtDirectory, "../../../", filepath.Dir(path)),
	}

	//for i, dir := range dirs {
	//	fmt.Printf("Dir[%d]= %s \n", i, dir)
	//}

	//must be an absolute path
	//error and problem and crash if not absolute path
	for i := range dirs {
		absPath, _ := filepath.Abs(dirs[i])
		dirs[i] = absPath
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			fmt.Printf("ResolveResourceDirectory: static resource dir= %s \n", dir)
			return dir
		}
	}
	panic("GUI directory not found")
	return ""
}

// DetermineResourcePath DEPRECATE
// From src/gui/http.go and src/mesh/gui/http.go
func DetermineResourcePath(staticDir string, resourceDir string, devDir string) (string, error) {
	//check "dev" directory first
	appLoc := filepath.Join(staticDir, devDir)
	// if !strings.HasPrefix(appLoc, "/") {
	// 	// Prepend the binary's directory path if appLoc is relative
	// 	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	// 	if err != nil {
	// 		return "", err
	// 	}

	// 	appLoc = filepath.Join(dir, appLoc)
	// }
	if _, err := os.Stat(appLoc); os.IsNotExist(err) {
		//check dist directory
		appLoc = filepath.Join(staticDir, resourceDir)
		// if !strings.HasPrefix(appLoc, "/") {
		// 	// Prepend the binary's directory path if appLoc is relative
		// 	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		// 	if err != nil {
		// 		return "", err
		// 	}

		// 	appLoc = filepath.Join(dir, appLoc)
		// }

		if _, err := os.Stat(appLoc); os.IsNotExist(err) {
			return "", err
		}
	}

	return appLoc, nil
}

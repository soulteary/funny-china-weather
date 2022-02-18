package main

import (
	"encoding/json"
	"fmt"
	"go/format"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	Minify "github.com/tdewolff/minify/v2"
	MinifyCSS "github.com/tdewolff/minify/v2/css"
	MinifyHTML "github.com/tdewolff/minify/v2/html"
	MinifySVG "github.com/tdewolff/minify/v2/svg"
)

const _TMP_ICON_PATH = "tmp/weather"
const _ASSETS_ICON_PATH = "icons"
const _TMP_ICON_CACHE = "icons.tmp"

func main() {
	os.RemoveAll(_TMP_ICON_PATH)
	_ = os.MkdirAll(_TMP_ICON_PATH, os.ModePerm)
	copyDirectory(_ASSETS_ICON_PATH, _TMP_ICON_PATH)
	minifySVGFiles(_TMP_ICON_PATH)
	createResourceFile(_TMP_ICON_PATH, "icons.go")
	os.RemoveAll(_TMP_ICON_PATH)
}

func scanFile(root, ext string) []string {
	var files []string
	filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ext {
			files = append(files, s)
		}
		return nil
	})
	return files
}

func minifySVGFiles(dirPath string) {
	m := Minify.New()
	m.AddFunc("text/html", MinifyHTML.Minify)
	m.Add("text/html", &MinifyHTML.Minifier{KeepDocumentTags: true, KeepQuotes: false})
	m.AddFunc("image/svg+xml", MinifySVG.Minify)
	m.AddFunc("text/css", MinifyCSS.Minify)

	files := scanFile(dirPath, ".svg")
	for _, file := range files {
		fileRaw, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Println("读取 SVG 文件出错，跳过处理", file)
		} else {
			fileMinified, err := m.Bytes("image/svg+xml", fileRaw)
			if err != nil {
				fmt.Println("压缩 SVG 文件出错，跳过处理", file, err)
			} else {
				err = os.WriteFile(file, fileMinified, os.ModePerm)
				if err != nil {
					fmt.Println("保存 SVG 文件出错，跳过处理", file, err)
				}
			}
		}
	}
}

func createResourceFile(src string, gofile string) {

	files, err := ioutil.ReadDir(src)
	if err != nil {
		log.Fatal(err)
		return
	}

	icons := make(map[string]string, len(files))
	for _, file := range files {
		if !file.IsDir() {
			fileName := file.Name()
			filePath := src + "/" + fileName
			fileRaw, err := ioutil.ReadFile(filePath)
			if err != nil {
				fmt.Println("读取 SVG 文件出错，跳过处理", fileName)
			} else {
				name := strings.ToLower(fileName)
				icons[strings.TrimSuffix(name, filepath.Ext(name))] = string(fileRaw)
			}
		}
	}

	file, _ := json.MarshalIndent(icons, "", " ")

	err = ioutil.WriteFile(_TMP_ICON_CACHE, file, os.ModePerm)
	if err != nil {
		fmt.Println("保存图标缓存文件失败", err)
	}

	jsonRaw, err := ioutil.ReadFile(_TMP_ICON_CACHE)
	if err != nil {
		log.Fatal(err)
		return
	}

	goFile := "package weather\nvar iconMap = map[string]string" + string(jsonRaw)
	goFile = strings.Replace(goFile, "\"\n}", "\",\n}", 1)
	content, _ := format.Source([]byte(goFile))

	err = ioutil.WriteFile(gofile, content, os.ModePerm)
	if err != nil {
		fmt.Println("保存程序文件出错", err)
	}
	os.Remove(_TMP_ICON_CACHE)
}

// https://stackoverflow.com/questions/51779243/copy-a-folder-in-go
func copyDirectory(scrDir, dest string) error {
	entries, err := ioutil.ReadDir(scrDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := createIfNotexists(destPath, 0755); err != nil {
				return err
			}
			if err := copyDirectory(sourcePath, destPath); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := copySymLink(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := copy(sourcePath, destPath); err != nil {
				return err
			}
		}

		if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
			return err
		}

		isSymlink := entry.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, entry.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

func copy(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}

	defer out.Close()

	in, err := os.Open(srcFile)
	defer in.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func createIfNotexists(dir string, perm os.FileMode) error {
	if exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

func copySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	rootPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	appFile, err := ioutil.ReadFile(filepath.Join(rootPath, "res/app.htm"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	appContent := string(appFile)

	// replace all script
	re := regexp.MustCompile(`\<script[^>]*?src=\"([\w\W]+?)\">\s*?\<\/script>`)
	matchs := re.FindAllStringSubmatch(appContent, -1)
	for _, match := range matchs {
		if len(match) > 1 {
			fileBytes, err := ioutil.ReadFile(filepath.Join(rootPath, match[1]))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			scriptContent := "<!-- " + match[1] + " -->\r\n <script type=\"text/tiscript\">\r\n" + string(fileBytes) + "\r\n</script>\r\n"
			appContent = strings.Replace(appContent, match[0], scriptContent, -1)
		}
	}

	// replace all css
	re = regexp.MustCompile(`@import\s+?url\(([^)]+?)\);`)
	matchs = re.FindAllStringSubmatch(appContent, -1)
	for _, match := range matchs {
		if len(match) > 1 {
			fileBytes, err := ioutil.ReadFile(filepath.Join(rootPath, match[1]))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			cssContent := "/* " + match[1] + " */\r\n" + string(fileBytes) + "\r\n"
			appContent = strings.Replace(appContent, match[0], cssContent, -1)
		}
	}

	// save to dist
	ioutil.WriteFile(filepath.Join(rootPath, "dist/app.htm"), []byte(appContent), os.ModePerm)
	fmt.Println("generate dist/app.htm success!!")

	binData := fmt.Sprintf("package dist\r\n\r\nvar DeployBinData string = `%s`\r\n", appContent)
	ioutil.WriteFile(filepath.Join(rootPath, "dist/app.go"), []byte(binData), os.ModePerm)
	fmt.Println("generate dist/dist.go success!!")

}

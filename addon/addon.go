package addon

import (
	"archive/zip"
	"ezserver/db"
	"ezserver/parser"
	"ezserver/utils"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func addToZip(zipWriter *zip.Writer, files map[string]string, baseDir string) {
	for name, content := range files {
		fullPath := baseDir + "/" + name
		f, err := zipWriter.Create(fullPath)
		utils.CheckErr(err)
		_, err = f.Write([]byte(content))
		utils.CheckErr(err)
	}
}

func GenerateAddon() {
	// генерируем файлы БД
	dbFiles := GenerateDB()
	totalFiles := len(dbFiles)

	// проверяем наличие директории для итогового архива
	outputDir := "./static/addon"
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err = os.MkdirAll(outputDir, 0755)
		utils.CheckErr(err)
	}

	zipPath := outputDir + "/IsengardArmory.zip"
	zipFile, err := os.Create(zipPath)
	utils.CheckErr(err)
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// добавляем все файлы внутрь папки IsengardArmory
	baseDir := "IsengardArmory"

	addToZip(zipWriter, dbFiles, baseDir)
	addToZip(zipWriter, map[string]string{
		"IsengardArmory.toc": GenerateTOC(totalFiles),
		"IsengardArmory.lua": GenerateMainLua(totalFiles),
	}, baseDir)

	log.Println("Аддон сгенерирован:", zipPath)
}

// joinWithComma объединяет строки через запятую
func joinWithComma(arr []string) string {
	return strings.Join(arr, ", ")
}

// escapeLuaString экранирует кавычки и слэши для Lua
func escapeLuaString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// GenerateTOC возвращает содержимое TOC для добавления в zip
func GenerateTOC(totalFiles int) string {
	tmpl := readTemplate("templates/template.toc")
	var dbFiles strings.Builder
	for i := 0; i < totalFiles; i++ {
		dbFiles.WriteString(fmt.Sprintf("DB%d.lua\n", i))
	}
	return strings.ReplaceAll(tmpl, "{{DB_FILES}}", dbFiles.String())
}

// GenerateMainLua генерирует основной Lua файл для аддона на основе шаблона
func GenerateMainLua(totalDB int) string {
	tmpl := readTemplate("templates/template.lua")
	var dbs strings.Builder
	for i := 0; i < totalDB; i++ {
		dbs.WriteString(fmt.Sprintf("DATABASE%d", i))
		if i < totalDB-1 {
			dbs.WriteString(", ")
		}
	}

	accountCount := db.CountUniqueLogins()
	characterCount := db.CountCharacters()

	tmpl = strings.ReplaceAll(tmpl, "{{DBS_PLACEHOLDER}}", dbs.String())
	tmpl = strings.ReplaceAll(tmpl, "{{TOTAL_DB}}", fmt.Sprintf("%d", totalDB))
	tmpl = strings.ReplaceAll(tmpl, "{{DATE}}", time.Now().Format("02.01.2006"))
	tmpl = strings.ReplaceAll(tmpl, "{{ACCOUNT_COUNT}}", fmt.Sprintf("%d", accountCount))
	tmpl = strings.ReplaceAll(tmpl, "{{CHARACTER_COUNT}}", fmt.Sprintf("%d", characterCount))
	return tmpl
}

// readTemplate читает шаблон из файла
func readTemplate(path string) string {
	data, err := os.ReadFile(path)
	utils.CheckErr(err)
	return string(data)
}

// GenerateDB извлекает данные из MongoDB и формирует Lua файлы
func GenerateDB() map[string]string {
	records := db.GetCharactersSorted()
	const chunkSize = 10000
	files := make(map[string]string)
	var sb strings.Builder
	fileIndex := 0
	lineCount := 0

	openNewFile := func(index int) {
		if sb.Len() > 0 {
			sb.WriteString("}\n")
			files[fmt.Sprintf("DB%d.lua", fileIndex)] = sb.String()
		}
		sb.Reset()
		sb.WriteString(fmt.Sprintf("DATABASE%d = {\n", index))
		lineCount = 0
		fileIndex = index
	}

	openNewFile(fileIndex)

	var currentLogin string
	var chars []parser.Character

	flushChars := func(login string, chars []parser.Character) {
		if len(chars) == 0 {
			return
		}
		charStrs := make([]string, len(chars))
		for i, ch := range chars {
			charStrs[i] = fmt.Sprintf(`{"%s", %d, %d, %d, "%s", %d}`,
				escapeLuaString(ch.Name), ch.LVL, ch.GS, ch.Race, escapeLuaString(ch.Guild), ch.Class)
		}
		line := fmt.Sprintf(`    ["%s"] = {%s},`, escapeLuaString(login), joinWithComma(charStrs))
		sb.WriteString(line + "\n")
		lineCount++
		if lineCount >= chunkSize {
			openNewFile(fileIndex + 1)
		}
	}

	for _, c := range records {
		if currentLogin == "" {
			currentLogin = c.Login
		}
		if c.Login != currentLogin {
			flushChars(currentLogin, chars)
			currentLogin = c.Login
			chars = chars[:0]
		}
		chars = append(chars, c)
	}
	flushChars(currentLogin, chars)

	if sb.Len() > 0 {
		sb.WriteString("}\n")
		files[fmt.Sprintf("DB%d.lua", fileIndex)] = sb.String()
	}

	return files
}

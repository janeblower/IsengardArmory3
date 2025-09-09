package addon

import (
	"archive/zip"
	"ezserver/db"
	"ezserver/parser"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// GenerateAddon собирает все файлы в zip, включая TOC и основной Lua
func GenerateAddon() {
	dbFiles := GenerateDB()
	totalFiles := len(dbFiles)

	zipFile, err := os.Create("./static/addon/IsengardArmory.zip")
	checkErr(err)
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	addFilesToZip(zipWriter, dbFiles)
	addFileToZip(zipWriter, "IsengardArmory.toc", GenerateTOC(totalFiles))
	addFileToZip(zipWriter, "IsengardArmory.lua", GenerateMainLua(totalFiles))

	log.Println("Аддон сгенерирован.")
}

// addFilesToZip добавляет несколько файлов в архив
func addFilesToZip(zipWriter *zip.Writer, files map[string]string) {
	for name, content := range files {
		addFileToZip(zipWriter, name, content)
	}
}

// addFileToZip добавляет один файл в архив
func addFileToZip(zipWriter *zip.Writer, name, content string) {
	f, err := zipWriter.Create(name)
	checkErr(err)
	_, err = f.Write([]byte(content))
	checkErr(err)
}

// checkErr завершает выполнение при ошибке
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
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
	checkErr(err)
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
				escapeLuaString(ch.Name), ch.LVL, ch.AP, ch.Class, escapeLuaString(ch.Guild), ch.Race)
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

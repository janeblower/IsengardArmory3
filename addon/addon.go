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

// Генерация аддона с использованием MongoStore
func GenerateAddon(store *db.MongoStore) {
	dbFiles := generateDB(store)
	totalFiles := len(dbFiles)

	outputDir := "./static/addon"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal("Ошибка создания директории:", err)
	}

	zipPath := outputDir + "/IsengardArmory.zip"
	zipFile, err := os.Create(zipPath)
	if err != nil {
		log.Fatal(err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	baseDir := "IsengardArmory"
	addToZip(zipWriter, dbFiles, baseDir)
	addToZip(zipWriter, map[string]string{
		"IsengardArmory.toc": generateTOC(totalFiles),
		"IsengardArmory.lua": generateMainLua(store, totalFiles),
	}, baseDir)

	log.Println("Аддон сгенерирован:", zipPath)
}

func addToZip(zipWriter *zip.Writer, files map[string]string, baseDir string) {
	for name, content := range files {
		fullPath := baseDir + "/" + name
		f, err := zipWriter.Create(fullPath)
		if err != nil {
			log.Fatal(err)
		}
		_, err = f.Write([]byte(content))
		if err != nil {
			log.Fatal(err)
		}
	}
}

// Генерация Lua DB файлов
func generateDB(store *db.MongoStore) map[string]string {
	records, err := store.GetCharactersSorted()
	if err != nil {
		log.Fatal("Ошибка получения персонажей:", err)
	}

	const chunkSize = 10000
	files := make(map[string]string)
	var sb strings.Builder
	fileIndex := 0
	lineCount := 0
	currentLogin := ""
	var chars []parser.Character

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

	flushChars := func(login string, chars []parser.Character) {
		if len(chars) == 0 {
			return
		}
		charStrs := make([]string, len(chars))
		for i, ch := range chars {
			charStrs[i] = fmt.Sprintf(`{"%s", %d, %d, %d, "%s", %d}`,
				escapeLuaString(ch.Name), ch.LVL, ch.GS, ch.Race, escapeLuaString(ch.Guild), ch.Class)
		}
		sb.WriteString(fmt.Sprintf(`    ["%s"] = {%s},`, escapeLuaString(login), strings.Join(charStrs, ", ")) + "\n")
		lineCount++
		if lineCount >= chunkSize {
			openNewFile(fileIndex + 1)
		}
	}

	openNewFile(fileIndex)

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

func generateTOC(totalFiles int) string {
	tmpl := readTemplate("templates/template.toc")
	var dbFiles strings.Builder
	for i := 0; i < totalFiles; i++ {
		dbFiles.WriteString(fmt.Sprintf("DB%d.lua\n", i))
	}
	return strings.ReplaceAll(tmpl, "{{DB_FILES}}", dbFiles.String())
}

func generateMainLua(store *db.MongoStore, totalDB int) string {
	tmpl := readTemplate("templates/template.lua")
	var dbs []string
	for i := 0; i < totalDB; i++ {
		dbs = append(dbs, fmt.Sprintf("DATABASE%d", i))
	}

	accountCount, _ := store.CountUniqueLogins()
	characterCount, _ := store.CountCharacters()

	tmpl = strings.ReplaceAll(tmpl, "{{DBS_PLACEHOLDER}}", strings.Join(dbs, ", "))
	tmpl = strings.ReplaceAll(tmpl, "{{TOTAL_DB}}", fmt.Sprintf("%d", totalDB))
	tmpl = strings.ReplaceAll(tmpl, "{{DATE}}", time.Now().Format("02.01.2006"))
	tmpl = strings.ReplaceAll(tmpl, "{{ACCOUNT_COUNT}}", fmt.Sprintf("%d", accountCount))
	tmpl = strings.ReplaceAll(tmpl, "{{CHARACTER_COUNT}}", fmt.Sprintf("%d", characterCount))
	return tmpl
}

func escapeLuaString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

func readTemplate(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return string(data)
}

package addon

import (
	"archive/zip"
	"context"
	"ezserver/db"
	"ezserver/parser"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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

	fmt.Println("Аддон сгенерирован.")
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
	coll := db.GetCollection("ezwow", "armory")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$login"}}}},
		{{Key: "$count", Value: "uniqueLogins"}},
	}
	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	var countUniqLogins []bson.M
	if err = cursor.All(ctx, &countUniqLogins); err != nil {
		log.Fatal(err)
	}

	countCharacters, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		panic(err)
	}

	tmpl = strings.ReplaceAll(tmpl, "{{DBS_PLACEHOLDER}}", dbs.String())
	tmpl = strings.ReplaceAll(tmpl, "{{TOTAL_DB}}", fmt.Sprintf("%d", totalDB))
	tmpl = strings.ReplaceAll(tmpl, "{{DATE}}", time.Now().Format("02.01.2006"))
	tmpl = strings.ReplaceAll(tmpl, "{{ACCOUNT_COUNT}}", fmt.Sprintf("%d", countUniqLogins[0]["uniqueLogins"]))
	tmpl = strings.ReplaceAll(tmpl, "{{CHARACTER_COUNT}}", fmt.Sprintf("%d", countCharacters))
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
	coll := db.GetCollection("ezwow", "armory")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	findOptions := options.Find().
		SetBatchSize(10000).
		SetSort(bson.D{{Key: "login", Value: 1}})

	cur, err := coll.Find(ctx, bson.M{}, findOptions)
	checkErr(err)
	defer cur.Close(ctx)

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

	for cur.Next(ctx) {
		var c parser.Character
		if err := cur.Decode(&c); err != nil {
			log.Println("decode error:", err)
			continue
		}

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

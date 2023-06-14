package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
)

// Информацию о пакетах можно брать из pkg.go.dev
// text/template для шаблонов
// flag для флагов командной строки
// Чтобы открыть файл или завершить программу с каким-то конкретным кодом выхода можно использовать пакет os
// Потом можно прочитать всё его содержимое с помощью пакета io
// json для превращения JSON'а в данные, с которыми можно работать в программе

// Первая итерация:
// 1. Программа принимает JSON в виде одного объекта через флаг `-json`
// 2. Программа принимает go шаблон через флаг `-template`
// 3. Программа выводит результат исполнения шаблона с контекстом этого JSON'а в стандартный вывод
// 4. В случае возникновения ошибок программа завершается с кодом 1
// 5. В случае возникновения ошибок программа оборачивает ошибки с помощью fmt.Errorf и печатает их в os.Stderr - стандартный выход для ошибок?

// Пример:
//
//	JSON: {"firstName": "Матвей", "lastName": "Вдовицын", "middleName": "Валентинович"}
//	Template: {{.lastName}} {{.firstName}} {{.middleName}}
//	Результат: Вдовицын Матвей Валентинович

// 1. [x] Все ошибки пишутся в os.Stdout, а не os.Stderr
// 2. [x] os.Exit(1) много повторяется для каждой ошибки, если забыть его добавить, то программа просто продолжит выполнение как ни в чем
//        не бывало
// 3. [ ] Мы не проверяем, что пользователь указал шаблон и данные для выполнения, давай будем возвращать ошибку, если такое происходит
// 4. [x] strings.builder чтобы запомнить всё, что записалось в шаблон и если нет ошибок, то вывести

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to execute template: %v", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		literalJson = flag.String("json", "", "Literal JSON value used for template execution")
		jsonFile    = flag.String("json-file", "", "Path to a file containing json")

		literalTemplate = flag.String("template", "", "Literal template value")
		templateFile    = flag.String("template-file", "", "Path to a file containing template")
	)
	flag.Parse()

	if *literalJson == "" && *jsonFile == "" {
		return errors.New("can't find json")
	}
	if *literalTemplate == "" && *templateFile == "" {
		return errors.New("can't find template")
	}
	if *literalJson != "" && *jsonFile != "" {
		return errors.New("you can't have two json")
	}
	if *literalTemplate != "" && *templateFile != "" {
		return errors.New("you can't have two templates")
	}

	var jsonData any

	if *literalJson != "" {
		err := json.Unmarshal([]byte(*literalJson), &jsonData)
		if err != nil {
			return err
		}

	} else {
		jsonFileData, err := os.Open(*jsonFile)
		if err != nil {
			return err
		}

		j, err := io.ReadAll(jsonFileData)
		if err != nil {
			return err
		}

		err = json.Unmarshal(j, &jsonData)
		if err != nil {
			return err
		}
	}

	templateData := template.New("templateData").Option("missingkey=error")

	if *literalTemplate != "" {
		_, err := templateData.Parse(*literalTemplate)
		if err != nil {
			return err
		}

	} else {
		templateFileData, err := os.Open(*templateFile)
		if err != nil {
			return err
		}

		t, err := io.ReadAll(templateFileData)
		if err != nil {
			return err
		}

		_, err = templateData.Parse(string(t))
		if err != nil {
			return err
		}
	}

	var res strings.Builder
	if err := templateData.Execute(&res, jsonData); err != nil {
		return err
	}

	fmt.Println(res.String())
	return nil

}

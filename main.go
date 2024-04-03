package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

func translateText(englishTexts <-chan string, translatedTexts chan<- string, wg *sync.WaitGroup, language string) {
	var text string
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")

	for input := range englishTexts {
		parts := strings.SplitN(input, "###", 2)
		position, _ := strconv.Atoi(parts[0])
		text = parts[1]
		time.Sleep(time.Duration(math.Floor(float64(position)/10)) * time.Second)
		client := openai.NewClient(openaiAPIKey)
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: openai.GPT4TurboPreview,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: "Translate the following text from english to " + language + ": " + text,
					},
				},
			},
		)

		if err != nil {
			time.Sleep(time.Duration(position*10) * time.Second)
			resp, err = client.CreateChatCompletion(
				context.Background(),
				openai.ChatCompletionRequest{
					Model: openai.GPT4TurboPreview,
					Messages: []openai.ChatCompletionMessage{
						{
							Role:    openai.ChatMessageRoleUser,
							Content: "Translate the following text from english to " + language + ": " + text,
						},
					},
				},
			)

			if err != nil {
				translatedTexts <- fmt.Sprintf("%d###%s", position, "{ERR| Invalid GPT response}") // move on, tried twice and failed
			} else {
				translatedTexts <- fmt.Sprintf("%d###%s", position, resp.Choices[0].Message.Content)
			}
		} else {
			translatedTexts <- fmt.Sprintf("%d###%s", position, resp.Choices[0].Message.Content)
		}

		wg.Done()
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <filepath> <language> <routines>\nfilepath = path of file to translate\nlanguage = language to translate to\nroutines = # of concurrent routines to run\n", os.Args[0])
		os.Exit(1)
	}

	// set filepath to first command line argument
	filePath := os.Args[1]
	language := os.Args[2]

	routineCount, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting routine count to integer: %v\n", err)
		os.Exit(1)
	}

	// Read the input text file.
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	var paragraphs []string
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var buffer bytes.Buffer
	counter := 0
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" { // Assuming an empty line represents a paragraph break.
			counter++
			if counter > 5 {
				paragraphs = append(paragraphs, buffer.String())
				buffer.Reset()
				counter = 0
			}
		} else {
			buffer.WriteString(line + "\n")
		}
	}
	// Add the last paragraph if exists.
	if buffer.Len() > 0 {
		paragraphs = append(paragraphs, buffer.String())
	}

	var wg sync.WaitGroup
	translatedTexts := make(chan string, len(paragraphs))
	englishTexts := make(chan string, 4)

	// start routineCount number of translate routines
	for i := 0; i < routineCount; i++ {
		go translateText(englishTexts, translatedTexts, &wg, language)
	}

	for i, paragraph := range paragraphs {
		if strings.TrimSpace(paragraph) == "" {
			continue
		}
		wg.Add(1)
		englishTexts <- fmt.Sprintf("%d###%s", i, paragraph)
	}

	wg.Wait()
	close(englishTexts)
	close(translatedTexts)

	// Collect translated texts and sort them based on their position.
	translatedParagraphs := make([]string, len(paragraphs))
	for t := range translatedTexts {
		parts := strings.SplitN(t, "###", 2)
		offset, _ := strconv.Atoi(parts[0])
		translatedParagraphs[offset] = parts[1]
	}

	// Reassemble the document.
	finalText := strings.Join(translatedParagraphs, "\n\n")
	fmt.Println(finalText)
}

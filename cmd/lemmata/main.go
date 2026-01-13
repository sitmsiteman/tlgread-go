package main

import (
        "flag"
        "bufio"
        "log"
        "fmt"
        "os"
        "strings"
)

type LemmaInfo struct {
        Lemma string
        ID    string
        Forms []string
}

// FindFormsByLemma searches greek-lemmata.txt for a headword
func FindFormsByLemma(filePath, targetLemma string) (*LemmaInfo, error) {
        file, err := os.Open(filePath)
        if err != nil {
                return nil, err
        }
        defer file.Close()

        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
                line := scanner.Text()
                // Format: lemma    ID    form1 (morph)    form2 (morph)...
                parts := strings.Split(line, "\t")
                if len(parts) < 3 {
                        continue
                }

                lemma := strings.TrimSpace(parts[0])
                if lemma == targetLemma {
                        return &LemmaInfo{
                                Lemma: lemma,
                                ID:    parts[1],
                                // The third part contains all forms and their morph data
                                Forms: strings.Split(parts[2], "\t"),
                        }, nil
                }
        }

        return nil, fmt.Errorf("lemma %s not found", targetLemma)
}

func main() {

        fPath := flag.String("f", "greek-lemmata.txt", "filename")
        word := flag.String("w", "", "word")
        flag.Parse()

        if *fPath == "" { log.Fatal("Usage: ./lemmata -f greek-lemmata.txt") }

        f, err := os.Open(*fPath)
        if err != nil { log.Fatal(err) }
        defer f.Close()

        filePath := *fPath

        // Example: Get all forms for the adverb "a(/dhn"
        info, err := FindFormsByLemma(filePath, *word)
        if err != nil {
                fmt.Println(err)
                return
        }

        fmt.Printf("Lemma: %s (ID: %s)\n", info.Lemma, info.ID)
        fmt.Println("Known inflections and variants:")
        for _, f := range info.Forms {
                if f != "" {
                        fmt.Printf(" - %s\n", strings.TrimSpace(f))
                }
        }
}


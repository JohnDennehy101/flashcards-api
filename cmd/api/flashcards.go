package main

import (
	"fmt"
	"net/http"
	"time"

	"flashcards-api.johndennehy101.tech/internal/data"
)

func (app *application) createFlashcardHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new flashcard")
}

func (app *application) showFlashcardHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	sectionTitle := "Order 40F - Intellectual Property"
	sectionType := "order"
	sourceFile := "Court Rules"
	fullText := `## **1 Definitions**

1. In this Order, unless the context or subject matter otherwise requires—

the “**1992 Act**” means the [Patents Act 1992](https://www.irishstatutebook.ie/eli/1992/act/1/enacted/en/html) (No. 1 of 1992);
the “**1996 Act**” means the [Trade Marks Act 1996](https://www.irishstatutebook.ie/eli/1996/act/6/enacted/en/html) (No. 6 of 1996);
the “**2000 Act**” means the [Copyright and Related Rights Act 2000](https://www.irishstatutebook.ie/eli/2000/act/28/enacted/en/html) (No. 28 of 2000);
the “**2001 Act**” means the [Industrial Designs Act 2001](https://www.irishstatutebook.ie/eli/2001/act/39/enacted/en/html) (No. 39 of 2001);
the “**2019 Act**” means the [Copyright and Other Intellectual Property Law Provisions Act 2019](https://www.irishstatutebook.ie/eli/2019/act/19/enacted/en/html) (No. 19 of 2019);
“**intellectual property claim**” has the same meaning as in section 2 of the [2000 Act](https://www.irishstatutebook.ie/eli/2000/act/28/enacted/en/html).`

	flashcard := data.Flashcard{
		ID:          id,
		CreatedAt:   time.Now(),
		Section:     &sectionTitle,
		SectionType: &sectionType,
		SourceFile:  &sourceFile,
		Text:        fullText,

		Question: "Is the Copyright and Related Rights Act 2000 mentioned in the definitions?",
		Type:     data.FlashcardYesNo,
		Content: data.YesNoContent{
			Correct:       true,
			Justification: "the “2000 Act” means the [Copyright and Related Rights Act 2000](https://www.irishstatutebook.ie/eli/2000/act/28/enacted/en/html) (No. 28 of 2000);",
		},
		Categories: []string{"intellectual property"},
		Version:    1,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"flashcard": flashcard}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

package model

import (
	"time"

	"github.com/google/uuid"
)

var levelToInterval = map[int]time.Duration{
	0: time.Hour,
	1: time.Hour * 24,
	2: time.Hour * 24 * 3,
	3: time.Hour * 24 * 7,
	4: time.Hour * 24 * 12,
	5: time.Hour * 24 * 21,
}

type PartOfSpeech string

const (
	PartOfSpeechUnknown     PartOfSpeech = "unknown"
	PartOfSpeechNoun        PartOfSpeech = "noun"
	PartOfSpeechVerb        PartOfSpeech = "verb"
	PortOfSpeechNumeral     PartOfSpeech = "numeral"
	PartOfSpeechAdjective   PartOfSpeech = "adjective"
	PartOfSpeechAdverb      PartOfSpeech = "adverb"
	PartOfSpeechPronoun     PartOfSpeech = "pronoun"
	PartOfSpeechPreposition PartOfSpeech = "preposition"
	PartOfSpeechConjunction PartOfSpeech = "conjunction"
	PartOfSpeechQuestion    PartOfSpeech = "question"
)

type NounForm string

const (
	NounFormUnknown            NounForm = "unknown"
	NounFormIndefiniteSingular NounForm = "indefinite_singular"
	NounFormDefiniteSingular   NounForm = "definite_singular"
	NounFormIndefinitePlural   NounForm = "indefinite_plural"
	NounFormDefinitePlural     NounForm = "definite_plural"
)

type VerbForm string

const (
	VerbFormUnknown VerbForm = "unknown"
	VerbFormPresent VerbForm = "present"
	VerbFormPast    VerbForm = "past"
)

type NumeralForm string

const (
	NumeralFormUnknown        NumeralForm = "unknown"
	NumeralFormCardinal       NumeralForm = "cardinal"
	NumeralFormOrdinal        NumeralForm = "ordinal"
	NumeralFormMultiplicative NumeralForm = "multiplicative"
	NumeralFormFractional     NumeralForm = "fractional"
)

type PronounForm string

const (
	PronounFormUnknown        PronounForm = "unknown"
	PronounFormPersonal       PronounForm = "personal"
	PronounFormPersonalObject PronounForm = "personal_object"
	PronounFormPossessive     PronounForm = "possessive"
)

type QuestionForm string

const (
	QuestionFormUnknown  QuestionForm = "unknown"
	QuestionFormQuestion QuestionForm = "question"
)

type AdverbForm string

const (
	AdverbFormUnknown  AdverbForm = "unknown"
	AdverbFormQuestion AdverbForm = "adverb"
)

type Vocab struct {
	Id           uuid.UUID    `json:"id"`
	Definition   string       `json:"definition"`
	PartOfSpeech PartOfSpeech `json:"part_of_speech"`
	Forms        []VocabForm  `json:"forms"`
	PausedUntil  *time.Time   `json:"pause_until,omitempty"`
}

func (v *Vocab) CanBeAddedToQueue(now time.Time) bool {
	return v.PausedUntil == nil
}

func (v *Vocab) UpdateFormSuccess(form VocabForm, now time.Time) {
	for i, vocabForm := range v.Forms {
		if vocabForm.Id == form.Id {
			vocabForm.SuccessInRow++
			vocabForm.LastSuccess = now
			if vocabForm.SuccessInRow >= 7 {
				vocabForm.SuccessInRow = 0
				vocabForm.Level++
				vocabForm.Level = min(vocabForm.Level, 5)
			}
			v.Forms[i] = vocabForm
			break
		}
	}
}

func (v *Vocab) UpdateFormError(form VocabForm) {
	for i, vocabForm := range v.Forms {
		if vocabForm.Id == form.Id {
			vocabForm.Level--
			vocabForm.Level = max(vocabForm.Level, 0)
			v.Forms[i] = vocabForm
			break
		}
	}
}

type VocabForm struct {
	Id           uuid.UUID `json:"id"`
	Value        string    `json:"value"`
	Form         string    `json:"form"`
	Level        int       `json:"level"`
	LastSuccess  time.Time `json:"last_success"`
	SuccessInRow int       `json:"success_in_row"`
}

func (v *VocabForm) CanBeAddedToQueue(t time.Time) bool {
	if v.Form == "definite_singular" || v.Form == "definite_plural" {
		return false
	}

	interval := time.Hour
	if i, ok := levelToInterval[v.Level]; ok {
		interval = i
	}

	return v.LastSuccess.Add(interval).Before(t)
}

type Pool struct {
	CreatedAt time.Time `json:"created_at"`
	Vocabs    []Vocab   `json:"vocabs"`
}

type Batch struct {
	Vocabs []Vocab `json:"vocabs"`
}

type VocabSet struct {
	Id        uuid.UUID   `json:"id"`
	Name      string      `json:"name"`
	VocabIds  []uuid.UUID `json:"vocab_ids"`
}

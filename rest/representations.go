package rest

import (
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/go-ard"
	"github.com/tliron/prudence/platform"
)

//
// Representations
//

type RepresentationEntry struct {
	Representation *Representation
	ContentType    ContentType
	Language       Language
}

type Representations struct {
	Entries []*RepresentationEntry
}

func CreateRepresentations(config ard.Value, jsContext *commonjs.Context) (*Representations, error) {
	var self Representations

	if err := platform.CreateFromConfigList(jsContext, config, "Representation", func(instance any, config_ ard.StringMap) {
		config__ := ard.With(config_).ConvertSimilar().NilMeansZero()
		contentTypes := platform.AsStringList(config__.Get("contentTypes"))
		languages := platform.AsStringList(config__.Get("languages"))
		self.Add(contentTypes, languages, instance.(*Representation))
	}); err != nil {
		return nil, err
	}

	return &self, nil
}

func (self *Representations) Add(contentTypes []string, languages []string, representation *Representation) {
	if len(contentTypes) == 0 {
		contentTypes = []string{""}
	}

	if len(languages) == 0 {
		languages = []string{""}
	}

	// The order signifies the *server* matching preferences
	for _, contentType := range contentTypes {
		contentType_ := NewContentType(contentType)
		for _, language := range languages {
			self.Entries = append(self.Entries, &RepresentationEntry{
				Representation: representation,
				ContentType:    contentType_,
				Language:       NewLanguage(language),
			})
		}
	}
}

func (self *Representations) NegotiateBest(restContext *Context) (*Representation, string, string, bool) {
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Content_negotiation

	contentTypePreferences := ParseContentTypePreferences(restContext.Request.Header.Get(HeaderAccept))
	languagePreferences := ParseLanguagePreferences(restContext.Request.Header.Get(HeaderAcceptLanguage))

	if len(languagePreferences) > 0 {
		// Try exact match of contentType and language
		for _, contentTypePreference := range contentTypePreferences {
			if contentTypePreference.Weight != 0.0 {
				for _, languagePreference := range languagePreferences {
					if languagePreference.Weight != 0.0 {
						for _, entry := range self.Entries {
							if contentTypePreference.Matches(entry.ContentType) && languagePreference.Matches(entry.Language, false) {
								return entry.Representation, entry.ContentType.Name, entry.Language.Name, true
							}
						}
					}
				}
			}
		}

		// Try exact match of contentType and soft match of language
		for _, contentTypePreference := range contentTypePreferences {
			if contentTypePreference.Weight != 0.0 {
				for _, languagePreference := range languagePreferences {
					if languagePreference.Weight != 0.0 {
						for _, entry := range self.Entries {
							if contentTypePreference.Matches(entry.ContentType) && languagePreference.Matches(entry.Language, true) {
								return entry.Representation, entry.ContentType.Name, entry.Language.Name, true
							}
						}
					}
				}
			}
		}
	}

	// Try exact match of contentType
	for _, contentTypePreference := range contentTypePreferences {
		if contentTypePreference.Weight != 0.0 {
			for _, entry := range self.Entries {
				if contentTypePreference.Matches(entry.ContentType) {
					return entry.Representation, entry.ContentType.Name, entry.Language.Name, true
				}
			}
		}
	}

	// TODO: for weight 0 should we expressly forbid matching any entry?
	// Probably not!

	// Try default representation (no contentType)
	for _, entry := range self.Entries {
		if entry.ContentType.Name == "" {
			return entry.Representation, "", "", true
		}
	}

	// Just pick the first one
	for _, entry := range self.Entries {
		return entry.Representation, entry.ContentType.Name, entry.Language.Name, true
	}

	return nil, "", "", false
}

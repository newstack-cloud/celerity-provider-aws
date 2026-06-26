package main

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// CloudFormation registry descriptions are written for CloudFormation and leak
// implementation detail we do not want in Bluelink docs: the ``AWS::Svc::Type``
// identifiers, CloudFormation doc links and macros, and CloudFormation framing
// ("templates", "stacks"). sanitiseDescription rewrites a raw schema description
// into clean provider prose. It returns "" when the source is pure boilerplate
// (e.g. "Resource Type definition for AWS::Events::Archive"), leaving the caller
// to substitute a synthesised fallback via fallbackDescription.

var (
	versionPrefixPattern  = regexp.MustCompile(`^Version: [^.]*\.\s*`)
	boilerplatePattern    = regexp.MustCompile(`(?i)^resource type definition for AWS::\S+?\.?$`)
	cfnTypePattern        = regexp.MustCompile("`*AWS::[A-Za-z0-9]+::[A-Za-z0-9]+`*")
	cfnLinkPattern        = regexp.MustCompile(`\[([^\]]*)\]\([^)]*AWSCloudFormation[^)]*\)`)
	doubleBacktickPattern = regexp.MustCompile("``+")
	multiSpacePattern     = regexp.MustCompile(`[ \t]{2,}`)
	// Only collapse a space before sentence punctuation that is itself followed by
	// whitespace or the end of the string, so legitimate leading-dot tokens such as
	// ".zip file" keep their preceding space.
	spaceBeforePunct = regexp.MustCompile(` +([.,;:])(\s|$)`)
)

// cfnIntrinsicPattern matches CloudFormation intrinsic-function references (the Fn::*
// family and the bare Ref function) along with the DependsOn attribute they co-occur
// with. Sentences carrying these describe CloudFormation template authoring, which has
// no Bluelink equivalent, so they are dropped wholesale rather than rewritten.
var cfnIntrinsicPattern = regexp.MustCompile("Fn::[A-Za-z0-9]+|`(?:Ref|DependsOn)`")

// orphanContinuationPrefixes flag a sentence as grammatically dependent on the one
// before it (a back-reference or discourse connective). When the sentence it depends
// on is removed as intrinsic guidance, the continuation is removed too so no dangling
// "This dependency..." fragment is left behind.
var orphanContinuationPrefixes = []string{
	"This ", "These ", "That ", "Those ", "It ",
	"For example", "For instance", "For more information",
}

// sentenceAbbreviations are period-terminated tokens that splitSentences must not treat
// as a sentence boundary.
var sentenceAbbreviations = []string{"e.g.", "i.e.", "etc.", "vs."}

// awsDocMacros expands the unexpanded AWS documentation shorthands that leak into
// registry descriptions. CloudFormation macros are mapped to "Bluelink" so the
// surrounding prose stays grammatical. Ordered longest-first at apply time so
// "CFNlong" is replaced before a bare "CFN", "DDBlong" before "DDB", etc.
var awsDocMacros = map[string]string{
	"CFNlong":  "Bluelink",
	"CFNshort": "Bluelink",
	"DDBlong":  "Amazon DynamoDB",
	"LAMlong":  "Lambda",
	"S3long":   "Amazon S3",
	"DDB":      "DynamoDB",
	"LAM":      "Lambda",
	"CFN":      "Bluelink",
}

// cfnPseudoParameters are CloudFormation pseudo-parameter identifiers (two-segment
// AWS::… tokens that the AWS::Svc::Type pattern does not match) that leak into
// CloudFormation intrinsic-function examples. They are rewritten to plain prose.
var cfnPseudoParameters = [][2]string{
	{"AWS::Region", "the region"},
	{"AWS::AccountId", "the account ID"},
	{"AWS::AccountID", "the account ID"},
	{"AWS::Partition", "the partition"},
	{"AWS::URLSuffix", "the URL suffix"},
	{"AWS::StackName", "the deployment name"},
	{"AWS::StackId", "the deployment ID"},
	{"AWS::NotificationARNs", "the notification ARNs"},
	{"AWS::NoValue", "no value"},
}

// cfnFramingTerms rebrands CloudFormation concepts onto their Bluelink
// equivalents so prose carried over from the registry reads as provider docs.
var cfnFramingTerms = [][2]string{
	{"AWS CloudFormation", "Bluelink"},
	{"CloudFormation", "Bluelink"},
	{"templates", "blueprints"},
	{"Templates", "Blueprints"},
	{"template", "blueprint"},
	{"Template", "Blueprint"},
	{"stacks", "deployments"},
	{"Stacks", "Deployments"},
	{"stack", "deployment"},
	{"Stack", "Deployment"},
	{"AWS-account", "AWS account"},
}

func sanitiseDescription(s string) string {
	s = strings.TrimSpace(s)
	s = versionPrefixPattern.ReplaceAllString(s, "")
	if boilerplatePattern.MatchString(s) {
		return ""
	}

	s = dropIntrinsicGuidance(s)
	s = cfnTypePattern.ReplaceAllStringFunc(s, friendlyTypePhrase)
	s = replacePseudoParameters(s)
	s = cfnLinkPattern.ReplaceAllString(s, "$1")
	s = replaceMacros(s)
	s = replaceFramingTerms(s)
	s = doubleBacktickPattern.ReplaceAllString(s, "`")

	s = multiSpacePattern.ReplaceAllString(s, " ")
	s = spaceBeforePunct.ReplaceAllString(s, "$1$2")
	return strings.TrimSpace(s)
}

func replacePseudoParameters(s string) string {
	for _, param := range cfnPseudoParameters {
		s = strings.ReplaceAll(s, param[0], param[1])
	}
	return s
}

func replaceMacros(s string) string {
	for _, macro := range macrosLongestFirst() {
		pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(macro) + `\b`)
		s = pattern.ReplaceAllString(s, awsDocMacros[macro])
	}
	return s
}

func macrosLongestFirst() []string {
	macros := make([]string, 0, len(awsDocMacros))
	for macro := range awsDocMacros {
		macros = append(macros, macro)
	}
	sortByLengthDesc(macros)
	return macros
}

func replaceFramingTerms(s string) string {
	for _, term := range cfnFramingTerms {
		pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(term[0]) + `\b`)
		s = pattern.ReplaceAllString(s, term[1])
	}
	return s
}

// dropIntrinsicGuidance removes whole sentences that give CloudFormation intrinsic-function
// guidance, preserving paragraph structure and the surrounding prose. A paragraph with no
// such sentence is returned byte-for-byte unchanged.
func dropIntrinsicGuidance(s string) string {
	if !cfnIntrinsicPattern.MatchString(s) {
		// No intrinsic guidance anywhere: leave the description byte-for-byte unchanged.
		return s
	}
	segments := strings.Split(s, "\n")
	kept := make([]string, 0, len(segments))
	for _, segment := range segments {
		body := strings.TrimSpace(segment)
		if body == "" {
			// Preserve blank-line paragraph separators verbatim.
			kept = append(kept, segment)
			continue
		}
		filtered := filterIntrinsicSentences(body)
		if filtered == "" {
			// The whole paragraph was intrinsic guidance.
			continue
		}
		if filtered == body {
			// Untouched paragraph: keep verbatim to preserve original whitespace.
			kept = append(kept, segment)
			continue
		}
		prefix := segment[:strings.IndexFunc(segment, isNotSpace)]
		kept = append(kept, prefix+filtered)
	}
	return strings.Join(kept, "\n")
}

func filterIntrinsicSentences(paragraph string) string {
	sentences := splitSentences(paragraph)
	kept := make([]string, 0, len(sentences))
	removedAny := false
	prevRemoved := false
	for _, sentence := range sentences {
		if cfnIntrinsicPattern.MatchString(sentence) ||
			(prevRemoved && isOrphanContinuation(sentence)) {
			removedAny = true
			prevRemoved = true
			continue
		}
		prevRemoved = false
		kept = append(kept, sentence)
	}
	if !removedAny {
		// Leave clean paragraphs untouched so regeneration produces no incidental churn.
		return paragraph
	}
	return strings.Join(kept, " ")
}

func isOrphanContinuation(sentence string) bool {
	for _, prefix := range orphanContinuationPrefixes {
		if strings.HasPrefix(sentence, prefix) {
			return true
		}
	}
	return false
}

// splitSentences breaks a paragraph into sentences on terminal punctuation followed by
// whitespace and a capitalised (or quoted/backticked) start, leaving abbreviations such
// as "e.g." intact.
func splitSentences(paragraph string) []string {
	runes := []rune(paragraph)
	var sentences []string
	start := 0
	for i := 0; i < len(runes); i++ {
		if !isSentenceTerminator(runes[i]) {
			continue
		}
		resume := i + 1
		if resume >= len(runes) || !isSpace(runes[resume]) {
			continue
		}
		for resume < len(runes) && isSpace(runes[resume]) {
			resume++
		}
		if resume >= len(runes) || !startsNewSentence(runes[resume]) {
			continue
		}
		if endsWithAbbreviation(string(runes[start : i+1])) {
			continue
		}
		sentences = append(sentences, strings.TrimSpace(string(runes[start:i+1])))
		start = resume
		i = resume - 1
	}
	if tail := strings.TrimSpace(string(runes[start:])); tail != "" {
		sentences = append(sentences, tail)
	}
	return sentences
}

func isNotSpace(r rune) bool           { return r != ' ' && r != '\t' }
func isSpace(r rune) bool              { return r == ' ' || r == '\t' }
func isSentenceTerminator(r rune) bool { return r == '.' || r == '!' || r == '?' }

func startsNewSentence(r rune) bool {
	return unicode.IsUpper(r) || r == '`' || r == '"'
}

func endsWithAbbreviation(segment string) bool {
	lower := strings.ToLower(strings.TrimSpace(segment))
	for _, abbrev := range sentenceAbbreviations {
		if strings.HasSuffix(lower, abbrev) {
			return true
		}
	}
	return false
}

// friendlyTypePhrase turns an AWS::Svc::Type identifier (with any surrounding
// backticks) into readable inline prose: "AWS::DynamoDB::Table" -> "DynamoDB
// table", "AWS::Events::EventBusPolicy" -> "Events event bus policy".
func friendlyTypePhrase(token string) string {
	core := strings.Trim(token, "`")
	parts := strings.Split(core, "::")
	if len(parts) != 3 {
		return token
	}
	service := parts[1]
	typeWords := strings.ToLower(labelFromName(parts[2]))
	return service + " " + typeWords
}

// fallbackDescription synthesises a resource description for a CFN type whose
// source description was pure boilerplate, e.g. "AWS::Events::Archive" ->
// "Manages an Events archive.".
func fallbackDescription(cfnType string) string {
	phrase := friendlyTypePhrase(cfnType)
	if phrase == cfnType {
		return ""
	}
	return fmt.Sprintf("Manages %s %s.", indefiniteArticle(phrase), phrase)
}

func indefiniteArticle(phrase string) string {
	if phrase == "" {
		return "a"
	}
	switch phrase[0] {
	case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
		return "an"
	default:
		return "a"
	}
}

func sortByLengthDesc(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && len(values[j]) > len(values[j-1]); j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}

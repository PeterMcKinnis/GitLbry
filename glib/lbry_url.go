package glib

import (
	"errors"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

type lbryUrl string


// Creates a new lbry url from the given string.  If 
// not start with lbry:// it is assmed.
func NewLbryUrl(x string) (lbryUrl, error) {
	x = normalize(x);

	if !strings.HasPrefix(x, "lbry://") {
		x = "lbry://" + x;
	}

	y := lbryUrl(x);
	if !y.IsValid() {
		return zero[lbryUrl](), errors.New("invalid lbry url")
	}
	return y, nil;
}


func normalize(lbryName string) string {
	return strings.ToLower(norm.NFD.String(lbryName))
}


// Returns true if the stream or channel has a 
// ClaimId modifier, Sequence Modifier, or AmountOrder Modifier
// ````
// lbry://@channel_name:23452/stream_name   true
// lbry://@channel_name/stream_name         false
// lbry://@ChAnNeL_NaMe/stream_name$2       true
// lbry://stream_name                       true
// ```
func HasModifiers(x string) bool {
	return strings.IndexAny(x, ":*$") != -1;
}


// For reference, this is regex that mostly validates urls
//lbry:\/\/(@[^=&#:*$@%?/]+(?:[:*$]\d+))?/([^=&#:*$@%?/]+(?:[:*$]\d+))?

// Returns the normalize channel name omitting any modifiers
// ````
// lbry://@channel_name:23452/stream_name   "channel_name", true
// lbry://@channel_name/stream_name         "channel_name", true
// lbry://@ChAnNeL_NaMe/stream_name         "channel_name", true
// lbry://stream_name                       "", flase
// ```
func (url lbryUrl) ChannelName() (string, bool) {
	
	name := string(url);

	// Check for no channel
	if !strings.HasPrefix(name, "lbry://@") {
		return "", false;
	}

	// Strip scheme and @
	name = name[8:];

	// Modifier, Stream, And Query
	end := strings.IndexAny(name, ":*$/?")
	if end != -1 {
		name = name[:end]
	}

	return name, len(name) > 0;
}




// Returns the normalize channel name and modifiers
// ````
// lbry://@channel_name:23452/stream_name   "channel_name:23452", true
// lbry://@channel_name$2/stream_name       "channel_name$2", true
// lbry://@ChAnNeL_NaMe/stream_name         "channel_name", true
// lbry://stream_name                       "", flase
// ```
func (url lbryUrl) ChannelWithModifiers() (string, bool) {
	
	name := string(url);

	// Check for no channel
	if !strings.HasPrefix(name, "lbry://@") {
		return "", false;
	}

	// Strip scheme and @
	name = name[8:];

	// Modifier, Stream, And Query
	end := strings.IndexAny(name, "/?")
	if end != -1 {
		name = name[:end]
	}

	return name, len(name) > 0;
}

// Returns the normalized stream name and modifiers
func (url lbryUrl) StreamWithModifiers() (string, bool) {
	
	x := string(url);

	// Remove scheme
	x = x[7:];

	// Remove channel
	if strings.HasPrefix(x, "@") {
		idx := strings.Index(x, "/");
		if idx == -1 {
			// url is for a channel
			return "", false
		}

		x = x[idx:]
	}

	// Remove Query
	idx := strings.Index(x, "?");
	if idx != -1 {
		x = x[:idx];
	}
	return x, true;
}

// Returns the normalize channel name omitting any modifiers
// ````
// lbry://@channel_name:23452/stream_name   "stream_name", true
// lbry://@channel_name/stream_name:2342    "stream_name", true
// lbry://StReAM_NaMe                       "stream_name", true
// lbry://@channel_name                     "", flase
// ```
func (url lbryUrl) StreamName() (string, bool) {

	name := string(url);

	// Strip Scheme
	name = name[7:]

	// Strip Channel
	if strings.HasPrefix(name, "@") {
		end := strings.IndexAny(name, "/")
		if end == -1 {
			// No stream, this is a channel url
			return "", false
		}
		name = name[end+1:]
	}

	// Strip Modifier and query
	end := strings.IndexAny(name, ":*$?")
	if end != -1 {
		name = name[:end]
	}

	return name, len(name) > 0;
}

func (url lbryUrl) IsValid() bool {
	x, ok := parseUrl(string(url));
	return ok && len(x) == 0;
}

func  IsUrlValid(x string) bool {
	x = normalize(x);
	x, ok := parseUrl(x);
	return ok && len(x) == 0;
}



func parseUrl(x string) (string, bool) {

	var ok bool

	x, ok = parseScheme(x);
	if !ok {
		return "", false;
	}

	x, ok = parsePath(x);
	if !ok {
		return "", false;
	}

	x, ok = parseOptionalQuery(x);
	if !ok {
		return "", false
	}

	return x, ok;

}

func parseScheme(x string) (string, bool) {
	if !strings.HasPrefix(x, "lbry://") {
		return "", false
	}
	return x[7:], true
}

func parsePath(x string) (string, bool) {
	var ok bool

	if strings.HasPrefix(x, "@") {

		// Channel Claim
		x = x[1:]
		x, ok = parseClaimAndModifier(x);
		if !ok {
			return "", false;
		}

		// Optional stream claim
		if strings.HasPrefix(x, "/") {
			x = x[1:]
			return parseClaimAndModifier(x);
		} 
			
		return x, ok;

	} else {
		return parseClaimAndModifier(x);
	}
}

func parseClaimAndModifier(x string) (string, bool) {

	var ok bool


	x, ok = parseName(x)
	if !ok {
		return "", false
	}

	return parseOptionalModifier(x) 

}

func parseOptionalModifier(x string) (string, bool) {
	if strings.HasPrefix(x, ":") {
		x = x[1:]
		return parseHex(x);
	} else if strings.HasPrefix(x, "*") || strings.HasPrefix(x, "$") {
		x = x[1:]
		return parsePositiveNumber(x)
	}
	return x, true;
}

func parsePositiveNumber(x string) (string, bool) {
	ok := false;
	for {

		if (len(x) == 0) {
			break;
		}

		r, size := utf8.DecodeRuneInString(x)
		if r == utf8.RuneError {
			return "", false
		}

		if !isDigitChar(r) {
			break;
		}

		ok = true;
		x = x[size:]

	}

	return x, ok
}

func parseHex(x string) (string, bool) {
	ok := false;
	for {

		if (len(x) == 0) {
			break;
		}

		r, size := utf8.DecodeRuneInString(x)
		if r == utf8.RuneError {
			return "", false
		}

		if !isHexChar(r) {
			break;
		}

		ok = true;
		x = x[size:]

	}

	return x, ok
}

func parseName(x string) (string, bool) {

	ok := false;
	for {

		if (len(x) == 0) {
			break;
		}

		r, size := utf8.DecodeRuneInString(x)
		if r == utf8.RuneError {
			return "", false
		}

		if !isNameChar(r) {
			break;
		}

		ok = true;
		x = x[size:]

	}

	return x, ok
}


func isNameChar(x rune) bool {
	
	if strings.ContainsAny(string(x), "=&#:*$@%?/") {
		return false;
	}
	
	return x == 0x9 ||
	 x== 0xA ||
	 x == 0xD ||
	 (x >= 0x20 && x <= 0xD7FF) || 
	 (x >= 0xE000 && x <= 0xFFFD) ||
	 (x >= 0x10000 && x <= 0x10FFFF)

}

func isHexChar(x rune) bool {
	return (x >= 0x61 && x <= 0x66) || (x >= 0x30 && x <= 0x39)
}


func isDigitChar(x rune) bool {
	return x >= 0x30 && x <= 0x39
}

func parseOptionalQuery(x string) (string, bool) {
	var ok bool;
	if !strings.HasPrefix(x, "?") {
		return x, true
	}
	x = x[1:]

	x, ok = parseQueryParameter(x);
	if !ok {
		return "", false
	}

	for strings.HasPrefix(x, "&") {
		x, ok = parseQueryParameter(x);
		if !ok {
			return "", false
		}
	}

	return x, true;

}

func parseQueryParameter(x string) (string, bool) {
	var ok bool;

	x, ok = parseName(x)
	if !ok {
		return "", false
	}

	if strings.HasPrefix(x, "=") {
		return parseName(x)
	}

	return x, ok;

}
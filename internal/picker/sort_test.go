package picker

import (
	"cmp"
	"slices"
	"testing"
)

// sortByBucket mirrors the sort in picker.Run() so tests exercise the same
// comparison the live code uses.
func sortByBucket(emojis []Emoji, frequencies map[string]int64) {
	slices.SortStableFunc(emojis, func(a, b Emoji) int {
		return cmp.Compare(
			frequencyBucket(frequencies[b.Char]),
			frequencyBucket(frequencies[a.Char]),
		)
	})
}

func chars(emojis []Emoji) []string {
	out := make([]string, len(emojis))
	for i, e := range emojis {
		out[i] = e.Char
	}
	return out
}

func TestSortBucketStableWithinBucket(t *testing.T) {
	// Input order represents emoji.txt order (the stable tiebreaker we want
	// to preserve). 🎉 has the lower count but appears first; 🚀 has the
	// higher count but appears second. Both sit in the [10..99] bucket, so
	// the input order must win.
	input := []Emoji{
		{Char: "🎉"},
		{Char: "🚀"},
	}
	freq := map[string]int64{
		"🎉": 12,
		"🚀": 80,
	}
	sortByBucket(input, freq)

	want := []string{"🎉", "🚀"}
	if got := chars(input); !slices.Equal(got, want) {
		t.Errorf("within-bucket order changed: got %v, want %v (emoji.txt order)", got, want)
	}
}

func TestSortBucketHigherBucketWins(t *testing.T) {
	// 😀 sits in the [10..99] bucket; 🥳 sits in the [100..999] bucket.
	// Despite 😀 appearing first in emoji.txt order, 🥳 must come out on top.
	input := []Emoji{
		{Char: "😀"},
		{Char: "🥳"},
	}
	freq := map[string]int64{
		"😀": 50,
		"🥳": 150,
	}
	sortByBucket(input, freq)

	want := []string{"🥳", "😀"}
	if got := chars(input); !slices.Equal(got, want) {
		t.Errorf("got %v, want %v (higher bucket should sort first)", got, want)
	}
}

func TestSortBucketBoundary(t *testing.T) {
	// At count=9, 🎉 is in bucket 1 and stays behind 🚀 (bucket 1 also,
	// but 🚀 is listed first in input). Bumping 🎉 from 9 -> 10 promotes
	// it into bucket 2, which moves it ahead of 🚀.
	input := []Emoji{
		{Char: "🚀"},
		{Char: "🎉"},
	}
	freq := map[string]int64{
		"🚀": 5,
		"🎉": 9,
	}

	sorted := slices.Clone(input)
	sortByBucket(sorted, freq)
	if got, want := chars(sorted), []string{"🚀", "🎉"}; !slices.Equal(got, want) {
		t.Errorf("at counts {🚀:5, 🎉:9} got %v, want %v (same bucket -> input order)", got, want)
	}

	freq["🎉"] = 10
	sorted = slices.Clone(input)
	sortByBucket(sorted, freq)
	if got, want := chars(sorted), []string{"🎉", "🚀"}; !slices.Equal(got, want) {
		t.Errorf("after 🎉 9->10 got %v, want %v (🎉 jumped to bucket 2)", got, want)
	}
}

func TestSortBucketCasualUseDoesNotReorder(t *testing.T) {
	// The user's complaint: a few extra uses must not perturb the order.
	// Start with three emojis in the [10..99] bucket; bump each by a few
	// uses and assert the order is unchanged.
	input := []Emoji{
		{Char: "😀"},
		{Char: "🎉"},
		{Char: "🚀"},
	}
	freq := map[string]int64{
		"😀": 20,
		"🎉": 40,
		"🚀": 80,
	}

	sorted := slices.Clone(input)
	sortByBucket(sorted, freq)
	before := chars(sorted)

	freq["😀"] += 3
	freq["🎉"] += 5
	freq["🚀"] += 1

	sorted = slices.Clone(input)
	sortByBucket(sorted, freq)
	after := chars(sorted)

	if !slices.Equal(before, after) {
		t.Errorf("casual use reordered list: before=%v after=%v", before, after)
	}
}

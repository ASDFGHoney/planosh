package diff

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const namingSimilarityThreshold = 0.5

// classify takes raw divergences and returns classified Divergence results.
func classify(divs []rawDivergence, totalRuns int, runDirs []string, fileSets []map[string]struct{}) []Divergence {
	var contentDivs, missingDivs []rawDivergence
	for _, d := range divs {
		if d.allRuns {
			contentDivs = append(contentDivs, d)
		} else {
			missingDivs = append(missingDivs, d)
		}
	}

	var result []Divergence
	for _, d := range contentDivs {
		result = append(result, classifyContent(d))
	}
	result = append(result, classifyMissing(missingDivs, totalRuns, runDirs, fileSets)...)
	return result
}

// classifyContent handles files present in all runs with differing content.
func classifyContent(d rawDivergence) Divergence {
	groups := groupByContent(d.contents)

	var parts []string
	for _, g := range groups {
		labels := make([]string, len(g))
		for j, r := range g {
			labels[j] = fmt.Sprintf("run-%d", r+1)
		}
		parts = append(parts, strings.Join(labels, ", "))
	}

	return Divergence{
		Kind:   KindCodePattern,
		Path:   d.path,
		Runs:   allIndices(len(d.contents)),
		Detail: fmt.Sprintf("%d개 변형: [%s]", len(groups), strings.Join(parts, "] vs [")),
	}
}

// classifyMissing handles files not present in all runs.
// Detection order: naming → structure → scope excess.
func classifyMissing(divs []rawDivergence, totalRuns int, runDirs []string, fileSets []map[string]struct{}) []Divergence {
	if len(divs) == 0 {
		return nil
	}

	var result []Divergence
	consumed := make(map[int]bool)

	// Group by parent directory for naming detection.
	dirGroups := make(map[string][]int) // dir → indices into divs
	for i, d := range divs {
		dirGroups[filepath.Dir(d.path)] = append(dirGroups[filepath.Dir(d.path)], i)
	}

	// Naming: same directory, non-overlapping runs, similar content.
	for _, indices := range dirGroups {
		if len(indices) < 2 {
			continue
		}
		for a := 0; a < len(indices); a++ {
			ia := indices[a]
			if consumed[ia] {
				continue
			}
			for b := a + 1; b < len(indices); b++ {
				ib := indices[b]
				if consumed[ib] {
					continue
				}
				da, db := divs[ia], divs[ib]
				if runsOverlap(da.present, db.present) {
					continue
				}
				ca := readFromFirstRun(runDirs, da.path, da.present)
				cb := readFromFirstRun(runDirs, db.path, db.present)
				if ca != nil && cb != nil && lineSimilarity(ca, cb) >= namingSimilarityThreshold {
					consumed[ia] = true
					consumed[ib] = true
					result = append(result, Divergence{
						Kind:   KindNaming,
						Path:   filepath.Dir(da.path),
						Runs:   mergeRuns(da.present, db.present),
						Detail: fmt.Sprintf("같은 위치, 다른 이름: %s vs %s", filepath.Base(da.path), filepath.Base(db.path)),
					})
				}
			}
		}
	}

	// Remaining: structure or scope excess.
	for i, d := range divs {
		if consumed[i] {
			continue
		}
		absent := absentRuns(d.present, totalRuns)
		dir := filepath.Dir(d.path)

		if dir != "." && !dirExistsInAllRuns(dir, absent, fileSets) {
			result = append(result, Divergence{
				Kind:   KindStructure,
				Path:   d.path,
				Runs:   d.present,
				Detail: fmt.Sprintf("디렉토리 구조 차이: %s", dir),
			})
		} else {
			result = append(result, Divergence{
				Kind:   KindScopeExcess,
				Path:   d.path,
				Runs:   d.present,
				Detail: "일부 run에서만 존재",
			})
		}
	}

	return result
}

// groupByContent groups run indices by normalized content.
func groupByContent(contents [][]byte) [][]int {
	type entry struct {
		norm string
		runs []int
	}
	var groups []entry
	for i, c := range contents {
		n := NormalizeWhitespace(c)
		found := false
		for j := range groups {
			if groups[j].norm == n {
				groups[j].runs = append(groups[j].runs, i)
				found = true
				break
			}
		}
		if !found {
			groups = append(groups, entry{norm: n, runs: []int{i}})
		}
	}
	result := make([][]int, len(groups))
	for i, g := range groups {
		result[i] = g.runs
	}
	return result
}

// lineSimilarity returns Jaccard similarity of non-empty, normalized line sets.
func lineSimilarity(a, b []byte) float64 {
	sa := toLineSet(NormalizeWhitespace(a))
	sb := toLineSet(NormalizeWhitespace(b))
	if len(sa) == 0 && len(sb) == 0 {
		return 1.0
	}
	inter := 0
	for line := range sa {
		if _, ok := sb[line]; ok {
			inter++
		}
	}
	union := len(sa)
	for line := range sb {
		if _, ok := sa[line]; !ok {
			union++
		}
	}
	if union == 0 {
		return 1.0
	}
	return float64(inter) / float64(union)
}

func toLineSet(s string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, line := range strings.Split(s, "\n") {
		t := strings.TrimSpace(line)
		if t != "" {
			set[t] = struct{}{}
		}
	}
	return set
}

func readFromFirstRun(runDirs []string, path string, presentIn []int) []byte {
	if len(presentIn) == 0 {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(runDirs[presentIn[0]], path))
	if err != nil {
		return nil
	}
	return data
}

// dirExistsInAllRuns checks whether a directory has files in every specified run.
func dirExistsInAllRuns(dir string, runIndices []int, fileSets []map[string]struct{}) bool {
	prefix := dir + "/"
	for _, r := range runIndices {
		found := false
		for f := range fileSets[r] {
			if strings.HasPrefix(f, prefix) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func runsOverlap(a, b []int) bool {
	set := make(map[int]bool, len(a))
	for _, r := range a {
		set[r] = true
	}
	for _, r := range b {
		if set[r] {
			return true
		}
	}
	return false
}

func mergeRuns(a, b []int) []int {
	result := make([]int, 0, len(a)+len(b))
	result = append(result, a...)
	result = append(result, b...)
	sort.Ints(result)
	return result
}

func absentRuns(present []int, total int) []int {
	set := make(map[int]bool, len(present))
	for _, r := range present {
		set[r] = true
	}
	var absent []int
	for i := 0; i < total; i++ {
		if !set[i] {
			absent = append(absent, i)
		}
	}
	return absent
}

func allIndices(n int) []int {
	result := make([]int, n)
	for i := range result {
		result[i] = i
	}
	return result
}

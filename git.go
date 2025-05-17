package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

const (
	MainBranch             = "master"
	VersionYamlFile string = "version.yaml"
	Major           string = "major"
	Minor           string = "minor"
	Patch           string = "patch"
) // or "master", depending on your repo

func runGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

func fetchTags() error {
	// Step 1: Ensure we're on the MAIN_BRANCH
	branch, err := runGitCommand("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	if branch != MainBranch {
		return fmt.Errorf("error: you must be on the %s branch. Current branch is '%s'", MainBranch, branch)
	}

	// Step 2: Fetch from remote
	_, err = runGitCommand("fetch", "origin", MainBranch)
	if err != nil {
		return fmt.Errorf("git fetch failed: %w", err)
	}

	// Step 3: Compare local and remote HEADs
	local, err := runGitCommand("rev-parse", "@")
	if err != nil {
		return fmt.Errorf("failed to get local HEAD: %w", err)
	}
	remote, err := runGitCommand("rev-parse", "origin/"+MainBranch)
	if err != nil {
		return fmt.Errorf("failed to get remote HEAD: %w", err)
	}

	if local != remote {
		return fmt.Errorf("error: your local %s branch is not up-to-date with remote. Please pull the latest changes", MainBranch)
	}

	// Step 4: Fetch all tags
	_, err = runGitCommand("fetch", "--tags")
	if err != nil {
		return fmt.Errorf("failed to fetch tags: %w", err)
	}

	fmt.Println("Successfully fetched all tags.")
	return nil
}

// getLatestSemverTag fetches all tags and returns the latest semver tag (vX.Y.Z).
func getLatestSemverTag() (string, error) {
	// Step 1: Fetch tags
	if _, err := runGitCommand("fetch", "--tags"); err != nil {
		return "", fmt.Errorf("failed to fetch tags: %w", err)
	}

	// Step 2: List all tags
	output, err := runGitCommand("tag", "-l", "v[0-9]*.[0-9]*.[0-9]*")
	if err != nil {
		return "", fmt.Errorf("failed to list tags: %w", err)
	}
	tags := strings.Split(output, "\n")

	// Step 3: Filter valid semver tags and sort them
	semverRegex := regexp.MustCompile(`^v\d+\.\d+\.\d+$`)
	validTags := []string{}
	for _, tag := range tags {
		if semverRegex.MatchString(tag) {
			validTags = append(validTags, tag)
		}
	}

	if len(validTags) == 0 {
		return "", fmt.Errorf("no semver tags found")
	}

	// Step 4: Sort using Go's sort and a version-aware comparison
	sort.Slice(validTags, func(i, j int) bool {
		return compareSemver(validTags[i], validTags[j]) < 0
	})

	// Return the last (latest) tag
	return validTags[len(validTags)-1], nil
}

// compareSemver compares two semantic version strings like "v1.2.3"
// Returns -1 if a < b, 0 if a == b, 1 if a > b
func compareSemver(a, b string) int {
	parse := func(tag string) (major, minor, patch int) {
		fmt.Sscanf(tag, "v%d.%d.%d", &major, &minor, &patch)
		return
	}
	ma, mi, pa := parse(a)
	mb, mi2, pb := parse(b)

	if ma != mb {
		if ma < mb {
			return -1
		}
		return 1
	}
	if mi != mi2 {
		if mi < mi2 {
			return -1
		}
		return 1
	}
	if pa != pb {
		if pa < pb {
			return -1
		}
		return 1
	}
	return 0
}

func BumpVersionAndCreateTag(currentVersion string, bumpType string)

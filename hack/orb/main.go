/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bufio"
	"container/heap"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	_ "knative.dev/pkg/system/testing"
	"sigs.k8s.io/karpenter/pkg/controllers/orb"
	pb "sigs.k8s.io/karpenter/pkg/controllers/orb/proto"
)

var (
	logPath               string // Points to where the logs are stored (whether from the user's PV or some local save of the files)
	nodepoolsYamlFilepath string
	// TODO: Mount PV for access locally, if desired. (Or could leave that to a given customer?)
)

// Options of all the scheduling actions for the user to choose off the command-line.
// The scheduling inputs from the associated action is reconstructed based on the timestamp reference.
type SchedulingMetadataOption struct {
	ID        int
	Action    string
	Timestamp time.Time
}

func (o *SchedulingMetadataOption) String() string {
	return fmt.Sprintf("%d. %s (%s)", o.ID, o.Action, o.Timestamp.Format("2006-01-02_15-04-05"))
}

// Parse the command line arguments
func init() {
	flag.StringVar(&logPath, "dir", "", "Path to the directory containing logs")
	flag.StringVar(&nodepoolsYamlFilepath, "yaml", "", "Path to the YAML file containing NodePool definitions")
	flag.Parse()
}

// This conducts ORB Reconstruction from the command-line.
func main() {
	options, err := readMetadataLogs()
	if err != nil {
		fmt.Println("Error reading metadata logs:", err)
		return
	}
	selectedOption := promptUserForOption(options)
	fmt.Printf("Pulling option: %s from this directory: %s\n", selectedOption, logPath) // Delete later
	reconstructedSchedulingInput, err := reconstructFromOption(selectedOption.Timestamp)
	if err != nil {
		fmt.Println("Error reconstructing scheduling input:", err)
		return
	}
	writeReconstructionToJSON(reconstructedSchedulingInput)

	/* (Above) Original Scope: Log trace information important to Provisioning and Disruption as a start for troubleshooting */
	/* ----------------------- */
	/* (Below) Stretch Goal:   Have the tool resimulate and compare to the original results */

	// results, err := pkg.Resimulate(reconstructedSchedulingInput, nodepoolsYamlFilepath)
	// if err != nil {
	// 	fmt.Println("Error resimulating:", err)
	// 	return
	// }
	// fmt.Println("Results are:", results)

	// TODO: Compare these results with the original results' set of nodeclaims / nodeclaim returned from EC2Fleet...
}

// Read all metadata log files and extract available options
func readMetadataLogs() ([]*SchedulingMetadataOption, error) {
	// Get all filenames starting with "SchedulingMetadata" within the dirPath provided
	files, err := os.ReadDir(logPath)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return nil, err
	}
	// Filter out files that don't start with "SchedulingMetadata"
	var metadataFiles []string
	regex := regexp.MustCompile(`^SchedulingMetadata.*\.log$`)
	// Iterate through the files and add matching filenames to the metadataFiles slice
	for _, file := range files {
		if regex.MatchString(file.Name()) {
			metadataFiles = append(metadataFiles, file.Name())
		}
	}

	// For each metadataFile, read its contents, deserialize from protobuf and save into it's orb.Metadata structure
	// Then, extract the options from the Metadata structure and return them as a slice of strings
	options := []*SchedulingMetadataOption{}
	allMetadata := orb.NewSchedulingMetadataHeap()
	for _, metadataFile := range metadataFiles {
		contents, err := ReadLog(metadataFile)
		if err != nil {
			fmt.Println("Error reading file contents:", err)
			return nil, err
		}

		// Unmarshal the metadataLogdata back into the metadatamap
		protoMetadataMap := &pb.SchedulingMetadataMap{}
		proto.Unmarshal(contents, protoMetadataMap)
		metadataHeap := orb.ReconstructSchedulingMetadataHeap(protoMetadataMap)

		// While metadataHeap isn't empty, pop off each metadata, and save the Action and timestamp to a string option line with a newline at the end
		// Then, append that option to the options slice
		for metadataHeap.Len() > 0 {
			// Recombine/order all metadata from multiple files
			heap.Push(allMetadata, heap.Pop(metadataHeap).(orb.SchedulingMetadata))
		}
	}

	for i, metadata := range *allMetadata {
		options = append(options, &SchedulingMetadataOption{ID: i, Action: metadata.Action, Timestamp: metadata.Timestamp})
	}
	return options, nil
}

func promptUserForOption(options []*SchedulingMetadataOption) *SchedulingMetadataOption {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Available options:")
	for _, option := range options {
		fmt.Println(option.String())
	}

	fmt.Print("Enter the option number: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 0 || choice >= len(options) {
		fmt.Printf("Invalid input \"%s\". Please enter a number between 0 and %d.\n", input, len(options)-1)
		return promptUserForOption(options)
	}
	return options[choice]
}

func reconstructFromOption(reconstructTime time.Time) (*orb.SchedulingInput, error) {
	baselineFilename, differencesFilenames := GetReconstructionFiles(reconstructTime)
	fmt.Println("Finding baseline file: ", baselineFilename)
	reconstructedBaseline, err := ReconstructSchedulingInput(baselineFilename)
	if err != nil {
		fmt.Println("Error executing option:", err)
		return nil, err
	}

	reconstructedDifferences, err := ReconstructDifferences(differencesFilenames)
	if err != nil {
		fmt.Println("Error reconstructing scheduling input differences", err)
		return nil, err
	}

	return orb.MergeDifferences(reconstructedBaseline, reconstructedDifferences, reconstructTime), nil
}

// Function to pull from a directory (either a PV or a local log folder)
func ReadLog(logname string) ([]byte, error) {
	path := filepath.Join(logPath, sanitizePath(logname))
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil, err
	}
	defer file.Close()

	contents, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file bytes:", err)
		return nil, err
	}
	return contents, nil
}

// This function tests whether we can read from the PV and reconstruct the data

// Function to get the files which will reconstruct a scheduling input based on a target "reconstructTime"
// This includes getting the the most recent baseline and the list of differences. It will return that as a tuple (string, []string)
func GetReconstructionFiles(reconstructTime time.Time) (string, []string) {
	baselineName, baselineTime := GetMostRecentBaseline(reconstructTime)
	differences := GetDifferencesFromBaseline(reconstructTime, baselineTime)
	return baselineName, differences
}

// Function to get the differences between the scheduling input and the baseline

// Based on a passed in time (time.Time), get the most recent Baseline filename from a list of filenames
func GetMostRecentBaseline(reconstructTime time.Time) (string, time.Time) {
	// Get all files in the directory
	files, err := os.ReadDir(logPath)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return "", time.Time{}
	}

	// Filter out files that don't start with "SchedulingInputBaseline"
	var baselineFiles []string
	regex := regexp.MustCompile(`^SchedulingInputBaseline.*\.log$`)
	for _, file := range files {
		if regex.MatchString(file.Name()) {
			baselineFiles = append(baselineFiles, file.Name())
		}
	}
	// If no baseline files found, return an empty string
	if len(baselineFiles) == 0 {
		return "", time.Time{}
	}

	// For baseline files that do start with that, after Baseline there should be _ then a timestamp formated as "2006-01-02 15-04-05"
	// Parse each of those timestamps from all these files and make a slice of timestamps
	timestamps := []time.Time{}
	regex = regexp.MustCompile(`_([0-9]{4}-[0-9]{2}-[0-9]{2}_[0-9]{2}-[0-9]{2}-[0-9]{2})\.log$`)
	// Iterate through the baselineFiles and extract the timestamp from each filename
	for _, baselineFile := range baselineFiles {
		match := regex.FindStringSubmatch(baselineFile)
		if len(match) > 1 {
			timestamp, err := time.Parse("2006-01-02_15-04-05", match[1])
			if err == nil {
				timestamps = append(timestamps, timestamp)
			}
		}
	}

	// sort those timestamps to be oldest to newest
	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i].Before(timestamps[j]) })
	mostRecentTimestamp := time.Time{}

	// Find the baseline immediately preceding the timestamp we're looking for.
	// This is the most recent baseline file
	for _, timestamp := range timestamps {
		if timestamp.Before(reconstructTime) || timestamp.Equal(reconstructTime) {
			mostRecentTimestamp = timestamp
		}
	}

	// Convert the most recent timestamp back to a filename
	mostRecentBaselineFilename := fmt.Sprintf("SchedulingInputBaseline_%s.log", mostRecentTimestamp.Format("2006-01-02_15-04-05"))

	// Return the most recent baseline filename
	return mostRecentBaselineFilename, mostRecentTimestamp
}

// Similar to the above get most recent baseline, this function will get the most recent scheduling input differences.
// The differences here those is that this is a slice of differences up to and including the SchedulingInputDifference_... file that contains the
// reconstructTime timestamp within it's changes (it could be the first, the last or somewhere in the middle); and that it returns the slice of strings
// of all those filenames.
func GetDifferencesFromBaseline(reconstructTime time.Time, baselineTime time.Time) []string {
	// Get all files in the directory
	files, err := os.ReadDir(logPath)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return nil
	}

	// Filter out files that don't start with "SchedulingInputDifferences"
	var differenceFiles []string
	regex := regexp.MustCompile(`^SchedulingInputDifferences.*\.log$`)
	for _, file := range files {
		if regex.MatchString(file.Name()) {
			differenceFiles = append(differenceFiles, file.Name())
		}
	}
	// If no difference files found, return an empty string
	if len(differenceFiles) == 0 {
		return nil
	}

	// Map each difference file to a pair of start and end times
	fileTimesMap := map[string][]time.Time{}
	regex = regexp.MustCompile(`_([0-9]{4}-[0-9]{2}-[0-9]{2}_[0-9]{2}-[0-9]{2}-[0-9]{2})_([0-9]{4}-[0-9]{2}-[0-9]{2}_[0-9]{2}-[0-9]{2}-[0-9]{2})\.log$`)
	for _, differenceFile := range differenceFiles {
		match := regex.FindStringSubmatch(differenceFile)
		if len(match) > 2 {
			startTime, err := time.Parse("2006-01-02_15-04-05", match[1])
			if err == nil {
				endTime, err := time.Parse("2006-01-02_15-04-05", match[2])
				if err == nil {
					fileTimesMap[differenceFile] = []time.Time{startTime, endTime}
				}
			}
		}
	}

	differenceFilesFromBaseline := []string{}
	for filename, times := range fileTimesMap {
		startTime, _ := times[0], times[1]
		if (startTime.Equal(baselineTime) || startTime.After(baselineTime)) &&
			(reconstructTime.Equal(startTime) || reconstructTime.After(startTime)) {
			differenceFilesFromBaseline = append(differenceFilesFromBaseline, filename)
		}
	}
	return differenceFilesFromBaseline
}

func writeReconstructionToJSON(schedulingInput *orb.SchedulingInput) error {
	reconstructedFilename := fmt.Sprintf("ReconstructedSchedulingInput_%s.json", schedulingInput.Timestamp.Format("2006-01-02_15-04-05"))
	path := filepath.Join(logPath, reconstructedFilename)
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	_, err = file.WriteString(schedulingInput.Json())
	if err != nil {
		fmt.Println("Error writing reconstruction to file:", err)
		return err
	}
	fmt.Println("Reconstruction written to file successfully!")
	return nil
}

// Reconstructs the scheduling input from a log file in PV or local folder
func ReconstructSchedulingInput(fileName string) (*orb.SchedulingInput, error) {
	readdata, err := ReadLog(fileName)
	if err != nil {
		fmt.Println("Error reading from PV:", err)
		return nil, err
	}
	si, err := orb.UnmarshalSchedulingInput(readdata)
	if err != nil {
		fmt.Println("Error converting PB to SI:", err)
		return nil, err
	}
	return si, nil
}

// Function for reconstructing inputs' differences
func ReconstructDifferences(fileNames []string) ([]*orb.SchedulingInputDifferences, error) {
	allDifferences := []*orb.SchedulingInputDifferences{}
	for _, filename := range fileNames {
		readdata, err := ReadLog(filename)
		if err != nil {
			fmt.Println("Error reading from PV:", err)
			return nil, err
		}
		differences, err := orb.UnmarshalBatchedDifferences(readdata)
		if err != nil {
			fmt.Println("Error converting PB to SI:", err)
			return nil, err
		}
		allDifferences = append(allDifferences, differences...)
	}
	return allDifferences, nil
}

// Security Issue Common Weakness Enumeration (CWE)-22,23 Path Traversal
// They highly recommend sanitizing inputs before accessing that path.
func sanitizePath(path string) string {
	// Remove any leading or trailing slashes, "../" or "./"...
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	path = regexp.MustCompile(`\.\.\/`).ReplaceAllString(path, "")
	path = regexp.MustCompile(`\.\/`).ReplaceAllString(path, "")
	path = strings.ReplaceAll(path, "../", "")

	return path
}

package main

import (
	"bufio"
	"encoding/json"
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

	"sigs.k8s.io/karpenter/pkg/controllers/orb"
	pb "sigs.k8s.io/karpenter/pkg/controllers/orb/proto"
)

type SchedulingMetadataOption struct {
	ID        int
	Action    string
	Timestamp time.Time
}

func (o *SchedulingMetadataOption) String() string {
	return fmt.Sprintf("%d. %s (%s)", o.ID, o.Action, o.Timestamp.Format("2006-01-02_15-04-05"))
}

func main() {
	dirPath := parseCommandLineArgs()
	options := readMetadataLogs(dirPath)
	selectedOption := promptUserForOption(options)

	fmt.Printf("Pulling this option: %s from this directory: %s\n", selectedOption, dirPath)
	resimulateFromOption(selectedOption.Timestamp, dirPath)
}

func parseCommandLineArgs() string {
	dirPath := flag.String("dir", "", "Path to the directory containing logs")
	flag.Parse()
	return *dirPath
}

func readMetadataLogs(dirPath string) []*SchedulingMetadataOption {
	// Read metadata log file and extract available options

	// Using OS. functions, get all filenames starting with "SchedulingMetadata" within the dirPath provided
	files, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return nil
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
	allMetadata := *orb.NewSchedulingMetadataHeap()
	for _, metadataFile := range metadataFiles {
		contents, err := ReadFromSampleFolder(dirPath, metadataFile)
		if err != nil {
			fmt.Println("Error reading file contents:", err)
			return nil
		}

		// Unmarshal the metadataLogdata back into the metadatamap
		protoMetadataMap := &pb.SchedulingMetadataMap{}
		proto.Unmarshal(contents, protoMetadataMap)
		metadataHeap := orb.ReconstructSchedulingMetadataHeap(protoMetadataMap)

		// While metadataHeap isn't empty, pop off each metadata, and save the Action and timestamp to a string option line with a newline at the end
		// Then, append that option to the options slice
		for metadataHeap.Len() > 0 {
			// Recombine/order all metadata from multiple files
			allMetadata.Push(metadataHeap.Pop().(orb.SchedulingMetadata))
		}
	}

	for i, metadata := range allMetadata {
		options = append(options, &SchedulingMetadataOption{ID: i, Action: metadata.Action, Timestamp: metadata.Timestamp})
	}
	return options
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
	if err != nil || choice < 0 || choice > len(options) {
		fmt.Printf("Invalid input \"%s\". Please enter a number between 0 and %d.\n", input, len(options))
		// fmt.Printf("You selected %s, but for testing I'm running just the first one\n", input)
		return promptUserForOption(options)
	}
	fmt.Printf("You selected %s\n", options[choice].String())
	return options[choice]
}

func resimulateFromOption(timestamp time.Time, dirPath string) {
	fileName := GetMostRecentBaseline(timestamp, dirPath)
	fmt.Println("Finding this file: ", fileName)
	_, err := ReadSampleandReconstruct(dirPath, fileName, timestamp)
	if err != nil {
		fmt.Println("Error executing option:", err)
	}

	// reconstructedDifferences := []*orb.SchedulingInputDifferences{}

	// // Resimulate from this scheduling input
	// ctx := signals.NewContext()

	// p := scheduler.NewScheduler()
	// results := scheduler.Solve(ctx, pods)

	// fmt.Println("Resimulating from this scheduling input:", reconstructedSchedulingInput) // Delete
}

// Function to pull from a sample folder
// TODO Just for demo, change later as necessary
func ReadFromSampleFolder(dirPath string, logname string) ([]byte, error) {
	path := filepath.Join(dirPath, sanitizePath(logname))
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

/* These will be part of the command-line printing representation... */

// Based on a passed in time (time.Time), get the most recent Baseline filename from a list of filenames
func GetMostRecentBaseline(reconstructTime time.Time, dirPath string) string {
	// Get all files in the directory
	files, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return ""
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
		return ""
	}

	fmt.Println("Looking for this timestamp: ", reconstructTime.Format("2006-01-02_15-04-05"))

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
	return mostRecentBaselineFilename
}

// // Similar to the above get most recent baseline, this function will get the most recent scheduling input differences.
// // The differences here those is that this is a slice of differences up to and including the SchedulingInputDifference_... file that contains the
// // reconstructTime timestamp within it's changes (it could be the first, the last or somewhere in the middle); and that it returns the slice of strings
// // of all those filenames.
// func GetDifferencesHistoryUpTilTime(reconstructTime time.Time, dirPath string) []string {
// 	// Get all files in the directory
// 	files, err := os.ReadDir(dirPath)
// 	if err != nil {
// 		fmt.Println("Error reading directory:", err)
// 		return nil
// 	}

// 	// Filter out files that don't start with "SchedulingInputDifference"
// 	var differenceFiles []string
// 	regex := regexp.MustCompile(`^SchedulingInputDifference.*\.log$`)
// 	for _, file := range files {
// 		if regex.MatchString(file.Name()) {
// 			differenceFiles = append(differenceFiles, file.Name())
// 		}
// 	}
// 	// If no difference files found, return an empty string
// 	if len(differenceFiles) == 0 {
// 		return nil
// 	}

// 	// For difference files that do start with that, after SchedulingInputDifference there should be _ then a timestamp formated as "2006-01-02 15-04-05"
// 	// Parse each of those timestamps from all these files and make a slice of timestamps
// 	timestamps := []time.Time{}
// 	regex = regexp.MustCompile(`_([0-9]{4}-[0-9]{2}-[0-9]{2}_[0-9]{2}-[0-9]{2}-[0-9]{2})\.log$`)
// 	// Iterate through the differenceFiles and extract the timestamp from each filename
// 	// This is the same as the above, but we're just getting the timestamps from the difference files
// 	for _, differenceFile := range differenceFiles {
// 		match := regex.FindStringSubmatch(differenceFile)
// 		if len(match) > 1 {
// 			timestamp, err := time.Parse("2006-01-02_15-04-05", match[1])
// 			if err == nil {
// 				timestamps = append(timestamps, timestamp)
// 			}
// 		}
// 	}

// 	// sort those timestamps to be oldest to newest

func ReadSampleandReconstruct(dirPath string, mostRecentBaselineFilename string, timestamp time.Time) (*orb.SchedulingInput, error) {
	si, err := ReconstructSampleSchedulingInput(dirPath, mostRecentBaselineFilename)
	if err != nil {
		fmt.Println("Error reconstructing scheduling input:", err)
		return nil, err
	}

	return si, nil
}

func printReconstructionToFile(dirPath string, timestamp time.Time, schedulingInput *orb.SchedulingInput) error {
	//fmt.Println("Reconstructed Scheduling Input looks like:\n" + si.String())
	reconstructedFilename := fmt.Sprintf("ReconstructedSchedulingInput_%s.log", timestamp.Format("2006-01-02_15-04-05"))
	path := filepath.Join(dirPath, reconstructedFilename)
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	_, err = file.WriteString(schedulingInput.String())
	if err != nil {
		fmt.Println("Error writing reconstruction to file:", err)
		return err
	}

	fmt.Println("Reconstruction written to file successfully!")
	return nil
}

// We're sort of artificially rebuilding the filename here, just to do a loopback test of sorts.
// In reality, we could just pull a file from a known directory, for known filename schemas in certain time ranges
func ReadPVandReconstruct(timestamp time.Time) error {
	timestampStr := timestamp.Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("SchedulingInputBaseline_%s.log", timestampStr)

	si, err := ReconstructSchedulingInput(fileName)
	if err != nil {
		fmt.Println("Error reconstructing scheduling input:", err)
		return err
	}

	fmt.Println("Reconstructed Scheduling Input looks like:\n" + si.String())
	reconstructedFilename := fmt.Sprintf("ReconstructedSchedulingInput_%s.log", timestampStr)
	path := filepath.Join("/data", reconstructedFilename)
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	_, err = file.WriteString(si.String())
	if err != nil {
		fmt.Println("Error writing reconstruction to file:", err)
		return err
	}

	fmt.Println("Reconstruction written to file successfully!")
	return nil
}

// Function for reconstructing inputs
// Read from the sample folder to check  from the Command Line)
func ReconstructSampleSchedulingInput(dirPath string, fileName string) (*orb.SchedulingInput, error) {
	readdata, err := ReadFromSampleFolder(dirPath, fileName)
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

// Function for reconstructing inputs
// Read from the PV to check (will be what the ORB tool does from the Command Line)
func ReconstructSchedulingInput(fileName string) (*orb.SchedulingInput, error) {
	readdata, err := ReadFromPV(fileName)
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

// Function to pull from an S3 bucket
func ReadFromPV(logname string) ([]byte, error) {
	path := filepath.Join("/data", sanitizePath(logname))
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

func DebugWriteSchedulingInputStringToLogFile(item *orb.SchedulingInput, timestampStr string) error {
	fileNameStringtest := fmt.Sprintf("SchedulingInputBaselineTEST_%s.log", timestampStr)
	pathStringtest := filepath.Join("/data", fileNameStringtest)
	file, err := os.Create(pathStringtest)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}
	defer file.Close()

	_, err = file.WriteString(item.String())
	if err != nil {
		fmt.Println("Error writing data to file:", err)
		return err
	}

	return nil
}

func DebugWriteSchedulingInputToJSONFile(item *orb.SchedulingInput, timestampStr string) error {
	fileNameJSONtest := fmt.Sprintf("SchedulingInputBaselineTEST_%s.json", timestampStr)
	pathJSONtest := filepath.Join("/data", fileNameJSONtest)
	file, err := os.Create(pathJSONtest)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}
	defer file.Close()

	jsondata, err := json.Marshal(item)
	if err != nil {
		fmt.Println("Error marshalling data to JSON:", err)
		return err
	}
	_, err = file.Write(jsondata)
	if err != nil {
		fmt.Println("Error writing data to file:", err)
		return err
	}

	return nil
}

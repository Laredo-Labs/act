package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/nektos/act/pkg/model"
)

func printList(plan *model.Plan, outputAsJson bool) error {
	type LineInfoDef struct {
		JobID   string              `json:"job_id"`
		JobName string              `json:"job_name"`
		Stage   string              `json:"stage"`
		WFName  string              `json:"workflow_name"`
		WFFile  string              `json:"workflow_file"`
		Events  string              `json:"events"`
		Matrix  []map[string]string `json:"matrix"`
		RunsOn  []string            `json:"runs_on"`
	}
	lineInfos := []LineInfoDef{}

	header := LineInfoDef{
		JobID:   "Job ID",
		JobName: "Job name",
		Stage:   "Stage",
		WFName:  "Workflow name",
		WFFile:  "Workflow file",
		Events:  "Events",
	}

	jobs := map[string]bool{}
	duplicateJobIDs := false

	jobIDMaxWidth := len(header.JobID)
	jobNameMaxWidth := len(header.JobName)
	stageMaxWidth := len(header.Stage)
	wfNameMaxWidth := len(header.WFName)
	wfFileMaxWidth := len(header.WFFile)
	eventsMaxWidth := len(header.Events)

	for i, stage := range plan.Stages {
		for _, r := range stage.Runs {
			matrixCells := []map[string]string{}
			jobID := r.JobID
			job := r.Workflow.GetJob(jobID)
			matrixes, err := job.GetMatrixes()
			if err != nil {
				return err
			}
			for _, combo := range matrixes {
				cell := map[string]string{}
				for k, v := range combo {
					cell[k] = fmt.Sprintf("%v", v)
				}
				matrixCells = append(matrixCells, cell)
			}

			line := LineInfoDef{
				JobID:   jobID,
				JobName: r.String(),
				Stage:   strconv.Itoa(i),
				WFName:  r.Workflow.Name,
				WFFile:  r.Workflow.File,
				Events:  strings.Join(r.Workflow.On(), `,`),
				Matrix:  matrixCells,
				RunsOn:  job.RunsOn(),
			}
			if _, ok := jobs[jobID]; ok {
				duplicateJobIDs = true
			} else {
				jobs[jobID] = true
			}
			lineInfos = append(lineInfos, line)
			if jobIDMaxWidth < len(line.JobID) {
				jobIDMaxWidth = len(line.JobID)
			}
			if jobNameMaxWidth < len(line.JobName) {
				jobNameMaxWidth = len(line.JobName)
			}
			if stageMaxWidth < len(line.Stage) {
				stageMaxWidth = len(line.Stage)
			}
			if wfNameMaxWidth < len(line.WFName) {
				wfNameMaxWidth = len(line.WFName)
			}
			if wfFileMaxWidth < len(line.WFFile) {
				wfFileMaxWidth = len(line.WFFile)
			}
			if eventsMaxWidth < len(line.Events) {
				eventsMaxWidth = len(line.Events)
			}
		}
	}

	if outputAsJson {
		jsonified, err := json.Marshal(lineInfos)
		if err != nil {
			return err
		}
		fmt.Println(string(jsonified))
	} else {
		jobIDMaxWidth += 2
		jobNameMaxWidth += 2
		stageMaxWidth += 2
		wfNameMaxWidth += 2
		wfFileMaxWidth += 2

		fmt.Printf("%*s%*s%*s%*s%*s%*s\n",
			-stageMaxWidth, header.Stage,
			-jobIDMaxWidth, header.JobID,
			-jobNameMaxWidth, header.JobName,
			-wfNameMaxWidth, header.WFName,
			-wfFileMaxWidth, header.WFFile,
			-eventsMaxWidth, header.Events,
		)
		for _, line := range lineInfos {
			fmt.Printf("%*s%*s%*s%*s%*s%*s\n",
				-stageMaxWidth, line.Stage,
				-jobIDMaxWidth, line.JobID,
				-jobNameMaxWidth, line.JobName,
				-wfNameMaxWidth, line.WFName,
				-wfFileMaxWidth, line.WFFile,
				-eventsMaxWidth, line.Events,
			)
		}
		if duplicateJobIDs {
			fmt.Print("\nDetected multiple jobs with the same job name, use `-W` to specify the path to the specific workflow.\n")
		}
	}
	return nil
}
